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
// `deps:` or `traces:`. Empty result is normal (root-level rules,
// orphan corpus); never an error.
//
// Reference forms accepted in deps:
//   <kind>/<id>          (canonical)
//
// Reference forms accepted in traces:
//   corpus/<topic>/<name>     (canonical)
//   <topic>/<name>            (treated as corpus by default)
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
			for _, t := range p.Traces {
				if t == want || t == wantBare ||
					strings.TrimPrefix(t, "corpus/") == strings.TrimPrefix(wantBare, "corpus/") {
					out = append(out, IncomingRef{From: p, Via: "traces"})
					break
				}
			}
		}
	}
	return out
}
