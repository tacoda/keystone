package mcp

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	kconfig "github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// registerSearchTool exposes the harness-wide substring search as an
// MCP tool. Agents call this when "where is the convention about X?"
// is the actual question.
func registerSearchTool(s *server.MCPServer, projectDir string) {
	s.AddTool(
		mcp.NewTool("keystone_search",
			mcp.WithDescription("Search every primitive (id, description, globs, traces, body) for a substring. Returns ranked hits w/ excerpt + match location. Use this before falling through to corpus or external sources — often the answer is already in the harness."),
			mcp.WithString("query",
				mcp.Required(),
				mcp.Description("Free-form search string. Case-insensitive substring match."),
			),
			mcp.WithNumber("limit",
				mcp.Description("Max hits (default 25, 0 = all)."),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			q, err := req.RequireString("query")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			limit := int(req.GetFloat("limit", 25))
			primitives, _, err := primitive.Walk(projectDir, kconfig.DefaultHarnessRoot)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			hits := primitive.Search(projectDir, primitives, q, limit)
			body, _ := json.MarshalIndent(map[string]any{
				"query": q,
				"count": len(hits),
				"hits":  hits,
			}, "", "  ")
			return mcp.NewToolResultText(string(body)), nil
		},
	)
}
