package web

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestOpenAuditLog_CreatesFileAndDir confirms openAuditLog creates
// the audit directory + a fresh per-session file with the expected
// name shape and exclusive-create semantics.
func TestOpenAuditLog_CreatesFileAndDir(t *testing.T) {
	root := t.TempDir()
	a, err := openAuditLog(root)
	if err != nil {
		t.Fatalf("openAuditLog: %v", err)
	}
	defer a.Close()

	if _, err := os.Stat(auditDirFor(root)); err != nil {
		t.Fatalf("audit dir not created: %v", err)
	}
	base := filepath.Base(a.Path())
	if !strings.HasPrefix(base, "session-") || !strings.HasSuffix(base, ".jsonl") {
		t.Errorf("session filename shape unexpected: %q", base)
	}
}

// TestAuditLog_AppendWritesJSONL confirms each Append produces a
// single valid JSONL line — readable by tailAuditFile.
func TestAuditLog_AppendWritesJSONL(t *testing.T) {
	root := t.TempDir()
	a, err := openAuditLog(root)
	if err != nil {
		t.Fatalf("openAuditLog: %v", err)
	}
	a.Append(auditEntry{
		Timestamp: "2026-06-18T12:00:00Z",
		Topics:    []string{"harness-changed", "primitives-changed"},
		Paths:     []string{".keystone/harness/guides/process/spec.md"},
		Summary:   ".keystone/harness/guides/process/spec.md",
	})
	a.Append(auditEntry{
		Timestamp: "2026-06-18T12:00:01Z",
		Topics:    []string{"harness-changed"},
		Summary:   "watcher fired (topics: harness-changed)",
	})
	if err := a.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	entries, err := tailAuditFile(a.Path(), 10)
	if err != nil {
		t.Fatalf("tail: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("want 2 entries, got %d", len(entries))
	}
	// Tail returns newest-first.
	if entries[0].Timestamp != "2026-06-18T12:00:01Z" {
		t.Errorf("tail order wrong; got %s first", entries[0].Timestamp)
	}
	if entries[1].Topics[1] != "primitives-changed" {
		t.Errorf("topics not preserved: %v", entries[1].Topics)
	}
}

// TestPruneAuditSessions confirms the "looser of keep / maxAge"
// retention rule: newest `keep` survive even if old; everything
// older than maxAge AND beyond `keep` is removed.
func TestPruneAuditSessions(t *testing.T) {
	dir := t.TempDir()

	// Create 5 files with synthetic mod times.
	mk := func(name string, ago time.Duration) {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
		t0 := time.Now().Add(-ago)
		if err := os.Chtimes(p, t0, t0); err != nil {
			t.Fatal(err)
		}
	}
	mk("session-20260101T000000Z-1.jsonl", 100*24*time.Hour) // very old
	mk("session-20260102T000000Z-2.jsonl", 80*24*time.Hour)  // very old
	mk("session-20260301T000000Z-3.jsonl", 60*24*time.Hour)  // old
	mk("session-20260601T000000Z-4.jsonl", 7*24*time.Hour)   // recent
	mk("session-20260615T000000Z-5.jsonl", 1*time.Hour)      // newest

	// keep=2, maxAge=30 days. Expected survivors: the 2 newest +
	// anything within 30 days. So #4 (7d, within age window) and
	// #5 (1h, within age window). The two newest also pass — same
	// set. #3 is 60d old AND beyond keep → removed. #1,#2 same.
	removed, err := pruneAuditSessions(dir, 2, 30*24*time.Hour)
	if err != nil {
		t.Fatalf("prune: %v", err)
	}
	if removed != 3 {
		t.Errorf("removed=%d want 3", removed)
	}

	survivors, _ := listAuditSessions(dir)
	if len(survivors) != 2 {
		t.Errorf("survivors=%d want 2: %v", len(survivors), survivors)
	}
}

// TestSummarizeAudit covers the three branches a watcher publish can
// land in: empty path set (debounce fired with no recorded path),
// single path (the common case), and multiple paths (batched edit).
func TestSummarizeAudit(t *testing.T) {
	cases := []struct {
		name   string
		paths  []string
		topics []string
		want   string
	}{
		{"empty paths", nil, []string{"harness-changed"}, "watcher fired (topics: harness-changed)"},
		{"single path", []string{".keystone/harness/guides/x.md"}, []string{"harness-changed"}, ".keystone/harness/guides/x.md"},
		{"multi path", []string{".keystone/a", ".keystone/b", ".keystone/c"}, []string{"harness-changed"}, ".keystone/a (+2 more)"},
	}
	for _, c := range cases {
		got := summarizeAudit(c.paths, c.topics)
		if got != c.want {
			t.Errorf("%s: got %q want %q", c.name, got, c.want)
		}
	}
}
