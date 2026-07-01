package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/tacoda/keystone/internal/framework/config"
)

// auditEntry is one line in the per-session JSONL audit log. The
// shape stays small and stable — anything fancy belongs in INDEX
// or the watcher event stream, not here.
type auditEntry struct {
	Timestamp string   `json:"ts"`              // RFC3339, UTC
	Topics    []string `json:"topics"`          // SSE topics this burst emitted
	Paths     []string `json:"paths,omitempty"` // dirty paths inside .charter/
	Summary   string   `json:"summary"`         // one-line human description
}

// auditLog persists charter change events to a per-session
// append-only JSONL file under `.charter/state/audit/`. One file
// per `keystone web serve` process; never overwritten. The
// dashboard reads the tail to render the audit widget; older
// sessions sit on disk until startup prune.
//
// Concurrency: the underlying file handle is goroutine-safe behind
// a single mutex. Watcher publishes are the only writer in
// production; tests may write directly.
type auditLog struct {
	dir    string
	path   string // current session file
	mu     sync.Mutex
	f      *os.File
	closed bool
}

// auditDirFor returns the canonical audit dir under a project. Kept
// as a function so tests can mirror the layout against a temp dir.
func auditDirFor(projectDir string) string {
	return filepath.Join(projectDir, config.DefaultCharterRoot, "state", "audit")
}

// openAuditLog creates the audit dir if missing and opens a fresh
// session file `session-<UTC>-<pid>.jsonl` for append. Uses
// O_CREATE|O_EXCL so we never silently overwrite an existing file —
// the timestamp+pid combo collision is effectively impossible, but
// the charter iron law says never overwrite, period.
func openAuditLog(projectDir string) (*auditLog, error) {
	dir := auditDirFor(projectDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("audit: mkdir: %w", err)
	}

	// Best-effort prune of stale sessions. Failures here are logged
	// to stderr and do not block startup — the new session file
	// matters more than the cleanup.
	if removed, err := pruneAuditSessions(dir, 50, 30*24*time.Hour); err != nil {
		fmt.Fprintf(os.Stderr, "keystone web: audit prune: %v\n", err)
	} else if removed > 0 {
		fmt.Fprintf(os.Stderr, "keystone web: audit pruned %d old session(s)\n", removed)
	}

	name := fmt.Sprintf("session-%s-%d.jsonl",
		time.Now().UTC().Format("20060102T150405Z"), os.Getpid())
	path := filepath.Join(dir, name)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("audit: open %s: %w", path, err)
	}
	return &auditLog{dir: dir, path: path, f: f}, nil
}

// Append serializes one entry and writes it to the session file.
// Safe to call from any goroutine. On write error, the log goes
// best-effort — it's a dashboard convenience, not a system-of-
// record. We do NOT crash the server on a busted audit log.
func (a *auditLog) Append(e auditEntry) {
	if a == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed || a.f == nil {
		return
	}
	data, err := json.Marshal(e)
	if err != nil {
		fmt.Fprintf(os.Stderr, "keystone web: audit marshal: %v\n", err)
		return
	}
	data = append(data, '\n')
	if _, err := a.f.Write(data); err != nil {
		fmt.Fprintf(os.Stderr, "keystone web: audit write: %v\n", err)
	}
}

// Close flushes and closes the session file. Idempotent.
func (a *auditLog) Close() error {
	if a == nil {
		return nil
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed || a.f == nil {
		return nil
	}
	a.closed = true
	return a.f.Close()
}

// Path returns the absolute path to the current session file.
func (a *auditLog) Path() string {
	if a == nil {
		return ""
	}
	return a.path
}

// pruneAuditSessions removes session files older than `maxAge` AND
// beyond the `keep` newest, whichever rule is looser. Returns the
// count removed plus the first error encountered. Designed to be
// run at server start once — not on every event.
func pruneAuditSessions(dir string, keep int, maxAge time.Duration) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	type fileMeta struct {
		name string
		mod  time.Time
	}
	files := make([]fileMeta, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasPrefix(e.Name(), "session-") || !strings.HasSuffix(e.Name(), ".jsonl") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, fileMeta{name: e.Name(), mod: info.ModTime()})
	}
	if len(files) == 0 {
		return 0, nil
	}
	// Newest first.
	sort.Slice(files, func(i, j int) bool { return files[i].mod.After(files[j].mod) })

	cutoff := time.Now().Add(-maxAge)
	removed := 0
	for i, fm := range files {
		// Looser-of-two rule: a file survives if it's among the
		// newest `keep` OR newer than the cutoff.
		if i < keep {
			continue
		}
		if fm.mod.After(cutoff) {
			continue
		}
		if err := os.Remove(filepath.Join(dir, fm.name)); err == nil {
			removed++
		}
	}
	return removed, nil
}

// tailAuditFile reads up to `n` entries from the end of the given
// JSONL file. Used by the audit widget to render the last few
// changes without slurping the whole file.
//
// Implementation note: reads the whole file (sessions are small —
// minutes-to-hours of localhost watcher events, capped by debounce)
// and walks line-by-line. If files ever get large enough that this
// matters, swap for a backward-seek reader.
func tailAuditFile(path string, n int) ([]auditEntry, error) {
	if n <= 0 {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	out := make([]auditEntry, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		var e auditEntry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			continue // skip malformed
		}
		out = append(out, e)
	}
	// Newest first.
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, nil
}

// summarizeAudit builds a short human description for an audit
// entry. Used both for the JSONL `summary` field (so the file is
// readable with `tail -f` alone) and for the widget rendering.
func summarizeAudit(paths []string, topics []string) string {
	switch len(paths) {
	case 0:
		return fmt.Sprintf("watcher fired (topics: %s)", strings.Join(topics, ", "))
	case 1:
		return paths[0]
	default:
		return fmt.Sprintf("%s (+%d more)", paths[0], len(paths)-1)
	}
}

// handleAuditWidget renders the audit log widget. Reads the tail of
// the current session by default; `?session=<file>` switches to a
// prior session listed by listAuditSessions.
func (s *server) handleAuditWidget(w http.ResponseWriter, r *http.Request) {
	dir := auditDirFor(s.projectDir)
	current := ""
	if s.audit != nil {
		current = filepath.Base(s.audit.Path())
	}
	sel := r.URL.Query().Get("session")
	if sel == "" {
		sel = current
	}
	var entries []auditEntry
	if sel != "" {
		// Defensive: never read outside the audit dir even if the
		// query string is hostile.
		safe := filepath.Base(sel)
		if entries2, err := tailAuditFile(filepath.Join(dir, safe), 25); err == nil {
			entries = entries2
		}
	}
	sessions, _ := listAuditSessions(dir)
	s.render(w, "_audit_widget.html", map[string]any{
		"Entries":  entries,
		"Sessions": sessions,
		"Current":  current,
		"Selected": sel,
	})
}

// listAuditSessions returns the audit-dir contents (newest first).
// Used by the session-history selector. Each result is a
// filename-only entry; callers join with `dir` if they need a
// full path.
func listAuditSessions(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	type fileMeta struct {
		name string
		mod  time.Time
	}
	files := make([]fileMeta, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), "session-") || !strings.HasSuffix(e.Name(), ".jsonl") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, fileMeta{name: e.Name(), mod: info.ModTime()})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].mod.After(files[j].mod) })
	out := make([]string, 0, len(files))
	for _, fm := range files {
		out = append(out, fm.name)
	}
	return out, nil
}
