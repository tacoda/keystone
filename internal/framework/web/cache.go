package web

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/mcp"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// healthRefreshInterval is how often the background ticker re-probes
// every configured source's health. 30s strikes a balance: stale
// enough that handlers never block, fresh enough that the dashboard
// catches a flapping source within a single user attention span.
const healthRefreshInterval = 30 * time.Second

// healthProbeTimeout caps a single source's probe. Independent of any
// per-adapter timeout — if an adapter ignores ctx, this still bounds
// the refresh wall-clock.
const healthProbeTimeout = 8 * time.Second

// healthProbeConcurrency bounds the worker pool that fans out probes.
// Most installs have ≤4 sources; the cap mainly protects against a
// pathological context.json with dozens of entries.
const healthProbeConcurrency = 8

// primitiveCache holds the most recent walk of the harness — primitive
// list + derived index. Page handlers read from here instead of
// re-walking the filesystem on every request. The watcher rebuilds
// the cache on a debounced file-change event.
type primitiveCache struct {
	projectDir string

	mu         sync.RWMutex
	primitives []primitive.Primitive
	index      primitive.Index
	lastError  error
}

func newPrimitiveCache(projectDir string) *primitiveCache {
	return &primitiveCache{projectDir: projectDir}
}

// refresh walks the harness and replaces the cached snapshot. Safe
// to call from any goroutine. Errors are stored on the cache so
// callers can surface them without re-walking.
func (c *primitiveCache) refresh() {
	prims, _, err := primitive.Walk(c.projectDir, config.DefaultHarnessRoot)
	if err != nil {
		c.mu.Lock()
		c.lastError = err
		c.mu.Unlock()
		return
	}
	idx := primitive.Build(prims, time.Now())
	c.mu.Lock()
	c.primitives = prims
	c.index = idx
	c.lastError = nil
	c.mu.Unlock()
}

// get returns a snapshot of the cached primitives + index. The
// returned slice/maps are NOT defensively copied — callers must
// treat them as read-only.
func (c *primitiveCache) get() ([]primitive.Primitive, primitive.Index, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.primitives, c.index, c.lastError
}

// healthCache holds the most recent source-list snapshot with each
// source's probed health. Page handlers read from here instead of
// triggering a synchronous fan-out of HTTP probes on every render.
// A background ticker keeps the snapshot fresh.
type healthCache struct {
	projectDir string

	mu        sync.RWMutex
	entries   []sourceEntry
	lastError error
}

func newHealthCache(projectDir string) *healthCache {
	return &healthCache{projectDir: projectDir}
}

// refresh re-reads context.json and probes every configured source's
// health in parallel under a bounded worker pool. One slow source
// can no longer hold up the whole refresh — the per-probe timeout
// fires independently per goroutine.
func (c *healthCache) refresh(ctx context.Context) {
	cfg, err := mcp.LoadContextConfig(c.projectDir)
	if err != nil {
		c.mu.Lock()
		c.lastError = err
		c.mu.Unlock()
		return
	}
	if cfg == nil {
		c.mu.Lock()
		c.entries = []sourceEntry{}
		c.lastError = nil
		c.mu.Unlock()
		return
	}

	results := make([]sourceEntry, len(cfg.Sources))
	sem := make(chan struct{}, healthProbeConcurrency)
	var wg sync.WaitGroup
	for i, src := range cfg.Sources {
		i, src := i, src
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			a := mcp.BuildAdapter(src)
			probeCtx, cancel := context.WithTimeout(ctx, healthProbeTimeout)
			defer cancel()
			h := a.Health(probeCtx)
			results[i] = sourceEntry{
				Name:   a.Name(),
				Type:   a.Type(),
				Health: h,
			}
		}()
	}
	wg.Wait()

	c.mu.Lock()
	c.entries = results
	c.lastError = nil
	c.mu.Unlock()
}

// get returns a snapshot of the cached source entries. Caller treats
// the slice as read-only.
func (c *healthCache) get() ([]sourceEntry, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.entries, c.lastError
}

// startCacheRefreshers kicks off the background refresh loops:
//
//   - primitiveCache rebuilds on every fsWatcher publish (already
//     debounced) plus a slow ticker fallback.
//   - healthCache rebuilds on a 30s ticker. Source-mutation handlers
//     also call s.healthCache.refresh directly for immediate feedback.
//
// Both caches do one synchronous fill before this returns so the
// first request never sees an empty snapshot.
func (s *server) startCacheRefreshers(ctx context.Context) {
	s.primitiveCache.refresh()
	s.healthCache.refresh(ctx)

	go func() {
		t := time.NewTicker(healthRefreshInterval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				s.healthCache.refresh(ctx)
			}
		}
	}()

	// Slow safety-net refresh for the primitive cache in case
	// fsnotify drops events (rare but possible under heavy churn or
	// on network filesystems).
	go func() {
		t := time.NewTicker(2 * time.Minute)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				s.primitiveCache.refresh()
			}
		}
	}()

	// Optional: log refresher startup once. Useful when diagnosing a
	// dashboard that looks stale — confirms the loop is running.
	fmt.Fprintf(os.Stderr, "keystone web: caches warm (primitive + health)\n")
}
