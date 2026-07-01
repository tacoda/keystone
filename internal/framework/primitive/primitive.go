// Package primitive defines the canonical descriptor that every charter
// file declares in its frontmatter, plus the index aggregator that the
// agent reads in lieu of the raw markdown tree.
//
// See docs/ports/primitive.md for the canonical schema. Bodies are not
// stored in this package — only descriptors. The agent opens body files
// (via the Path field) on activation.
package primitive

import "strings"

// Kind names every primitive that can appear in the charter. The set is
// closed; unknown kinds are a lint error.
//
// 2.0 explicitly excludes MCP prompts: prompts only make sense when an
// MCP server and a client UI surface them to the human, and a CLI-only
// charter has no path to invoke them. Slash commands cover human-
// triggered templated workflows; skills cover agent-self-triggered
// patterns. A future MCP-only catalog can live outside this contract.
//
// Two-layer taxonomy. Framework primitives wrap agent primitives the
// way an ORM wraps SQL: the framework primitive is the canonical
// authoring surface, the agent primitive is the raw host-native
// equivalent kept as an escape hatch.
//
//	Framework abstractions (encouraged by default — author here):
//	  wrappers       guide, sensor (→ rule)
//	                 action        (→ command)
//	                 playbook      (→ skill)
//	                 persona       (→ subagent)
//	  standalone     corpus, eval, source
//
//	Agent abstractions (raw host-native passthrough — escape hatches):
//	  rule, skill, subagent, command
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
	// 3.0 canonical vocabulary. A keystone primitive is to an agent
	// primitive as an ActiveRecord model is to a DB table: same concept,
	// plus validations / associations / conveniences the bare host file
	// lacks. The keystone names (guide/sensor/playbook) stay where the
	// abstraction maps 1-to-many to host mechanisms by `mode:`; the host
	// names (command/skill/agent) stay where the concept is identical.
	KindGuide    Kind = "guide"    // ambient, glob-scoped directive → rule (inferential) | LSP (computational)
	KindSensor   Kind = "sensor"   // a check that reacts to a signal/phase → verdict (exit/HTTP status); gates
	KindCommand  Kind = "command"  // a unit of work / lifecycle step (was action)
	KindSkill    Kind = "skill"    // composed capability
	KindPlaybook Kind = "playbook" // composed sequence of commands with gates
	KindAgent    Kind = "agent"    // a role spawned as a subagent (was persona/subagent)
	KindPattern  Kind = "pattern"  // a recurring recipe, in prose; on-demand
	KindPosture  Kind = "posture"  // tool/permission posture → settings.json permissions
	KindTool     Kind = "tool"     // author-defined callable → keystone MCP server registration
	KindDocument Kind = "document" // governed output doc; `type:` carries the subtype (e.g. feature)

	// Standalone primitives.
	KindCorpus Kind = "corpus" // deep reasoning, loaded on-demand
	KindEval   Kind = "eval"
	// `source` is no longer an authorable kind: external-system access is a
	// `tool` (transport: cli=curl / mcp).

	// Composition primitive. A concern is a reusable fragment of
	// frontmatter + body that other primitives pull in via the
	// `includes:` field. It also routes to corpus (carries `corpus:`
	// refs) without inlining the reasoning. Concerns are leaves and the
	// host primitive always wins on scalar fields. See compose.go.
	KindConcern Kind = "concern"
)

// KnownKinds is the closed set the indexer and linter validate against.
// `rule` is intentionally absent — it is a projection-target name, not an
// authorable kind (author a `guide`).
var KnownKinds = []Kind{
	KindGuide, KindSensor, KindCommand, KindSkill, KindPlaybook,
	KindAgent, KindPattern, KindPosture, KindTool, KindDocument,
	KindCorpus, KindEval, KindConcern,
}

