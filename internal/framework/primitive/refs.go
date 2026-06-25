package primitive

import "strings"

// Reverse references — given a set of primitives, return who points
// at whom. Used by the dashboard's primitive-detail "referenced by"
// panel + by the prune heuristic.

// IncomingRef captures one primitive pointing at the target.
type IncomingRef struct {
	From Primitive `json:"from"`
	Via  string    `json:"via"` // "deps" | "traces"
}

// IncomingRefs returns every primitive that references target via
// `deps:`, `traces:`, or `includes:`. Empty result is normal
// (root-level rules, orphan corpus, concerns no one composes); never
// an error.
//
// Reference forms accepted in deps:
//   <kind>/<id>          (canonical)
//
// Reference forms accepted in traces:
//   corpus/<topic>/<name>     (canonical)
//   <topic>/<name>            (treated as corpus by default)
//
// Reference forms accepted in includes (only meaningful when target
// is a concern):
//   <id>                 (bare concern id; canonical)
func IncomingRefs(primitives []Primitive, target Primitive) []IncomingRef {
	want := target.Kind + "/" + target.ID
	wantBare := target.ID
	var out []IncomingRef
	for _, p := range primitives {
		for _, d := range p.Deps {
			if d == want {
				out = append(out, IncomingRef{From: p, Via: "deps"})
				break
			}
		}
		// Traces are typically corpus references — match against
		// `corpus/<id>` AND bare `<id>`.
		if target.Kind == string(KindCorpus) {
			for _, t := range p.Corpus {
				if t == want || t == wantBare ||
					strings.TrimPrefix(t, "corpus/") == strings.TrimPrefix(wantBare, "corpus/") {
					out = append(out, IncomingRef{From: p, Via: "traces"})
					break
				}
			}
		}
		// Includes are bare concern ids — only the target's id is
		// meaningful here. Composition is depth-1, so a single pass
		// over `includes` covers the entire reverse-lookup.
		if target.Kind == string(KindConcern) {
			for _, inc := range p.Includes {
				if inc == wantBare {
					out = append(out, IncomingRef{From: p, Via: "includes"})
					break
				}
			}
		}
	}
	return out
}
