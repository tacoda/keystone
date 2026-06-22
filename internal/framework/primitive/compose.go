package primitive

import (
	"fmt"
	"sort"
)

// Compose resolves every primitive's `includes:` list against the
// concern set in the same primitive slice, returning a new slice in
// which each host primitive's list fields have been union-merged with
// every concern it includes.
//
// Composition contract:
//
//   - Concerns are leaves. A primitive with `kind: concern` MUST NOT
//     declare its own `includes:` — that's reported as a ComposeError
//     and the primitive is returned with composition skipped (so the
//     indexer can still surface the structural issue downstream).
//   - List fields union (deduped, stable-ordered): Globs, Triggers,
//     Tools, Tags, Traces, Deps, HostTriggers. The host primitive's
//     own values come LAST in the merged list so a re-listed value
//     can demote a concern-contributed one.
//   - Scalar fields stay HOST-WINS: Kind, ID, Description, Severity,
//     Model, Phase, Tier, Args. A concern can never override what
//     the host primitive declares about itself.
//   - The raw `Includes` slice on the host is preserved untouched —
//     callers that want to render "this primitive includes X, Y, Z"
//     read the original list; the merged effects are observed via
//     the union'd list fields.
//
// Compose does NOT mutate concern primitives. They keep their own
// declared frontmatter and are still indexed as `kind: concern`
// entries the agent can read directly. The composition only affects
// the HOST primitives that include them.
//
// Cycle freedom is structural: since concerns can't include other
// concerns, the depth is always one (host → concern). No graph
// traversal required.
func Compose(primitives []Primitive) ([]Primitive, []ComposeError) {
	concerns, errs := indexConcerns(primitives)

	out := make([]Primitive, len(primitives))
	for i, p := range primitives {
		if Kind(p.Kind) == KindConcern {
			out[i] = p
			continue
		}
		merged, mergeErrs := composeOne(p, concerns)
		out[i] = merged
		errs = append(errs, mergeErrs...)
	}
	return out, errs
}

// ComposeError describes one violation of the composition contract
// (unknown concern id, leaf-violation, duplicate include). The
// indexer surfaces these alongside parse warnings; `keystone lint`
// upgrades them to errors.
type ComposeError struct {
	Path    string
	ID      string
	Message string
}

func (e ComposeError) Error() string {
	return fmt.Sprintf("%s [%s]: %s", e.Path, e.ID, e.Message)
}

// indexConcerns builds {id → Primitive} for every kind=concern entry
// and reports any concern that itself declares `includes:` (leaf
// violation). Returned errors are stable-ordered for deterministic
// test output.
func indexConcerns(primitives []Primitive) (map[string]Primitive, []ComposeError) {
	concerns := map[string]Primitive{}
	var errs []ComposeError
	for _, p := range primitives {
		if Kind(p.Kind) != KindConcern {
			continue
		}
		if len(p.Includes) > 0 {
			errs = append(errs, ComposeError{
				Path: p.Path, ID: p.ID,
				Message: "concerns are leaves — `includes:` is not allowed on kind=concern",
			})
			continue
		}
		concerns[p.ID] = p
	}
	sort.SliceStable(errs, func(i, j int) bool {
		if errs[i].Path != errs[j].Path {
			return errs[i].Path < errs[j].Path
		}
		return errs[i].ID < errs[j].ID
	})
	return concerns, errs
}

// composeOne returns p with concern contributions merged in. The
// host primitive's own values appear LAST in each list so a
// re-listed value demotes the concern's version (handy when an
// included concern adds a tool the host wants to override or
// drop downstream).
func composeOne(host Primitive, concerns map[string]Primitive) (Primitive, []ComposeError) {
	if len(host.Includes) == 0 {
		return host, nil
	}
	merged := host
	seen := map[string]struct{}{}
	var errs []ComposeError

	// Resolved-then-host order means concern values come FIRST and
	// the host's own contribute LAST. We accumulate from concerns
	// in `includes:` array order, then append the host's own
	// declared lists at the end.
	var (
		globs        []string
		triggers     []string
		tools        []string
		tags         []string
		traces       []string
		deps         []string
		hostTriggers []HostTrigger
	)

	for _, id := range host.Includes {
		if _, dup := seen[id]; dup {
			errs = append(errs, ComposeError{
				Path: host.Path, ID: host.ID,
				Message: fmt.Sprintf("duplicate concern in includes: %q", id),
			})
			continue
		}
		seen[id] = struct{}{}
		c, ok := concerns[id]
		if !ok {
			errs = append(errs, ComposeError{
				Path: host.Path, ID: host.ID,
				Message: fmt.Sprintf("unknown concern id %q", id),
			})
			continue
		}
		globs = append(globs, c.Globs...)
		triggers = append(triggers, c.Triggers...)
		tools = append(tools, c.Tools...)
		tags = append(tags, c.Tags...)
		traces = append(traces, c.Traces...)
		deps = append(deps, c.Deps...)
		hostTriggers = append(hostTriggers, c.HostTriggers...)
	}

	// Append host's own (declared) values last; dedupe each list
	// preserving the FIRST occurrence so concern values stay ahead
	// when the host hasn't re-listed them.
	merged.Globs = dedupeStable(append(globs, host.Globs...))
	merged.Triggers = dedupeStable(append(triggers, host.Triggers...))
	merged.Tools = dedupeStable(append(tools, host.Tools...))
	merged.Tags = dedupeStable(append(tags, host.Tags...))
	merged.Traces = dedupeStable(append(traces, host.Traces...))
	merged.Deps = dedupeStable(append(deps, host.Deps...))
	merged.HostTriggers = dedupeHostTriggers(append(hostTriggers, host.HostTriggers...))

	return merged, errs
}

// dedupeStable returns xs with later duplicates removed; relative
// order is preserved. Empty / nil-safe.
func dedupeStable(xs []string) []string {
	if len(xs) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(xs))
	out := make([]string, 0, len(xs))
	for _, x := range xs {
		if _, dup := seen[x]; dup {
			continue
		}
		seen[x] = struct{}{}
		out = append(out, x)
	}
	return out
}

// dedupeHostTriggers removes triggers whose (phase, matcher, command)
// tuple is identical. Order preserved. A concern contributing a hook
// the host also declares ends up as a single entry.
func dedupeHostTriggers(xs []HostTrigger) []HostTrigger {
	if len(xs) == 0 {
		return nil
	}
	type key struct{ phase, matcher, cmd string }
	seen := make(map[key]struct{}, len(xs))
	out := make([]HostTrigger, 0, len(xs))
	for _, t := range xs {
		k := key{t.Phase, t.Matcher, t.Command}
		if _, dup := seen[k]; dup {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, t)
	}
	return out
}
