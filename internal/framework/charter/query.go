package charter

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

// Query composes "what context applies here?" — the front door both the
// CLI (`keystone charter query`) and the MCP server
// (`keystone_charter_query`) share, so the two never disagree.
//
// Input is either a set of repository files (a region) or a free-text
// task, or both. Composition:
//
//  1. Everywhere primitives — rules without `paths:` always apply; they
//     cost one descriptor line each. Hooks, skills, commands, and agents
//     are invocable, not ambient: they enter through their event,
//     description match (task), or paths — never by flooding every query.
//  3. Path-scoped primitives — any primitive whose `paths:` matches a
//     queried file. MatchPath (doublestar) decides.
//  4. Cited references — the `corpus:` entries of matched rules and
//     corpus, resolved by id.
//  2. Task match — when a task is given, token-overlap-scored entries
//     lead; they are the most specific reason anything matched.
//
// The result carries descriptors only; callers print bodies on demand
// (`--full`), whose paths come back with the entry.
type QueryResult struct {
	// Entries is the composed context, everywhere primitives first (in
	// kind order), then path-scoped (kind order), then task-scored
	// (score descending). Provenance travels with each entry.
	Entries []QueryEntry
}

// QueryEntry is one composed primitive: the descriptor plus how it got in.
type QueryEntry struct {
	Kind        string
	ID          string
	Description string
	Path        string
	Provenance  string
	// Why records the reason this entry matched: "everywhere",
	// "paths: <glob> matched <file>", "cited by <kind>/<id>", or
	// "task: <score>".
	Why string
}

// Query runs the composition over composed primitives. files may be
// empty; task may be empty; at least one should be set.
func Query(prims []primitive.Primitive, files []string, task string) QueryResult {
	var res QueryResult
	seen := map[string]bool{}  // kind/id → already in the result
	cited := map[string]bool{} // reference ids already queued
	citeQueue := []string{}

	add := func(p primitive.Primitive, why string, _ int) {
		key := p.Kind + "/" + p.ID
		if seen[key] {
			return
		}
		seen[key] = true
		res.Entries = append(res.Entries, QueryEntry{
			Kind: p.Kind, ID: p.ID, Description: p.Description,
			Path: p.Path, Provenance: p.Provenance, Why: why,
		})
		for _, ref := range p.Corpus {
			if !cited[ref] {
				cited[ref] = true
				citeQueue = append(citeQueue, ref)
			}
		}
	}

	// 1. Task-scored entries lead when a task is given — they are the
	// most specific reason anything matched; everywhere rules follow
	// (they always apply, but they are the least specific).
	if task != "" {
		type scored struct {
			p     primitive.Primitive
			score int
		}
		var matches []scored
		for _, p := range prims {
			if s := taskScore(p, task); s > 0 {
				matches = append(matches, scored{p, s})
			}
		}
		sort.SliceStable(matches, func(i, j int) bool { return matches[i].score > matches[j].score })
		for _, m := range matches {
			add(m.p, "task", m.score)
		}
	}

	// 2. Everywhere primitives — rules without paths: only.
	everywhere := make([]primitive.Primitive, 0, len(prims))
	for _, p := range prims {
		if p.Kind == string(primitive.KindRule) && len(p.Paths) == 0 {
			everywhere = append(everywhere, p)
		}
	}
	primitive.SortByKind(everywhere)
	for _, p := range everywhere {
		add(p, "everywhere", 0)
	}

	// 2. Path-scoped primitives against each queried file.
	for _, p := range prims {
		for _, glob := range p.Paths {
			hit := ""
			for _, f := range files {
				if primitive.MatchPath(glob, filepath.ToSlash(f)) {
					hit = f
					break
				}
			}
			if hit != "" {
				add(p, "paths: "+glob+" matched "+hit, 0)
				break
			}
		}
	}

	// 3. References cited by anything already added. A path-matched rule
	// brings its citations in even when the everywhere set skipped other
	// no-path knowledge.
	for _, ref := range citeQueue {
		for _, p := range prims {
			if p.Kind == string(primitive.KindReference) && p.ID == ref {
				add(p, "cited by a matched primitive", 0)
			}
		}
	}

	return res
}

// taskScore scores one primitive against the task text: token overlap
// over id, description, and tags. Deliberately naive — a keyword signal,
// not a search engine (semantic search stays out of scope).
// stopwords never score — they match everything and bury real hits.
var stopwords = map[string]bool{
	"the": true, "a": true, "an": true, "and": true, "or": true,
	"to": true, "of": true, "in": true, "for": true, "is": true,
	"are": true, "how": true, "what": true, "when": true,
}

func taskScore(p primitive.Primitive, task string) int {
	hay := strings.ToLower(p.ID + " " + p.Description + " " + strings.Join(p.Tags, " "))
	hayTokens := tokenSet(hay)
	score := 0
	for _, tok := range strings.Fields(strings.ToLower(task)) {
		tok = strings.Trim(tok, ".,:;()[]\"'")
		if len(tok) < 3 || stopwords[tok] {
			continue
		}
		if hayTokens[tok] {
			score++
		}
	}
	return score
}

func tokenSet(s string) map[string]bool {
	out := map[string]bool{}
	for _, f := range strings.Fields(s) {
		f = strings.Trim(f, ".,:;()[]\"'-_/")
		if f != "" {
			out[f] = true
			// id segments split on separators: process/verification and
			// go/style both become searchable words.
			for _, seg := range strings.Split(f, "/") {
				if seg != "" {
					out[seg] = true
				}
			}
		}
	}
	return out
}
