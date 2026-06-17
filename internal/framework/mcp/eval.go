package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/tacoda/keystone/internal/framework/eval"
)

// registerEvalTools wires the MCP surface for harness evals.
//
// Tools:
//   keystone_eval_list             List every eval id in the harness.
//   keystone_eval_run [filter]     Run all (or filtered) evals; return a JSON report.
//   keystone_eval_report           Run + render markdown (human consumption).
func registerEvalTools(s *server.MCPServer, projectDir string) {
	s.AddTool(
		mcp.NewTool("keystone_eval_list",
			mcp.WithDescription("List every eval id declared in .keystone/harness/evals/. Returns descriptors (id, level, description) without running anything."),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			specs, err := eval.LoadAll(projectDir)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out := make([]map[string]any, 0, len(specs))
			for _, sp := range specs {
				out = append(out, map[string]any{
					"id":          sp.ID,
					"level":       sp.Level,
					"levels":      sp.Levels,
					"description": sp.Description,
				})
			}
			body, _ := json.MarshalIndent(map[string]any{
				"count": len(out),
				"evals": out,
			}, "", "  ")
			return mcp.NewToolResultText(string(body)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("keystone_eval_run",
			mcp.WithDescription("Run every eval and return a JSON report. Pass `filter` to narrow by substring on the eval id. Static + sensor levels are implemented; agent level is reserved for 2.1."),
			mcp.WithString("filter",
				mcp.Description("Substring filter on eval ids. Empty = all."),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			filter := req.GetString("filter", "")
			specs, err := eval.LoadAll(projectDir)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			rep := eval.Run(ctx, projectDir, specs, filter)
			body, _ := json.MarshalIndent(rep, "", "  ")
			return mcp.NewToolResultText(string(body)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("keystone_eval_report",
			mcp.WithDescription("Run evals and return a human-readable markdown report. Same engine as keystone_eval_run; different format."),
			mcp.WithString("filter",
				mcp.Description("Substring filter on eval ids. Empty = all."),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			filter := req.GetString("filter", "")
			specs, err := eval.LoadAll(projectDir)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			rep := eval.Run(ctx, projectDir, specs, filter)
			return mcp.NewToolResultText(eval.RenderMarkdown(rep)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("keystone_eval_baseline",
			mcp.WithDescription("Run evals against the current harness AND a git ref's worktree; return regressions/fixes/new/removed. Use this in PR reviews to confirm a harness change improved (or at minimum didn't regress) the measured behavior."),
			mcp.WithString("ref",
				mcp.Required(),
				mcp.Description("Git ref to compare against (e.g. `main`, a tag, a commit SHA)."),
			),
			mcp.WithString("filter",
				mcp.Description("Substring filter on eval ids. Empty = all."),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			ref, err := req.RequireString("ref")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			filter := req.GetString("filter", "")
			diff, err := eval.RunWithBaseline(ctx, projectDir, ref, filter)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			body, _ := json.MarshalIndent(diff, "", "  ")
			return mcp.NewToolResultText(string(body)), nil
		},
	)

	_ = fmt.Sprintf
}
