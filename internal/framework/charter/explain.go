package charter

import (
	"fmt"
	"strings"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

// Explanation is a human-readable account of one primitive: what it is,
// how/when it activates, what it links to, and where it projects.
type Explanation struct {
	Kind        string
	ID          string
	Description string
	Provenance  string
	Activation  string   // how/when this primitive fires or activates
	Links       []string // "corpus: x", "deps: y", …
	ProjectsTo  string   // host-native path it projects to, or ""
	BodyPath    string
}

// Matches reports whether primitive p has the given id and (when kind is
// non-empty) the given kind — the lookup both the CLI and MCP use.
func Matches(p primitive.Primitive, id, kind string) bool {
	if p.ID != id {
		return false
	}
	return kind == "" || primitive.IsKind(p.Kind, kind)
}

// Explain builds an Explanation for a primitive from its descriptor.
func Explain(p primitive.Primitive) Explanation {
	e := Explanation{
		Kind:        p.Kind,
		ID:          p.ID,
		Description: p.Description,
		Provenance:  p.Provenance,
		BodyPath:    p.Path,
		Links:       explainLinks(p),
		ProjectsTo:  primitive.ProjectionRelPath(p),
	}
	if f := activators[primitive.Kind(p.Kind)]; f != nil {
		e.Activation = f(p)
	} else {
		e.Activation = "activated according to its kind."
	}
	return e
}

// activators renders the kind-specific "how does this fire/activate"
// line. A map (not a switch) keeps Explain's complexity flat.
var activators = map[primitive.Kind]func(primitive.Primitive) string{
	primitive.KindRule: func(p primitive.Primitive) string {
		if len(p.Paths) == 0 {
			return "ambient rule — active every turn in its topic."
		}
		return "ambient rule — active when a touched file matches: " + strings.Join(p.Paths, ", ")
	},
	primitive.KindHook: func(p primitive.Primitive) string {
		on := orAmbient(p.Event)
		if p.Mode == "inferential" {
			return fmt.Sprintf("inferential check — an agent review on %s, returns: %s.", on, orNone(p.Returns))
		}
		return fmt.Sprintf("computational check — runs `%s` on %s; gates on its status.", orNone(p.Run), on)
	},
	primitive.KindTool: func(p primitive.Primitive) string {
		t := orDefault(p.Transport, "cli")
		if strings.TrimSpace(p.Event) != "" {
			return fmt.Sprintf("external callable (transport: %s) — fires as a side-effect on %s.", t, p.Event)
		}
		return fmt.Sprintf("external callable (transport: %s) — invoked on demand.", t)
	},
	primitive.KindSubagent: func(p primitive.Primitive) string {
		return "spawned as a subagent by id; tools: " + orNone(strings.Join(p.Tools, ", "))
	},
	primitive.KindCommand: func(p primitive.Primitive) string {
		return "a unit of work; phase: " + orAny(p.Phase)
	},
	primitive.KindSkill: func(p primitive.Primitive) string {
		return "activates when its description matches the task" + gatesSuffix(p.Gates) + "."
	},
	primitive.KindReference: func(p primitive.Primitive) string {
		return "reasoning — loaded on demand via a rule's corpus: link."
	},
	primitive.KindPermission: func(p primitive.Primitive) string {
		return "allow/ask/deny tool rules — projects to each host's permission config."
	},
	primitive.KindDocument: func(p primitive.Primitive) string {
		return "a governed output template (type: " + orNone(p.Type) + ")."
	},
	primitive.KindEval: func(p primitive.Primitive) string {
		return "a fixture-based regression test for the charter."
	},
	primitive.KindConcern: func(p primitive.Primitive) string {
		return "a composition mixin — pulled into other primitives via includes:."
	},
}

// gatesSuffix names a skill's human gates (carried over from a 4.x playbook).
func gatesSuffix(gates []string) string {
	if len(gates) == 0 {
		return ""
	}
	return "; gates: " + strings.Join(gates, ", ")
}

func explainLinks(p primitive.Primitive) []string {
	var out []string
	if len(p.Corpus) > 0 {
		out = append(out, "corpus: "+strings.Join(p.Corpus, ", "))
	}
	if len(p.Deps) > 0 {
		out = append(out, "deps: "+strings.Join(p.Deps, ", "))
	}
	if len(p.Includes) > 0 {
		out = append(out, "includes: "+strings.Join(p.Includes, ", "))
	}
	return out
}

func orNone(s string) string {
	if strings.TrimSpace(s) == "" {
		return "—"
	}
	return s
}

func orAny(s string) string {
	if strings.TrimSpace(s) == "" {
		return "any"
	}
	return s
}

func orAmbient(s string) string {
	if strings.TrimSpace(s) == "" {
		return "its phase (ambient)"
	}
	return s
}

func orDefault(s, def string) string {
	if strings.TrimSpace(s) == "" {
		return def
	}
	return s
}
