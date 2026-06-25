// Package mcp builds the keystone MCP server. The server is a thin
// wrapper around `internal/framework/primitive` — it walks the harness
// at request time, exposes the primitive index + bodies as MCP tools
// and resources, and lets host agents (Claude Code, Cursor, Codex,
// etc.) consult the harness without reading every markdown file.
//
// All read paths reuse the same code the CLI runs, so the two
// interfaces (CLI authoring, MCP runtime dispatch) never disagree on
// what the harness contains.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	kconfig "github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// Version is stamped into the MCP `serverInfo` block. Bumped on every
// notable surface change.
const Version = "2.0.0"

// instructions is the short, top-of-server description host agents
// receive when they `initialize` the server. Intentionally terse —
// detailed contracts live in `docs/ports/primitive.md`.
const instructions = `keystone MCP — the harness, served.

The harness lives at .keystone/harness/ and is indexed at
.keystone/INDEX.json. Both this server and the keystone CLI read the
same files; they never disagree.

## Runtime resolution flow

When a scenario arises, work the harness in stages. Only escalate when
the previous stage doesn't carry enough information:

  1. RULES.   Call keystone_list_primitives kind=guide (and kind=sensor)
              to find applicable rules for the touched files / phase.
              Cascade: project wins by default; policies refine via
              nesting in keystone.json; strict items lock absolutely.
  2. CORPUS.  If a rule's body is too terse, call keystone_get_corpus
              id=<rule-id> to open the linked reasoning. Corpus is
              opt-in — never auto-loaded.
  3. EXTERNAL. If corpus is still insufficient, call
              keystone_source_query source=<name> query=<...> to reach
              configured external sources (Linear, Confluence, folders,
              URLs). Configured in .keystone/context.json.
  4. APPLY.   NEVER apply an external-source result silently. Surface
              it to the user and ask: "apply this? at what level —
              project, team policy, org policy?"
  5. CONFLICTS. Any contradiction between rules, corpus, and external
              answers triggers a question to the user. The server
              returns data; the agent reasons about contradictions.

## Read tools

  keystone_list_primitives [kind=<k>] [glob=<g>]
      Filter the index. Returns descriptors only.
  keystone_get_primitive kind=<k> id=<i>
      Returns one primitive's full body.
  keystone_get_corpus id=<rule-id>
      Follow a rule's traces: field; return the linked corpus body.
      Use this when a rule's body is not enough to act.

## Write tools (each shells out to the keystone binary)

  keystone_new_<kind>          Scaffold a new primitive.
                               <kind> ∈ {guide, corpus, sensor, action,
                                         playbook, rule, skill, subagent,
                                         command, adapter, policy}.
  keystone_harness_bootstrap   Scaffold the harness (init equivalent).
  keystone_target_add          Add another agent target.
  keystone_index_refresh       Rebuild .keystone/INDEX.json.
  keystone_project_refresh     Regenerate .claude/ host projections.

## External-source tools

  keystone_source_list                       Names + healths of configured sources.
  keystone_source_query   source=<n> query=<q>   Adapter-routed query.
  keystone_source_health  source=<n>             Reachability + auth state.

## Resources

  keystone://index                     — the full INDEX.json
  keystone://primitive/{kind}/{id}     — one primitive body
  keystone://harness/status            — install audit (layout + counts)
  keystone://source/list               — all configured external sources
  keystone://source/{name}/health      — adapter reachability

Activate by reading the index first, then opening primitive bodies on
demand. Never read every guide/action/sensor file blindly — let the
agent's matching machinery (globs, triggers, phase) decide what fires.

After any write tool, call keystone_index_refresh so INDEX.json stays
current. If the change touched skill/subagent/command, also call
keystone_project_refresh so .claude/ host projections regenerate.
`

// Options configures the server. Mostly project rooting; everything
// else is derived from the harness on disk.
type Options struct {
	// ProjectDir is the consumer project root (the dir holding
	// keystone.json + .keystone/). Defaults to cwd at server start.
	ProjectDir string
}

// New returns an unstarted server. Call Serve to drive it over stdio.
func New(opts Options) (*server.MCPServer, error) {
	if opts.ProjectDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("resolve cwd: %w", err)
		}
		opts.ProjectDir = cwd
	}
	abs, err := filepath.Abs(opts.ProjectDir)
	if err != nil {
		return nil, fmt.Errorf("abs project dir: %w", err)
	}

	s := server.NewMCPServer(
		"keystone",
		Version,
		server.WithInstructions(instructions),
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
	)

	registerTools(s, abs)
	registerWriteTools(s, abs)
	registerSourceTools(s, abs)
	registerResources(s, abs)
	registerSourceResources(s, abs)
	registerSkillResources(s, abs)
	registerPrompts(s, abs)
	registerEvalTools(s, abs)
	registerSearchTool(s, abs)

	return s, nil
}

