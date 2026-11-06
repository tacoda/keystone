// Package mcp builds the keystone MCP server. The server is a thin
// wrapper around `internal/framework/primitive` — it walks the charter
// at request time, exposes the primitive index + bodies as MCP tools
// and resources, and lets host agents (Claude Code, Cursor, Codex,
// etc.) consult the charter without reading every markdown file.
//
// All read paths reuse the same code the CLI runs, so the two
// interfaces (CLI authoring, MCP runtime dispatch) never disagree on
// what the charter contains.
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
const Version = "3.0.0" // 5.0: constraint/capability kinds, keystone_get_reference

// instructions is the short, top-of-server description host agents
// receive when they `initialize` the server. Intentionally terse —
// detailed contracts live in `docs/ports/primitive.md`.
const instructions = `keystone MCP — the charter, served.

The charter's sources live in .agents/ and are indexed at
.keystone/INDEX.json. This server and the keystone CLI read the same
files; they never disagree.

## Two categories

Constraints limit, gate, or check what the agent does:
  rule        applies when a touched file matches its paths: (or everywhere)
  reference   the reasoning behind a rule, opened on demand
  hook        a check that runs on an event; a non-zero exit blocks
  permission  allow / ask / deny tool rules
  eval        a test that the charter changes agent behavior

Capabilities add things the agent can do:
  skill       Agent Skills SKILL.md; activates when its description matches
  command     a unit of work the user asks for by name
  subagent    a role the agent delegates to
  tool        an author-defined callable (transport: mcp is served here)
  document    a template for governed output

## Runtime resolution flow

When a scenario arises, work the charter in stages. Only escalate when
the previous stage does not carry enough information:

  1. RULES.      Call keystone_list_primitives kind=rule to find the rules
                 for the touched files and phase. Cascade: the project
                 wins by default; policies refine; strict items are absolute.
  2. REFERENCES. If a rule's body is too terse, call
                 keystone_get_reference id=<rule-id> to open the reasoning
                 it links (its corpus: field). Never auto-loaded.
  3. CONFLICTS.  A contradiction between a rule and its reference is a
                 question for the user. The server returns data; the agent
                 reasons about contradictions.

## Read tools

  keystone_list_primitives [kind=<k>|category=<c>] [paths=<p>] [tags=…]
      Filter the index. Returns descriptors only.
  keystone_get_primitive kind=<k> id=<i>
      Returns one primitive's full body.
  keystone_get_reference id=<rule-id>
      Follow a rule's corpus: field; return the linked reference bodies.
  keystone_signal_list
      List signals (events a hook, tool, or subagent subscribes to via
      event:) and their subscribers, plus the bridged host events.
  keystone_charter_coverage
      Files a rule governs vs. uncharted (matched by no rule's paths:).
  keystone_explain id=<i> [kind=<k>]
      Explain a primitive — how it activates, what it links to, where it
      projects. Use before reading a whole body.
  keystone_budget
      Tokens each host loads every session, tokens on demand per kind,
      and any keystone.json budgets cap exceeded.
  keystone_charter_conformance
      Rubric verdict (CONFORMANT|DRIFTING|NON-CONFORMANT) from objective
      checks: cascade, frontmatter validity, pairing, coverage.

## Reactions (what fires when)

A signal or host event fires; the reacting primitive subscribes via
event:, like a skill declares triggers:
  - hook      — a check → verdict (exit status); a failure blocks.
  - tool      — an external callable; on demand, or a side effect on an event.
  - subagent  — an inferential review → structured returns:.

## Write tools (each shells out to the keystone binary)

  keystone_new_<kind>          Scaffold a new primitive. <kind> is one of
                               rule, reference, hook, permission, eval,
                               skill, command, subagent, tool, document,
                               adapter, policy.
  keystone_charter_bootstrap   Scaffold the charter (init equivalent).
  keystone_target_add          Add another host.
  keystone_index_refresh       Rebuild .keystone/INDEX.json.
  keystone_project_refresh     Regenerate every host's projection.

## Resources

  keystone://index                     — the full INDEX.json
  keystone://primitive/{kind}/{id}     — one primitive body
  keystone://charter/status            — install audit (layout + counts)

Activate by reading the index first, then opening primitive bodies on
demand. Never read every rule, hook, or command file blindly — let the
host's matching (paths:, descriptions, events) decide what applies.

After any write tool, call keystone_index_refresh so INDEX.json stays
current, then keystone_project_refresh so every host sees the change.
`

