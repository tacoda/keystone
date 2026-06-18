package web

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/tacoda/keystone/internal/framework/config"
)

// fsWatcher tails .keystone/ for changes and republishes them onto
// the SSE hub as htmx-flavored events. The dashboard subscribes to
// /events with `sse-swap` directives; an SSE message with an HTML
// fragment swaps the matching DOM node.
//
// The watcher debounces — bursts (a save, a chmod, a sync) collapse
// into one event after debounceWindow. Avoids hammering the
// dashboard during bulk edits.
//
// During each debounce window the watcher records every dirty path
// it sees. At publish time those paths are classified into the
// narrowest SSE topic(s) that apply (see topics.go) so widgets only
// re-fetch when something they care about actually moved. The
// coarse `harness-changed` topic is always emitted, so widgets that
// subscribe to it still tick on every burst.
type fsWatcher struct {
	projectDir     string
	hub            *sseHub
	w              *fsnotify.Watcher
	debounceWindow time.Duration

	// onChange runs after every debounced burst, before the SSE
	// publish. The server wires this to primitiveCache.refresh so the
	// cache is up to date by the time the dashboard re-fetches.
	// Optional — nil is allowed.
	onChange func()

	// onPublish runs after every successful publish with the dirty
	// path set + the SSE topics emitted. The server wires this to
	// the audit log so each burst leaves a JSONL line on disk.
	// Optional — nil is allowed.
	onPublish func(paths []string, topics []sseTopic)

	mu    sync.Mutex
	timer *time.Timer
	dirty map[string]struct{} // path set accumulated across one debounce
}

func newFSWatcher(projectDir string, hub *sseHub, onChange func()) (*fsWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("init fsnotify: %w", err)
	}

	watched := []string{
		filepath.Join(projectDir, ".keystone"),
		filepath.Join(projectDir, config.DefaultHarnessRoot),
	}
	for _, root := range watched {
		if err := addRecursive(w, root); err != nil {
			// Best-effort: a missing subtree is not fatal. The
			// watcher still publishes events from whatever subtrees
			// did register.
			fmt.Fprintf(os.Stderr, "keystone web: skip watch on %s: %v\n", root, err)
		}
	}

	return &fsWatcher{
		projectDir:     projectDir,
		hub:            hub,
		w:              w,
		debounceWindow: 250 * time.Millisecond,
		onChange:       onChange,
		dirty:          map[string]struct{}{},
	}, nil
}

func (fw *fsWatcher) Start(ctx context.Context) {
	go fw.loop(ctx)
}

func (fw *fsWatcher) loop(ctx context.Context) {
	defer fw.w.Close()
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-fw.w.Events:
			if !ok {
				return
			}
			// Ignore irrelevant churn — .git, .swp, files our own
			// write path produces.
			base := filepath.Base(ev.Name)
			if strings.HasPrefix(base, ".") && base != ".keystone" {
				continue
			}
			if strings.HasSuffix(base, "~") || strings.HasSuffix(base, ".swp") {
				continue
			}
			// New directory? Watch it too — recursive coverage.
			if ev.Op&fsnotify.Create == fsnotify.Create {
				if info, err := os.Stat(ev.Name); err == nil && info.IsDir() {
					_ = addRecursive(fw.w, ev.Name)
				}
			}
			fw.record(ev.Name)
			fw.fire()
		case err, ok := <-fw.w.Errors:
			if !ok {
				return
			}
			fmt.Fprintf(os.Stderr, "keystone web: watcher error: %v\n", err)
		}
	}
}

// fire schedules a single debounced publish. Multiple events inside
// debounceWindow collapse into one.
func (fw *fsWatcher) fire() {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	if fw.timer != nil {
		fw.timer.Stop()
	}
	fw.timer = time.AfterFunc(fw.debounceWindow, fw.publish)
}

// record stashes a dirty path observed during the current debounce
// window. The set is drained at publish time. Cheap; bounded by the
// debounce window, not the watcher's lifetime.
func (fw *fsWatcher) record(path string) {
	fw.mu.Lock()
	fw.dirty[path] = struct{}{}
	fw.mu.Unlock()
}

// publish emits one SSE event per topic the debounce burst touched.
// Topics are derived from the dirty path set via topicsForPath();
// the coarse `harness-changed` topic always fires so generic
// subscribers tick on every burst.
//
// We deliberately do NOT compute the diff server-side. The dashboard
// re-fetches from the REST API on signal — keeps the watcher's job
// trivial and the data path single-sourced.
func (fw *fsWatcher) publish() {
	// Drain the dirty set first so the classify-and-publish path
	// doesn't race a fresh burst.
	fw.mu.Lock()
	paths := make([]string, 0, len(fw.dirty))
	for p := range fw.dirty {
		paths = append(paths, p)
	}
	fw.dirty = map[string]struct{}{}
	fw.mu.Unlock()

	// Rebuild any registered caches BEFORE notifying the dashboard.
	// The dashboard re-fetches on the SSE ping; we want the cache
	// warm by the time the request lands so the fetch doesn't race
	// the refresh.
	if fw.onChange != nil {
		fw.onChange()
	}

	// Classify. The first burst on a fresh watcher (no recorded
	// paths — e.g. a direct fire from tests) still emits the coarse
	// topic so anything listening to harness-changed ticks.
	perPath := make([][]sseTopic, 0, len(paths))
	for _, p := range paths {
		perPath = append(perPath, topicsForPath(fw.projectDir, p))
	}
	topics := unionTopics(perPath)
	if len(topics) == 0 {
		topics = []sseTopic{topicHarness}
	}

	now := time.Now().Format(time.RFC3339)
	for _, t := range topics {
		var data string
		switch t {
		case topicHarness:
			// The shell's "live" pill — out-of-band swap so it
			// updates anywhere the layout is mounted.
			data = fmt.Sprintf(`<span id="last-updated" hx-swap-oob="true" class="updated">updated %s</span>`, now)
		default:
			// Narrow topics carry an empty payload. Widgets
			// subscribe via `hx-trigger="sse:<topic>"` and refetch
			// themselves — keeps the watcher dumb and widgets
			// in charge of what "fresh" looks like.
			data = " "
		}
		fw.hub.Publish(sseEvent{Name: string(t), Data: data})
	}

	if fw.onPublish != nil {
		fw.onPublish(paths, topics)
	}
}

// addRecursive walks `root` and registers every directory with the
// fsnotify watcher. fsnotify is non-recursive on every supported
// platform — we add subdirs manually.
func addRecursive(w *fsnotify.Watcher, root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if info.IsDir() {
			return w.Add(path)
		}
		return nil
	})
}
