// Package primitive defines the canonical descriptor that every harness
// file declares in its frontmatter, plus the index aggregator that the
// agent reads in lieu of the raw markdown tree.
//
// See docs/ports/primitive.md for the canonical schema. Bodies are not
// stored in this package — only descriptors. The agent opens body files
// (via the Path field) on activation.
package primitive

// Kind names every primitive that can appear in the harness. The set is
// closed; unknown kinds are a lint error.
//
// 2.0 explicitly excludes MCP prompts: prompts only make sense when an
// MCP server and a client UI surface them to the human, and a CLI-only
// harness has no path to invoke them. Slash commands cover human-
// triggered templated workflows; skills cover agent-self-triggered
// patterns. A future MCP-only catalog can live outside this contract.
//
// Two-layer taxonomy. Framework primitives wrap agent primitives the
// way an ORM wraps SQL: the framework primitive is the canonical
// authoring surface, the agent primitive is the raw host-native
// equivalent kept as an escape hatch.
//
//   Framework abstractions (encouraged by default — author here):
//     wrappers       guide, sensor (→ rule)
//                    action        (→ command)
//                    playbook      (→ skill)
//                    persona       (→ subagent)
//     standalone     corpus, eval, source
//
//   Agent abstractions (raw host-native passthrough — escape hatches):
//     rule, skill, subagent, command
//
// Wrap mechanics:
//
//   - persona/action/playbook compile to a host-native file under
//     .claude/ via `keystone project`. The framework primitive owns
//     the target path; collisions with same-id agent primitives are
//     lint errors.
//   - guide/sensor are conceptual wrappers — they share the agent's
//     "rule" semantics but carry richer structure (severity, globs,
//     paired corpus, sensor kind). No file projection; the agent
//     reads them via INDEX.json.
//
// Both layers share the same canonical frontmatter shape and land in
// the same INDEX.json. The distinction is *conceptual* — what each
// primitive's contract is, not how it's stored or served.
type Kind string

const (
	// Framework abstractions — wrappers.
	KindGuide    Kind = "guide"    // wraps rule (conceptual)
	KindSensor   Kind = "sensor"   // wraps rule (conceptual)
	KindAction   Kind = "action"   // wraps command (projection)
	KindPlaybook Kind = "playbook" // wraps skill   (projection)
	KindPersona  Kind = "persona"  // wraps subagent (projection)

	// Framework abstractions — standalone.
	KindCorpus Kind = "corpus"
	KindEval   Kind = "eval"
	KindSource Kind = "source"

	// Composition primitive. A concern is a reusable fragment of
	// frontmatter + body that other primitives pull in via the
	// `includes:` field. Composition over inheritance — concerns are
	// leaves (cannot include other concerns) and the host primitive
	// always wins on scalar fields. See `primitive/compose.go`.
	KindConcern Kind = "concern"

	// Agent abstractions — escape hatches.
	KindRule     Kind = "rule"
	KindSkill    Kind = "skill"
	KindSubagent Kind = "subagent"
	KindCommand  Kind = "command"
)

// KnownKinds is the closed set the indexer and linter validate against.
var KnownKinds = []Kind{
	KindGuide, KindSensor, KindAction, KindPlaybook, KindPersona,
	KindCorpus, KindEval, KindSource,
	KindConcern,
	KindRule, KindSkill, KindSubagent, KindCommand,
}

// Severity is the rule authority level. Replaces H2-tier parsing.
type Severity string

const (
	SeverityMust   Severity = "must"
	SeverityShould Severity = "should"
	SeverityMay    Severity = "may"
)

