// Package budget estimates context-window consumption for files loaded by
// the agent at session start (ambient) or on demand. The 1.0 estimator is
// a whitespace-approximate count — fast, deterministic, no external deps —
// documented as a heuristic that under-counts compared to real model
// tokenizers like tiktoken. Phase 5's --tokenizer=tiktoken opt-in (PLAN
// open question) lands later.
//
// Measure (load.go) computes what each host loads, always and on demand,
// from the projection plan; caps come from keystone.json `budgets`
// (caps.go).
package budget

import "strings"

// Estimate returns the approximate token count for content. The
// approximation is "count whitespace-separated runs of non-whitespace
// characters" — close enough for relative comparisons between files
// and ports, low enough overhead to scan an entire charter in
// milliseconds.
//
// Empirically this under-counts compared to tiktoken's BPE by ~25–35%
// for English prose. Multiply by ~1.35 if you want a conservative
// "real-tokenizer" estimate.
func Estimate(content []byte) int {
	return len(strings.Fields(string(content)))
}