// canonicalDirKind maps a canonical charter subdirectory to the kind it
// holds. Convention over configuration: a file's directory is its kind
// declaration, so `kind:` may be omitted from frontmatter and inferred.
// No `rules/` entry — `rule` is not a kind.
var canonicalDirKind = map[string]Kind{
	"guides":    KindGuide,
	"sensors":   KindSensor,
	"commands":  KindCommand,
	"skills":    KindSkill,
	"playbooks": KindPlaybook,
	"agents":    KindAgent,
	"patterns":  KindPattern,
	"posture":   KindPosture,
	"tools":     KindTool,
	"documents": KindDocument,
	"corpus":    KindCorpus,
	"concerns":  KindConcern,
	"evals":     KindEval,
}

// resolveKind returns the explicit kind when set, else the kind inferred
// from the path. Convention over configuration with an explicit override.
func resolveKind(explicit, relPath string) string {
	if strings.TrimSpace(explicit) != "" {
		return explicit
	}
	return string(InferKind(relPath))
}

// InferKind returns the kind implied by a primitive's canonical
// directory — the path segment immediately under "charter/" — or "" when
// the path is off-convention. relPath is project-relative POSIX. This is
// the convention-over-configuration entry point: Walk fills an empty
// `kind:` from the directory, so authors need not declare it.
func InferKind(relPath string) Kind {
	const marker = "charter/"
	i := strings.Index(relPath, marker)
	if i < 0 {
		return ""
	}
	rest := relPath[i+len(marker):]
	seg := rest
	if j := strings.Index(rest, "/"); j >= 0 {
		seg = rest[:j]
	}
	return canonicalDirKind[seg]
}

// Severity is the rule authority level. Replaces H2-tier parsing.
type Severity string

const (
	SeverityMust   Severity = "must"
	SeverityShould Severity = "should"
	SeverityMay    Severity = "may"
)

// GuideTier is the authority level of an inferential guide — the rule it
// projects to. `preference` is the default; `iron-law` and `golden-rule`
// are reserved for the rare directives that warrant them.
type GuideTier string

const (
	TierIronLaw    GuideTier = "iron-law"
	TierGoldenRule GuideTier = "golden-rule"
	TierPreference GuideTier = "preference"
)

