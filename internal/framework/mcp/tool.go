package mcp

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

// registerToolPrimitives registers every `kind: tool` with `transport: mcp`
// as an MCP tool on keystone's own server. The handler shells out to the
// tool's `run:` script, passing each declared arg as KEYSTONE_ARG_<NAME>.
// cli/plugin tools reach the agent through their own transports, not here.
//
// Read at startup from the walked harness; a missing harness is a silent
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

// toolMCPName renders a tool id as an MCP tool name: keystone_tool_<id>, with
// namespace / hierarchy separators flattened to underscores.
func toolMCPName(id string) string {
	return "keystone_tool_" + strings.NewReplacer(":", "_", "/", "_", "-", "_").Replace(id)
}

// execTool runs the tool's run: handler with the call arguments exported as
// KEYSTONE_ARG_<NAME> env vars.
func execTool(ctx context.Context, projectDir string, p primitive.Primitive, req mcp.CallToolRequest) (string, error) {
	cmd := exec.CommandContext(ctx, "bash", "-c", p.Run)
	cmd.Dir = projectDir
	env := os.Environ()
	for _, a := range p.Args {
		env = append(env, "KEYSTONE_ARG_"+strings.ToUpper(a.Name)+"="+req.GetString(a.Name, ""))
	}
	cmd.Env = env
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	return buf.String(), err
}
