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
// Two-layer taxonomy:
//
//   Framework abstractions (encouraged by default — keystone's own
//   structured primitives):
//     - guide, corpus, sensor, action, playbook
//
//   Agent abstractions (host-native primitives kept first-class so
//   users can extend what the host already understands):
//     - rule, skill, subagent, command
//
// Both layers share the same canonical frontmatter shape and land in
// the same INDEX.json. The distinction is *conceptual* — what each
// primitive's contract is, not how it's stored or served.
type Kind string

const (
	// Framework abstractions.
	KindGuide    Kind = "guide"
	KindCorpus   Kind = "corpus"
	KindSensor   Kind = "sensor"
	KindAction   Kind = "action"
	KindPlaybook Kind = "playbook"
	KindEval     Kind = "eval"
	KindSource   Kind = "source"

	// Agent abstractions.
	KindRule     Kind = "rule"
	KindSkill    Kind = "skill"
	KindSubagent Kind = "subagent"
	KindCommand  Kind = "command"
	KindPersona  Kind = "persona"
)

// KnownKinds is the closed set the indexer and linter validate against.
var KnownKinds = []Kind{
	KindGuide, KindCorpus, KindSensor, KindAction, KindPlaybook, KindEval, KindSource,
	KindRule, KindSkill, KindSubagent, KindCommand, KindPersona,
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
	Args     []Arg    `yaml:"args,omitempty"     json:"args,omitempty"`

	Traces []string `yaml:"traces,omitempty" json:"traces,omitempty"`
	Deps   []string `yaml:"deps,omitempty"   json:"deps,omitempty"`

	Severity string `yaml:"severity,omitempty" json:"severity,omitempty"`
	Tier     string `yaml:"tier,omitempty"     json:"tier,omitempty"`
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