// Serve runs the server over stdio until the client disconnects or the
// context is cancelled. Returns nil on a clean exit.
func Serve(ctx context.Context, opts Options) error {
	s, err := New(opts)
	if err != nil {
		return err
	}
	return server.ServeStdio(s)
}

// -- tools --------------------------------------------------------------

func registerTools(s *server.MCPServer, projectDir string) {
	s.AddTool(
		mcp.NewTool("keystone_list_primitives",
			mcp.WithDescription("List harness primitives, optionally filtered by kind, glob, or tag(s). Returns descriptors only — open bodies via keystone_get_primitive. The returned descriptors reflect the COMPOSED view (concern `includes:` already merged into list fields)."),
			mcp.WithString("kind",
				mcp.Description("Filter by kind (guide, corpus, sensor, action, playbook, persona, concern, rule, skill, subagent, command, eval, source). Omit for all."),
			),
			mcp.WithString("glob",
				mcp.Description("Filter primitives that declare this glob pattern in their `globs:` frontmatter. Exact-string match on the pattern."),
			),
			mcp.WithArray("tags",
				mcp.Description("Filter primitives that declare ALL of these tags (AND). Tags merge across includes — a concern's tags propagate to the primitives that include it."),
				mcp.WithStringItems(),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			kindFilter := req.GetString("kind", "")
			globFilter := req.GetString("glob", "")
			tagFilter := req.GetStringSlice("tags", nil)

			idx, err := loadIndex(projectDir)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			out := make([]primitive.Primitive, 0, len(idx.Primitives))
			for _, p := range idx.Primitives {
				if kindFilter != "" && p.Kind != kindFilter {
					continue
				}
				if globFilter != "" && !contains(p.Globs, globFilter) {
					continue
				}
				if !hasAllStrings(p.Tags, tagFilter) {
					continue
				}
				out = append(out, p)
			}

			body, err := json.MarshalIndent(map[string]any{
				"count":      len(out),
				"primitives": out,
			}, "", "  ")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(string(body)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("keystone_get_corpus",
			mcp.WithDescription("Open the corpus reasoning linked from a guide or rule. Follows the primitive's `traces:` field — returns one or more corpus bodies. Use this in stage 2 of the resolution flow, only when a rule's body isn't enough to act."),
			mcp.WithString("id",
				mcp.Required(),
				mcp.Description("Source primitive's id (typically a guide or rule). Returns all corpus entries the source traces to."),
			),
			mcp.WithString("kind",
				mcp.Description("Source primitive's kind (default: guide)."),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id, err := req.RequireString("id")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			kind := req.GetString("kind", "guide")

			source, _, err := loadPrimitiveBody(projectDir, kind, id)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if len(source.Corpus) == 0 {
				body, _ := json.MarshalIndent(map[string]any{
					"source_kind":  source.Kind,
					"source_id":    source.ID,
					"traces":       []string{},
					"corpus":       []any{},
					"note":         "this primitive has no traces — no linked corpus",
				}, "", "  ")
				return mcp.NewToolResultText(string(body)), nil
			}

			corpus := []map[string]any{}
			for _, ref := range source.Corpus {
				cKind, cID := parseTraceRef(ref)
				p, body, err := loadPrimitiveBody(projectDir, cKind, cID)
				if err != nil {
					corpus = append(corpus, map[string]any{
						"trace": ref,
						"error": err.Error(),
					})
					continue
				}
				corpus = append(corpus, map[string]any{
					"trace":       ref,
					"kind":        p.Kind,
					"id":          p.ID,
					"path":        p.Path,
					"description": p.Description,
					"body":        body,
				})
			}
			out, _ := json.MarshalIndent(map[string]any{
				"source_kind": source.Kind,
				"source_id":   source.ID,
				"traces":      source.Corpus,
				"corpus":      corpus,
			}, "", "  ")
			return mcp.NewToolResultText(string(out)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("keystone_get_primitive",
			mcp.WithDescription("Return the full body of a single primitive given its kind and id. The body includes the frontmatter and the markdown that follows."),
			mcp.WithString("kind",
				mcp.Required(),
				mcp.Description("Primitive kind (guide, corpus, sensor, action, playbook, rule, skill, subagent, command)."),
			),
			mcp.WithString("id",
				mcp.Required(),
				mcp.Description("Primitive id, e.g. `process/spec`, `verify`, `keystone:index`."),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			kind, err := req.RequireString("kind")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			id, err := req.RequireString("id")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			p, body, err := loadPrimitiveBody(projectDir, kind, id)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			result, _ := json.MarshalIndent(map[string]any{
				"kind":        p.Kind,
				"id":          p.ID,
				"path":        p.Path,
				"description": p.Description,
				"body":        body,
			}, "", "  ")
			return mcp.NewToolResultText(string(result)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("keystone_show",
			mcp.WithDescription("Return one primitive's descriptor PLUS forward + reverse cross-references in a single call. Surfaces: the composed view of the target (tags / tools / globs / host_triggers after `includes:` resolution), `included_by` (when target is a concern), `traces` (forward) and `traced_by` (reverse) for guide↔corpus pairs, severity / model / phase / provenance. Use this when you need to understand a primitive's neighborhood in one shot — saves N follow-up keystone_get_primitive calls."),
			mcp.WithString("kind",
				mcp.Required(),
				mcp.Description("Primitive kind (guide, sensor, persona, playbook, concern, action, corpus, etc.)."),
			),
			mcp.WithString("id",
				mcp.Required(),
				mcp.Description("Primitive id."),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			kind, err := req.RequireString("kind")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			id, err := req.RequireString("id")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			idx, err := loadIndex(projectDir)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			view, found := buildShowView(idx.Primitives, kind, id)
			if !found {
				return mcp.NewToolResultError(fmt.Sprintf("no primitive with kind=%s id=%s", kind, id)), nil
			}
			out, err := json.MarshalIndent(view, "", "  ")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(string(out)), nil
		},
	)
}

// hasAllStrings reports whether `have` contains every element of
// `want`. Empty `want` is a vacuous pass. Used by the tag-filter on
// keystone_list_primitives — AND semantics, not OR.
func hasAllStrings(have, want []string) bool {
	if len(want) == 0 {
		return true
	}
	set := make(map[string]struct{}, len(have))
	for _, h := range have {
		set[h] = struct{}{}
	}
	for _, w := range want {
		if _, ok := set[w]; !ok {
			return false
		}
	}
	return true
}

// showView is the structured snapshot keystone_show returns. Mirrors
// the CLI's `keystone show` output so MCP clients see the same shape
// the user sees in their terminal.
type showView struct {
	Kind         string                  `json:"kind"`
	ID           string                  `json:"id"`
	Description  string                  `json:"description"`
	Path         string                  `json:"path"`
	Provenance   string                  `json:"provenance,omitempty"`
	Severity     string                  `json:"severity,omitempty"`
	Model        string                  `json:"model,omitempty"`
	Phase        string                  `json:"phase,omitempty"`
	Tags         []string                `json:"tags,omitempty"`
	Tools        []string                `json:"tools,omitempty"`
	Globs        []string                `json:"globs,omitempty"`
	Triggers     []string                `json:"triggers,omitempty"`
	Includes     []string                `json:"includes,omitempty"`
	IncludedBy   []string                `json:"included_by,omitempty"`
	Traces       []string                `json:"traces,omitempty"`
	TracedBy     []string                `json:"traced_by,omitempty"`
	HostTriggers []primitive.HostTrigger `json:"host_triggers,omitempty"`
}

// buildShowView assembles the showView for one (kind, id) pair from
// the INDEX primitive set. The INDEX records the COMPOSED descriptor
// (Compose has already merged concern contributions), so the returned
// list fields reflect what the agent actually sees.
func buildShowView(primitives []primitive.Primitive, kind, id string) (showView, bool) {
	var target primitive.Primitive
	found := false
	for _, p := range primitives {
		if p.Kind == kind && p.ID == id {
			target = p
			found = true
			break
		}
	}
	if !found {
		return showView{}, false
	}
	v := showView{
		Kind:         target.Kind,
		ID:           target.ID,
		Description:  target.Description,
		Path:         target.Path,
		Provenance:   target.Provenance,
		Severity:     target.Severity,
		Model:        target.Model,
		Phase:        target.Phase,
		Tags:         target.Tags,
		Tools:        target.Tools,
		Globs:        target.Globs,
		Triggers:     target.Triggers,
		Includes:     target.Includes,
		Traces:       target.Corpus,
		HostTriggers: target.HostTriggers,
	}
	// Reverse-lookup: who includes this primitive (only meaningful
	// when target is a concern, but unconditional so future kinds
	// that become include-able just work).
	for _, p := range primitives {
		for _, inc := range p.Includes {
			if inc == target.ID {
				v.IncludedBy = append(v.IncludedBy, p.Kind+"/"+p.ID)
			}
		}
	}
	sort.Strings(v.IncludedBy)
	// Reverse-lookup: who traces to this primitive (guide → corpus).
	for _, p := range primitives {
		for _, tr := range p.Corpus {
			if tr == target.ID {
				v.TracedBy = append(v.TracedBy, p.Kind+"/"+p.ID)
			}
		}
	}
	sort.Strings(v.TracedBy)
	return v, true
}

// -- resources ----------------------------------------------------------

func registerResources(s *server.MCPServer, projectDir string) {
	s.AddResource(
		mcp.NewResource("keystone://index",
			"Harness primitive index",
			mcp.WithResourceDescription("The full .keystone/INDEX.json — descriptors for every primitive in the harness."),
			mcp.WithMIMEType("application/json"),
		),
		func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			path := filepath.Join(projectDir, kconfig.KeystoneDir(kconfig.DefaultHarnessRoot), kconfig.IndexName)
			data, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("read index: %w", err)
			}
			return []mcp.ResourceContents{
				mcp.TextResourceContents{
					URI:      req.Params.URI,
					MIMEType: "application/json",
					Text:     string(data),
				},
			}, nil
		},
	)

	s.AddResource(
		mcp.NewResource("keystone://harness/status",
			"Harness install status",
			mcp.WithResourceDescription("Summary of what's installed at .keystone/harness/: primitive count by kind, INDEX.json freshness."),
			mcp.WithMIMEType("application/json"),
		),
		func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			idx, err := loadIndex(projectDir)
			if err != nil {
				return nil, err
			}
			body, _ := json.MarshalIndent(map[string]any{
				"version":         idx.Version,
				"generated":       idx.Generated,
				"primitive_count": len(idx.Primitives),
				"by_kind":         idx.ByKind,
				"project_dir":     projectDir,
			}, "", "  ")
			return []mcp.ResourceContents{
				mcp.TextResourceContents{
					URI:      req.Params.URI,
					MIMEType: "application/json",
					Text:     string(body),
				},
			}, nil
		},
	)

	// Resource template: primitive bodies addressable by URI.
	s.AddResourceTemplate(
		mcp.NewResourceTemplate("keystone://primitive/{kind}/{id}",
			"Primitive body",
			mcp.WithTemplateDescription("Returns one primitive's body. URI form: keystone://primitive/<kind>/<id> (id may use slashes for hierarchical primitives like guides)."),
			mcp.WithTemplateMIMEType("text/markdown"),
		),
		func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			kind, id, err := parsePrimitiveURI(req.Params.URI)
			if err != nil {
				return nil, err
			}
			_, body, err := loadPrimitiveBody(projectDir, kind, id)
			if err != nil {
				return nil, err
			}
			return []mcp.ResourceContents{
				mcp.TextResourceContents{
					URI:      req.Params.URI,
					MIMEType: "text/markdown",
					Text:     body,
				},
			}, nil
		},
	)
}

// -- helpers ------------------------------------------------------------

func loadIndex(projectDir string) (*primitive.Index, error) {
	// Prefer the on-disk INDEX.json if fresh; otherwise re-walk. The
	// re-walk is cheap (filesystem-bound) and the alternative would be
	// stale-data hazards if a primitive was added since the last
	// `keystone index` run.
	primitives, _, err := primitive.Walk(projectDir, kconfig.DefaultHarnessRoot)
	if err != nil {
		return nil, err
	}
	idx := primitive.Build(primitives, time.Now())
	return &idx, nil
}

func loadPrimitiveBody(projectDir, kind, id string) (primitive.Primitive, string, error) {
	primitives, _, err := primitive.Walk(projectDir, kconfig.DefaultHarnessRoot)
	if err != nil {
		return primitive.Primitive{}, "", err
	}
	for _, p := range primitives {
		if p.Kind == kind && p.ID == id {
			data, err := os.ReadFile(filepath.Join(projectDir, p.Path))
			if err != nil {
				return p, "", fmt.Errorf("read %s: %w", p.Path, err)
			}
			return p, string(data), nil
		}
	}
	return primitive.Primitive{}, "", fmt.Errorf("no primitive with kind=%q id=%q", kind, id)
}

// parseTraceRef interprets a `traces:` entry. The canonical form is
// `corpus/<topic>/<name>`; the kind defaults to "corpus" when not
// explicitly prefixed. Accepts both forms.
func parseTraceRef(ref string) (kind, id string) {
	const corpusPrefix = "corpus/"
	if strings.HasPrefix(ref, corpusPrefix) {
		return "corpus", ref
	}
	// Fallback: assume corpus.
	return "corpus", "corpus/" + ref
}

func parsePrimitiveURI(uri string) (kind, id string, err error) {
	const prefix = "keystone://primitive/"
	if !strings.HasPrefix(uri, prefix) {
		return "", "", fmt.Errorf("URI must start with %s", prefix)
	}
	rest := strings.TrimPrefix(uri, prefix)
	slash := strings.IndexByte(rest, '/')
	if slash < 0 {
		return "", "", fmt.Errorf("URI must include both kind and id: %s", uri)
	}
	return rest[:slash], rest[slash+1:], nil
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
