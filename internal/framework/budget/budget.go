// Package budget estimates context-window consumption for files loaded by
// the agent at session start (ambient) or on demand. The 1.0 estimator is
// a whitespace-approximate count — fast, deterministic, no external deps —
// documented as a heuristic that under-counts compared to real model
// tokenizers like tiktoken. Phase 5's --tokenizer=tiktoken opt-in (PLAN
// open question) lands later.
//
// Per-port budgets are declared in keystone.json's `budgets` block (the
// shape already defined as config.BudgetSpec). The Allocator type
// aggregates per-port and per-file consumption; the Report function
// produces a sorted breakdown.
package budget

import (
	"sort"
	"strings"

	"github.com/tacoda/keystone/internal/framework/config"
)

// Estimate returns the approximate token count for content. The
// approximation is "count whitespace-separated runs of non-whitespace
// characters" — close enough for relative comparisons between files
// and ports, low enough overhead to scan an entire harness in
// milliseconds.
//
// Empirically this under-counts compared to tiktoken's BPE by ~25–35%
// for English prose. Multiply by ~1.35 if you want a conservative
// "real-tokenizer" estimate.
func Estimate(content []byte) int {
	return len(strings.Fields(string(content)))
}

// FileEntry is one file's contribution to a port's load.
type FileEntry struct {
	Path   string // relative to project root, slash-separated
	Tokens int
}

// PortReport is the per-port breakdown built by Report.
type PortReport struct {
	Port      string
	Tokens    int          // total across all files for this port
	MaxTokens int          // from keystone.json's budgets block; 0 means no cap
	OverBy    int          // Tokens - MaxTokens; 0 or negative means within budget
	TopFiles  []FileEntry  // top contributors, sorted desc by Tokens
}

// IsOverBudget reports whether this port exceeded its declared cap.
// Returns false when MaxTokens is unset (0).
func (p PortReport) IsOverBudget() bool {
	return p.MaxTokens > 0 && p.Tokens > p.MaxTokens
}

// Allocator aggregates per-port, per-file consumption as the doctor walks
// the harness tree. Add is safe to call from one goroutine; the Allocator
// is not thread-safe.
type Allocator struct {
	files map[string][]FileEntry // port → entries
}

// NewAllocator returns an empty Allocator.
func NewAllocator() *Allocator {
	return &Allocator{files: map[string][]FileEntry{}}
}

// Add records `tokens` against `port`, attributing them to `path` for the
// top-contributors list.
func (a *Allocator) Add(port, path string, tokens int) {
	a.files[port] = append(a.files[port], FileEntry{Path: path, Tokens: tokens})
}

// Report renders the per-port breakdown. Each port's TopFiles is sorted
// descending by Tokens and truncated to topN (use 0 for "all"). Budgets
// from cfg are applied when set; ports with no declared budget have
// MaxTokens == 0.
func (a *Allocator) Report(cfg *config.ProjectConfig, topN int) []PortReport {
	ports := make([]string, 0, len(a.files))
	for p := range a.files {
		ports = append(ports, p)
	}
	sort.Strings(ports)

	out := make([]PortReport, 0, len(ports))
	for _, p := range ports {
		entries := append([]FileEntry(nil), a.files[p]...)
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].Tokens != entries[j].Tokens {
				return entries[i].Tokens > entries[j].Tokens
			}
			return entries[i].Path < entries[j].Path
		})

		total := 0
		for _, e := range entries {
			total += e.Tokens
		}

		max := 0
		if cfg != nil {
			if spec, ok := cfg.Budgets[p]; ok {
				max = spec.MaxTokens
			}
		}

		top := entries
		if topN > 0 && len(entries) > topN {
			top = entries[:topN]
		}

		overBy := 0
		if max > 0 && total > max {
			overBy = total - max
		}

		out = append(out, PortReport{
			Port:      p,
			Tokens:    total,
			MaxTokens: max,
			OverBy:    overBy,
			TopFiles:  top,
		})
	}
	return out
}

// PortForPath returns the port name a relative harness path belongs to,
// or "" when the path does not map to any port (e.g. README.md at the
// harness root, learning/ state, archive/ state). The caller can use ""
// as a signal to skip the file.
//
// Examples (with harnessRoot="harness"):
//
//	"harness/guides/process/spec.md" → "guides"
//	"harness/corpus/principles/tdd.md" → "corpus"
//	"harness/sensors/build.md" → "sensors"
//	"harness/adapters/claude-code/lifecycle.md" → "adapters"
//	"harness/learning/inbox/X.md" → ""
//	"harness/README.md" → ""
func PortForPath(relPath, harnessRoot string) string {
	prefix := harnessRoot + "/"
	if !strings.HasPrefix(relPath, prefix) {
		return ""
	}
	rest := strings.TrimPrefix(relPath, prefix)
	if i := strings.IndexByte(rest, '/'); i >= 0 {
		port := rest[:i]
		switch port {
		case "guides", "corpus", "sensors", "actions", "playbooks", "adapters":
			return port
		}
	}
	return ""
}