// Frontmatter is what each charter file declares between the `---` fences.
// Fields not relevant to a given kind are simply omitted; the indexer
// records what's present and the linter checks what's required.
type Frontmatter struct {
	Kind        string `yaml:"kind"        json:"kind"`
	ID          string `yaml:"id"          json:"id"`
	Description string `yaml:"description" json:"description"`

	Globs    []string `yaml:"globs,omitempty"    json:"globs,omitempty"`
	Phase    string   `yaml:"phase,omitempty"    json:"phase,omitempty"`
	Triggers []string `yaml:"triggers,omitempty" json:"triggers,omitempty"`
	Tools    []string `yaml:"tools,omitempty"    json:"tools,omitempty"`
	Model    string   `yaml:"model,omitempty"    json:"model,omitempty"`
	Args     []Arg    `yaml:"args,omitempty"     json:"args,omitempty"`

	// Mode picks the computational vs inferential nature of a guide /
	// sensor / hook. computational → runs a `run:` shell command/script;
	// inferential → dispatches an `agent:` whose output conforms to a
	// `returns:` schema. Default: guide → inferential, sensor →
	// computational.
	Mode string `yaml:"mode,omitempty" json:"mode,omitempty"`

	// Event is the signal or host phase a primitive subscribes to via its
	// `on:` field — a sensor (check), tool (side-effect), or agent (review)
	// self-declares what fires it, the way a skill declares `triggers:`.
	// A host phase (PreToolUse…) bridges into the host; a signal
	// (pre-verify, on-gate, a custom name) is keystone-fired. The YAML key
	// is `on:`; `event:` is accepted as a back-compat alias.
	Event string `yaml:"on,omitempty" json:"on,omitempty"`

	// Run is the shell command/script a computational hook/sensor/tool
	// executes. NOT the `command` kind — that is an agent-driven unit of
	// work; `run:` is a plain shell string.
	Run string `yaml:"run,omitempty" json:"run,omitempty"`

	// Transport is how a `tool` callable reaches the agent: cli | mcp |
	// plugin. keystone's generic sense of "tool" — not MCP-only.
	Transport string `yaml:"transport,omitempty" json:"transport,omitempty"`

	// Agent is the agent an inferential hook/sensor dispatches; Returns is
	// the structured-result schema that agent must emit (the dispatcher
	// validates against it and surfaces it as feedback).
	Agent   string `yaml:"agent,omitempty"   json:"agent,omitempty"`
	Returns string `yaml:"returns,omitempty" json:"returns,omitempty"`

	// Allow / Ask / Deny are a posture's tool-permission lists. They
	// project to the host's permissions block (Claude Code:
	// .claude/settings.json `permissions`).
	Allow []string `yaml:"allow,omitempty" json:"allow,omitempty"`
	Ask   []string `yaml:"ask,omitempty"   json:"ask,omitempty"`
	Deny  []string `yaml:"deny,omitempty"  json:"deny,omitempty"`

	// Corpus cites the reasoning behind this primitive (renamed from
	// `traces:` in 3.0 — plain over jargon). Loaded on-demand: the agent
	// opens the corpus body only when it needs the why.
	Corpus []string `yaml:"corpus,omitempty" json:"corpus,omitempty"`

	// Deps is retired from the authoring surface in 3.0 (nothing loaded
	// it). Kept on the descriptor for back-compat reads; not scaffolded,
	// not documented. "See also" is a prose [[link]] in the corpus body.
	Deps []string `yaml:"deps,omitempty" json:"deps,omitempty"`

	// Document graph (3.0). A command `produces:` / `consumes:` document
	// ids; `stop:` is its explicit done-condition. A document declares
	// its own `gates:` lifecycle, a `type:` subtype (e.g. feature), and
	// `produced_by:` the command that writes it. `supersedes:` lets a new
	// rule or document flag the older one it replaces.
	Produces []string `yaml:"produces,omitempty"    json:"produces,omitempty"`
	Consumes []string `yaml:"consumes,omitempty"    json:"consumes,omitempty"`
	Stop     string   `yaml:"stop,omitempty"        json:"stop,omitempty"`
	Gates    []string `yaml:"gates,omitempty"       json:"gates,omitempty"`
	// Gate is the current lifecycle state of a document *instance* (under
	// .charter/work/). Templates declare `gates:` (the ordered set);
	// instances carry `gate:` (where they are now). `keystone document
	// promote` advances it.
	Gate       string   `yaml:"gate,omitempty"        json:"gate,omitempty"`
	Type       string   `yaml:"type,omitempty"        json:"type,omitempty"`
	ProducedBy string   `yaml:"produced_by,omitempty" json:"produced_by,omitempty"`
	Supersedes []string `yaml:"supersedes,omitempty"  json:"supersedes,omitempty"`

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
	// inline in the sensor's `.charter/sensors/<id>.md` rather
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

// Primitive is one indexed charter file: its descriptor plus the
// path the agent opens to read the body. Paths are repo-relative
// POSIX.
//
// Provenance is derived from Path at walk-time — never authored in
// frontmatter. "project" for any charter-root file outside a vendored
// policy; "policy/<name>" for vendored policy content. Answers
// "where did this rule come from?" without the install step needing
// to mutate files.
type Primitive struct {
	Frontmatter `yaml:",inline" json:",inline"`
	Path        string `json:"path"`
	Provenance  string `json:"provenance,omitempty"`
}

// Index is the aggregator emitted at <charter-root>/INDEX.json. The
// agent reads this once at session start and consults Path only when
// it decides to activate the primitive.
type Index struct {
	Version    string              `json:"version"`
	Generated  string              `json:"generated"`
	Primitives []Primitive         `json:"primitives"`
	ByKind     map[string][]string `json:"by_kind"`
	ByGlob     map[string][]string `json:"by_glob,omitempty"`
}

// IndexVersion is the schema version emitted into Index.Version. Bumped
// when the descriptor shape changes in a way that consumers must notice.
const IndexVersion = "2.0"
