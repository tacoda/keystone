package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/tacoda/keystone/internal/framework/budget"
	"github.com/tacoda/keystone/internal/framework/charter"
	kconfig "github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// registerCharterViews exposes the read-side charter views — signals and
// coverage — over MCP, so a host agent sees the same concepts the CLI and
// dashboard do.
func registerCharterViews(s *server.MCPServer, projectDir string) {
	registerSignalListTool(s, projectDir)
	registerQueryTool(s, projectDir)
	registerCoverageTool(s, projectDir)
	registerCharterMover(s, projectDir)
	registerExplainTool(s, projectDir)
	registerConformanceTool(s, projectDir)
	registerBudgetTool(s, projectDir)
}

// registerBudgetTool reports the context budget: what each host loads into
// every session, what loads on demand, and any keystone.json cap exceeded.
func registerBudgetTool(s *server.MCPServer, projectDir string) {
	s.AddTool(
		mcp.NewTool("keystone_budget",
			mcp.WithDescription("Report the context budget: per host, the tokens loaded every session (root files, skill/command/subagent listings, pathless rules); on demand, the tokens per kind and the largest files; and every keystone.json budgets cap exceeded. Word-count estimate (real tokenizers count about 1.35x more)."),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			r, err := budget.MeasureProject(projectDir)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			body, _ := json.MarshalIndent(r, "", "  ")
			return mcp.NewToolResultText(string(body)), nil
		},
	)
}

// registerConformanceTool reports the charter-conformance rubric: an
// objective PASS/PARTIAL/FAIL per criterion plus an overall verdict.
func registerConformanceTool(s *server.MCPServer, projectDir string) {
	s.AddTool(
		mcp.NewTool("keystone_charter_conformance",
			mcp.WithDescription("Score the charter-conformance rubric: cascade integrity, frontmatter validity, rule↔reference pairing, and coverage — each PASS/PARTIAL/FAIL — with an overall verdict (CONFORMANT | DRIFTING | NON-CONFORMANT)."),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			rub, err := charter.Conformance(projectDir, kconfig.DefaultAgentsRoot)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			body, _ := json.MarshalIndent(rub, "", "  ")
			return mcp.NewToolResultText(string(body)), nil
		},
	)
}