// Frontmatter is what each harness file declares between the `---` fences.
// Fields not relevant to a given kind are simply omitted; the indexer
// records what's present and the linter checks what's required.
type Frontmatter struct {
	Kind        string   `yaml:"kind"        json:"kind"`
	ID          string   `yaml:"id"          json:"id"`
	Description string   `yaml:"description" json:"description"`

	Globs    []string `yaml:"globs,omitempty"    json:"globs,omitempty"`
	Phase    string   `yaml:"phase,omitempty"    json:"phase,omitempty"`
	Triggers []string `yaml:"triggers,omitempty" json:"triggers,omitempty"`
	Tools    []string `yaml:"tools,omitempty"    json:"tools,omitempty"`
	Model    string   `yaml:"model,omitempty"    json:"model,omitempty"`
	Args     []Arg    `yaml:"args,omitempty"     json:"args,omitempty"`

	Traces []string `yaml:"traces,omitempty" json:"traces,omitempty"`
	Deps   []string `yaml:"deps,omitempty"   json:"deps,omitempty"`

	Severity string `yaml:"severity,omitempty" json:"severity,omitempty"`
	Tier     string `yaml:"tier,omitempty"     json:"tier,omitempty"`

	// Tags is an orthogonal taxonomy used by `keystone list --tag`.
	// Kebab-case strings; the linter enforces the shape. Tags are
	// union-merged across includes (a concern can contribute tags).
	Tags []string `yaml:"tags,omitempty" json:"tags,omitempty"`

	// Includes lists concerns this primitive composes in. Each id
	// must resolve to a `kind: concern` primitive. Walker resolves
	// at parse time — list fields union (deduped, host's values
	// last), scalar fields are host-wins. Concerns are leaves:
	// a concern cannot itself declare `includes:`. See compose.go.
	Includes []string `yaml:"includes,omitempty" json:"includes,omitempty"`

	// HostTriggers is the per-host activation surface for computational
	// sensors. Same pattern as other primitive frontmatter — declared
	// inline in the sensor's `.keystone/harness/sensors/<id>.md` rather
	// than in a separate config file. Each entry projects to one host
	// hook entry (Claude Code: .claude/settings.json; Cursor: equivalent
	// hook config; etc.) via per-host adapters. Empty for LLM-judgment
	// sensors (review-*, code-debt) — those activate via actions, not
	// ambient hook fire.
	HostTriggers []HostTrigger `yaml:"host_triggers,omitempty" json:"host_triggers,omitempty"`
}

// HostTrigger is one host-native hook declaration in a sensor's
// frontmatter. Field names use snake_case for YAML consistency with
// the rest of keystone's primitive schema.
//
// Phase values: PreToolUse | PostToolUse | Stop | UserPromptSubmit
// (Claude Code's vocabulary; other host adapters translate at their
// projection layer).
type HostTrigger struct {
	Phase   string `yaml:"phase"             json:"phase"`
	Matcher string `yaml:"matcher,omitempty" json:"matcher,omitempty"`
	Command string `yaml:"command"           json:"command"`
	Timeout int    `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

// Arg is a parameter for command and prompt primitives. Free-form on
// purpose — different hosts accept different argument shapes; the
// canonical descriptor just captures what the author declared.
type Arg struct {
	Name        string `yaml:"name"                  json:"name"`
	Type        string `yaml:"type,omitempty"        json:"type,omitempty"`
	Required    bool   `yaml:"required,omitempty"    json:"required,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// Primitive is one indexed harness file: its descriptor plus the
// path the agent opens to read the body. Paths are repo-relative
// POSIX.
//
// Provenance is derived from Path at walk-time — never authored in
// frontmatter. "project" for any harness-root file outside a vendored
// policy; "policy/<name>" for vendored policy content. Answers
// "where did this rule come from?" without the install step needing
// to mutate files.
type Primitive struct {
	Frontmatter `yaml:",inline" json:",inline"`
	Path        string `json:"path"`
	Provenance  string `json:"provenance,omitempty"`
}

// Index is the aggregator emitted at <harness-root>/INDEX.json. The
// agent reads this once at session start and consults Path only when
// it decides to activate the primitive.
type Index struct {
	Version    string                `json:"version"`
	Generated  string                `json:"generated"`
	Primitives []Primitive           `json:"primitives"`
	ByKind     map[string][]string   `json:"by_kind"`
	ByGlob     map[string][]string   `json:"by_glob,omitempty"`
}

// IndexVersion is the schema version emitted into Index.Version. Bumped
// when the descriptor shape changes in a way that consumers must notice.
const IndexVersion = "2.0"