// Options configures the server. Mostly project rooting; everything
// else is derived from the charter on disk.
type Options struct {
	// ProjectDir is the consumer project root (the dir holding
	// keystone.json + .agents/). Defaults to cwd at server start.
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
	registerResources(s, abs)
	registerSkillResources(s, abs)
	registerPrompts(s, abs)
	registerEvalTools(s, abs)
	registerSearchTool(s, abs)
	registerCharterViews(s, abs)
	registerToolPrimitives(s, abs)

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
			mcp.WithDescription("List charter primitives, optionally filtered by kind, category, path pattern, or tag(s). Returns descriptors only — open bodies via keystone_get_primitive. The returned descriptors reflect the COMPOSED view (`includes:` already merged into list fields)."),
			mcp.WithString("kind",
				mcp.Description("Filter by kind: a constraint (rule, reference, hook, permission, eval) or a capability (skill, command, subagent, tool, document). Omit for all."),
			),
			mcp.WithString("category",
				mcp.Description("Filter by category: constraint or capability. Omit for both."),
			),
			mcp.WithString("paths",
				mcp.Description("Filter primitives that declare this pattern in their `paths:` frontmatter. Exact-string match on the pattern. (`glob` is the 4.x name.)"),
			),
			mcp.WithArray("tags",
				mcp.Description("Filter primitives that declare ALL of these tags (AND). Tags merge across includes."),
				mcp.WithStringItems(),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			f := listFilter{
				kind:     req.GetString("kind", ""),
				category: req.GetString("category", ""),
				path:     req.GetString("paths", req.GetString("glob", "")),
				tags:     req.GetStringSlice("tags", nil),
			}

			idx, err := loadIndex(projectDir)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			out := make([]primitive.Primitive, 0, len(idx.Primitives))
			for _, p := range idx.Primitives {
				if f.match(p) {
					out = append(out, p)
				}
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
		mcp.NewTool("keystone_get_reference",
			mcp.WithDescription("Open the references linked from a rule. Follows the rule's `corpus:` field — returns one or more reference bodies. Use this in stage 2 of the resolution flow, only when a rule's body isn't enough to act."),
			mcp.WithString("id",
				mcp.Required(),
				mcp.Description("Source primitive's id (typically a rule). Returns every reference it links."),
			),
			mcp.WithString("kind",
				mcp.Description("Source primitive's kind (default: rule)."),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id, err := req.RequireString("id")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			kind := req.GetString("kind", string(primitive.KindRule))

			source, _, err := loadPrimitiveBody(projectDir, kind, id)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if len(source.Corpus) == 0 {
				body, _ := json.MarshalIndent(map[string]any{
					"source_kind": source.Kind,
					"source_id":   source.ID,
					"traces":      []string{},
					"references":  []any{},
					"note":        "this primitive links no reference",
				}, "", "  ")
				return mcp.NewToolResultText(string(body)), nil
			}

			refs := []map[string]any{}
			for _, ref := range source.Corpus {
				p, body, err := loadReference(projectDir, ref)
				if err != nil {
					refs = append(refs, map[string]any{
						"trace": ref,
						"error": err.Error(),
					})
					continue
				}
				refs = append(refs, map[string]any{
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
				"references":  refs,
			}, "", "  ")
			return mcp.NewToolResultText(string(out)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("keystone_get_primitive",
			mcp.WithDescription("Return the full body of a single primitive given its kind and id. The body includes the frontmatter and the markdown that follows."),
			mcp.WithString("kind",
				mcp.Required(),
				mcp.Description("Primitive kind: rule, reference, hook, permission, eval, skill, command, subagent, tool, or document (4.x names are accepted)."),
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
			mcp.WithDescription("Return one primitive's descriptor PLUS forward + reverse cross-references in a single call. Surfaces: the composed view of the target (tags / tools / paths / host_triggers after `includes:` resolution), `included_by` (when another primitive includes it), `traces` (forward) and `traced_by` (reverse) for rule↔reference pairs, severity / model / phase / provenance. Use this when you need to understand a primitive's neighborhood in one shot — saves N follow-up keystone_get_primitive calls."),
			mcp.WithString("kind",
				mcp.Required(),
				mcp.Description("Primitive kind: rule, reference, hook, permission, eval, skill, command, subagent, tool, or document."),
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
	Paths        []string                `json:"paths,omitempty"`
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
		if primitive.IsKind(p.Kind, kind) && p.ID == id {
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
		Paths:        target.Paths,
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
			"Charter primitive index",
			mcp.WithResourceDescription("The full .keystone/INDEX.json — descriptors for every primitive in the charter."),
			mcp.WithMIMEType("application/json"),
		),
		func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			path := filepath.Join(projectDir, kconfig.StateRoot, kconfig.IndexName)
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
		mcp.NewResource("keystone://charter/status",
			"Charter install status",
			mcp.WithResourceDescription("Summary of what's installed at .agents/: primitive count by kind, INDEX.json freshness."),
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
			mcp.WithTemplateDescription("Returns one primitive's body. URI form: keystone://primitive/<kind>/<id> (id may use slashes for hierarchical primitives like rules)."),
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
	primitives, _, err := primitive.Walk(projectDir, kconfig.DefaultAgentsRoot)
	if err != nil {
		return nil, err
	}
	idx := primitive.Build(primitives, time.Now())
	return &idx, nil
}

func loadPrimitiveBody(projectDir, kind, id string) (primitive.Primitive, string, error) {
	primitives, _, err := primitive.Walk(projectDir, kconfig.DefaultAgentsRoot)
	if err != nil {
		return primitive.Primitive{}, "", err
	}
	for _, p := range primitives {
		if primitive.IsKind(p.Kind, kind) && p.ID == id {
			data, err := os.ReadFile(filepath.Join(projectDir, p.Path))
			if err != nil {
				return p, "", fmt.Errorf("read %s: %w", p.Path, err)
			}
			return p, string(data), nil
		}
	}
	return primitive.Primitive{}, "", fmt.Errorf("no primitive with kind=%q id=%q", kind, id)
}

// listFilter is keystone_list_primitives' filter; an empty field matches all.
type listFilter struct {
	kind, category, path string
	tags                 []string
}

func (f listFilter) match(p primitive.Primitive) bool {
	if f.kind != "" && !primitive.IsKind(p.Kind, f.kind) {
		return false
	}
	if f.category != "" && string(primitive.CategoryOf(primitive.Kind(p.Kind))) != f.category {
		return false
	}
	if f.path != "" && !contains(p.Paths, f.path) {
		return false
	}
	return hasAllStrings(p.Tags, f.tags)
}

// loadReference resolves a `corpus:` entry to its reference: by id, or by
// the 4.x id that carries a `corpus/` prefix.
func loadReference(projectDir, ref string) (primitive.Primitive, string, error) {
	kind := string(primitive.KindReference)
	p, body, err := loadPrimitiveBody(projectDir, kind, ref)
	if err != nil && !strings.HasPrefix(ref, "corpus/") {
		return loadPrimitiveBody(projectDir, kind, "corpus/"+ref)
	}
	return p, body, err
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