// registerExplainTool explains a primitive — how it activates, what it
// links to, where it projects — so an agent can understand a skill,
// command, sensor, guide, etc. without reading its whole body.
func registerExplainTool(s *server.MCPServer, projectDir string) {
	s.AddTool(
		mcp.NewTool("keystone_explain",
			mcp.WithDescription("Explain a primitive by id: its kind, how/when it activates (a rule's paths:, a hook's event:, a skill's description, …), what it links to, and where it projects. Optionally narrow by kind."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Primitive id (e.g. 'keystone:verify', 'idioms/go/stdlib-first').")),
			mcp.WithString("kind", mcp.Description("Narrow to this kind when an id is shared across kinds.")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id, err := req.RequireString("id")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			kind := req.GetString("kind", "")
			prims, _, err := primitive.Walk(projectDir, kconfig.DefaultAgentsRoot)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			for _, p := range prims {
				if charter.Matches(p, id, kind) {
					body, _ := json.MarshalIndent(charter.Explain(p), "", "  ")
					return mcp.NewToolResultText(string(body)), nil
				}
			}
			return mcp.NewToolResultError("no primitive with id " + id), nil
		},
	)
}

// registerSignalListTool lists the signals a hook/sensor/tool/agent may
// subscribe to via `on:` — built-ins plus project-declared — with the
// primitives that currently subscribe, and the host phases (bridged).
func registerSignalListTool(s *server.MCPServer, projectDir string) {
	s.AddTool(
		mcp.NewTool("keystone_signal_list",
			mcp.WithDescription("List keystone signals (framework events a primitive subscribes to via `on:`): built-in + project-declared, each with its current subscribers, plus the closed set of host phases (bridged, not signals)."),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var custom []string
			if cfg, err := kconfig.ReadProjectConfig(projectDir); err == nil && cfg != nil {
				custom = cfg.Signals
			}
			subs := map[string][]string{}
			if prims, _, err := primitive.Walk(projectDir, kconfig.DefaultAgentsRoot); err == nil {
				for _, p := range prims {
					ev := strings.TrimSpace(p.Event)
					if primitive.IsSignal(ev) {
						subs[ev] = append(subs[ev], p.Kind+"/"+p.ID)
					}
				}
			}
			body, _ := json.MarshalIndent(map[string]any{
				"signals":     charter.Signals(custom),
				"subscribers": subs,
				"host_phases": primitive.HostPhases,
			}, "", "  ")
			return mcp.NewToolResultText(string(body)), nil
		},
	)
}

// registerQueryTool answers "what context applies here?" through the
// same charter.Query composition the CLI uses — everywhere rules,
// path-matched rules and corpus, cited references, task-scored entries.
// Descriptors only; the agent opens bodies via keystone_get_primitive.
func registerQueryTool(s *server.MCPServer, projectDir string) {
	s.AddTool(
		mcp.NewTool("keystone_charter_query",
			mcp.WithDescription("What context applies? Compose the relevant charter for a region (file paths) and/or a task (free text): rules without paths: match everywhere; rules and corpus match by paths:; cited references come along; other entries score by keyword overlap on id/description/tags. Returns descriptors with the reason each matched — open bodies with keystone_get_primitive."),
			mcp.WithArray("files",
				mcp.Description("Repo-relative file paths of the region (the files you are about to touch)."),
				mcp.WithStringItems(),
			),
			mcp.WithString("task",
				mcp.Description("What you are doing, free text. Scores id/description/tags by token overlap."),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			files := req.GetStringSlice("files", nil)
			task := req.GetString("task", "")
			if len(files) == 0 && task == "" {
				return mcp.NewToolResultError("pass files, a task, or both"), nil
			}
			prims, _, err := primitive.Walk(projectDir, kconfig.DefaultAgentsRoot)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			composed, _ := primitive.Compose(prims)
			res := charter.Query(composed, files, task)
			type entry struct {
				Kind, ID, Description, Path, Provenance, Why string
			}
			out := make([]entry, 0, len(res.Entries))
			for _, e := range res.Entries {
				out = append(out, entry{e.Kind, e.ID, e.Description, e.Path, e.Provenance, e.Why})
			}
			body, _ := json.MarshalIndent(out, "", "  ")
			return mcp.NewToolResultText(string(body)), nil
		},
	)
}

// registerCoverageTool reports which project files a guide governs and
// which are uncharted (matched by no guide's globs).
func registerCoverageTool(s *server.MCPServer, projectDir string) {
	s.AddTool(
		mcp.NewTool("keystone_charter_coverage",
			mcp.WithDescription("Report charter coverage: how many project source files a rule governs vs. uncharted (matched by no rule's paths:), grouped by top-level region. Surfaces where the agent runs with no ambient rule."),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			res, err := charter.Coverage(projectDir, kconfig.DefaultAgentsRoot)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			body, _ := json.MarshalIndent(map[string]any{
				"total":     res.Total,
				"governed":  res.Governed,
				"uncharted": len(res.Uncharted),
				"regions":   res.UnchartedByRegion(),
			}, "", "  ")
			return mcp.NewToolResultText(string(body)), nil
		},
	)
}

// registerCharterMover exposes move / promote / demote over MCP,
// shelling out to the keystone binary (the write-tools pattern): the
// mover edits files and rewrites citations, so it runs in-process with
// the CLI, never against a stale index.
func registerCharterMover(s *server.MCPServer, projectDir string) {
	s.AddTool(
		mcp.NewTool("keystone_charter_move",
			mcp.WithDescription("Relocate a charter primitive: move it to another kind (optionally renaming inside that kind), or use keystone_charter_promote / keystone_charter_demote to step the corpus \u2192 reference \u2192 rule ladder. Rewrites citations that point at the moved id and re-indexes."),
			mcp.WithString("kind_id",
				mcp.Required(),
				mcp.Description("The primitive, as kind/id (e.g. corpus/billing/props)."),
			),
			mcp.WithString("target",
				mcp.Required(),
				mcp.Description("Where it goes: new-kind, or new-kind/new-path (e.g. rule/billing/billing-gate)."),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			kindID, err := req.RequireString("kind_id")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			target, err := req.RequireString("target")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out, err := execKeystone(ctx, "charter", "move", kindID, target, "--dir", projectDir)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("move failed: %v\n%s", err, out)), nil
			}
			return mcp.NewToolResultText(out), nil
		},
	)
	for verb, dir := range map[string]string{"promote": "+1", "demote": "-1"} {
		verb, dir := verb, dir
		s.AddTool(
			mcp.NewTool("keystone_charter_"+verb,
				mcp.WithDescription(verb+": one step "+map[string]string{"promote": "up", "demote": "down"}[dir]+" the corpus \u2192 reference \u2192 rule ladder; rewrites citations and re-indexes."),
				mcp.WithString("kind_id", mcp.Required(), mcp.Description("The primitive, as kind/id.")),
			),
			func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				kindID, err := req.RequireString("kind_id")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				out, err := execKeystone(ctx, "charter", verb, kindID, "--dir", projectDir)
				if err != nil {
					return mcp.NewToolResultError(fmt.Sprintf("%s failed: %v\n%s", verb, err, out)), nil
				}
				return mcp.NewToolResultText(out), nil
			},
		)
	}
}
