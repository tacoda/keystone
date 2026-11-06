package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/tacoda/keystone/internal/framework/primitive"
	"github.com/tacoda/keystone/internal/framework/projection"
	"github.com/tacoda/keystone/internal/framework/toolrun"
)

// registerToolPrimitives registers every `kind: tool` with `transport: mcp`
// as an MCP tool on keystone's own server. The handler shells out to the
// tool's `run:` script, passing each declared arg as KEYSTONE_ARG_<NAME>.
// cli/plugin tools reach the agent through their own transports, not here.
//
// Read at startup from the walked charter; a missing charter is a silent
// no-op (the server still serves its built-in tools).
func registerToolPrimitives(s *server.MCPServer, projectDir string) {
	idx, err := loadIndex(projectDir)
	if err != nil {
		return
	}
	for _, p := range idx.Primitives {
		if primitive.Kind(p.Kind) == primitive.KindTool && p.Transport == "mcp" {
			registerOneTool(s, projectDir, p)
		}
	}
}

// registerOneTool wires a single tool primitive onto the MCP server.
func registerOneTool(s *server.MCPServer, projectDir string, p primitive.Primitive) {
	opts := []mcp.ToolOption{mcp.WithDescription(p.Description)}
	for _, a := range p.Args {
		opts = append(opts, mcp.WithString(a.Name, mcp.Description(a.Description)))
	}
	s.AddTool(
		mcp.NewTool(toolMCPName(p.ID), opts...),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			out, err := execTool(ctx, projectDir, p, req)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("tool %s failed: %v\n%s", p.ID, err, out)), nil
			}
			return mcp.NewToolResultText(out), nil
		},
	)
}

// toolMCPName renders a tool id as an MCP tool name (keystone_tool_<id>);
// AGENTS.md lists the same name.
func toolMCPName(id string) string { return projection.MCPToolName(id) }

// execTool runs the tool's run: handler (toolrun.Run) with the declared
// arguments the call supplied.
func execTool(ctx context.Context, projectDir string, p primitive.Primitive, req mcp.CallToolRequest) (string, error) {
	args := map[string]string{}
	for _, a := range p.Args {
		if v := req.GetString(a.Name, ""); v != "" {
			args[a.Name] = v
		}
	}
	return toolrun.Run(ctx, projectDir, p, args)
}
