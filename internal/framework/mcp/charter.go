package mcp

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/tacoda/keystone/internal/framework/charter"
	kconfig "github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// registerCharterViews exposes the read-side charter views — signals and
// coverage — over MCP, so a host agent sees the same concepts the CLI and
// dashboard do.
func registerCharterViews(s *server.MCPServer, projectDir string) {
	registerSignalListTool(s, projectDir)
	registerCoverageTool(s, projectDir)
	registerExplainTool(s, projectDir)
}

// registerExplainTool explains a primitive — how it activates, what it
// links to, where it projects — so an agent can understand a skill,
// command, sensor, guide, etc. without reading its whole body.
func registerExplainTool(s *server.MCPServer, projectDir string) {
	s.AddTool(
		mcp.NewTool("keystone_explain",
			mcp.WithDescription("Explain a primitive by id: its kind, how/when it activates (a guide's globs, a sensor's on:, a skill's triggers, …), what it links to, and where it projects. Optionally narrow by kind."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Primitive id (e.g. 'keystone:verify', 'guides/idioms/go/stdlib-first').")),
			mcp.WithString("kind", mcp.Description("Narrow to this kind when an id is shared across kinds.")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id, err := req.RequireString("id")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			kind := req.GetString("kind", "")
			prims, _, err := primitive.Walk(projectDir, kconfig.DefaultCharterRoot)
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
			if prims, _, err := primitive.Walk(projectDir, kconfig.DefaultCharterRoot); err == nil {
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

// registerCoverageTool reports which project files a guide governs and
// which are uncharted (matched by no guide's globs).
func registerCoverageTool(s *server.MCPServer, projectDir string) {
	s.AddTool(
		mcp.NewTool("keystone_charter_coverage",
			mcp.WithDescription("Report charter coverage: how many project source files a guide governs vs. uncharted (matched by no guide's globs), grouped by top-level region. Surfaces where the agent runs with no ambient rule."),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			res, err := charter.Coverage(projectDir, kconfig.DefaultCharterRoot)
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
