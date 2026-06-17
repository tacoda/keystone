package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerWriteTools adds the mutate-the-harness tools: every `new`
// generator, plus `harness_bootstrap` (runs init) and `target_add`.
//
// Each tool shells out to the keystone binary itself so the MCP layer
// reuses the CLI's argument parsing + flag handling unchanged. Same
// behavior either way — no duplication of authoring logic.
//
// The running binary path comes from os.Executable() so the spawned
// child is the *same* keystone the host launched (not whatever
// happens to be first on PATH).
func registerWriteTools(s *server.MCPServer, projectDir string) {
	// Generator tools (one per kind). Each takes a string `id` and
	// optionally `kind` (for sensors). The handler synthesizes the
	// equivalent CLI arg slice and execs the keystone binary.
	type gen struct {
		tool, verb, idArg, desc string
		extras                  []generatorFlag
	}

	gens := []gen{
		{
			tool:  "keystone_new_guide",
			verb:  "guide",
			idArg: "<topic>/<name>",
			desc:  "Scaffold a framework guide + paired corpus stub. id form: '<topic>/<name>' (e.g. 'process/release').",
		},
		{
			tool:  "keystone_new_corpus",
			verb:  "corpus",
			idArg: "<topic>/<name>",
			desc:  "Scaffold a corpus reasoning entry. id form: '<topic>/<name>'.",
		},
		{
			tool:  "keystone_new_sensor",
			verb:  "sensor",
			idArg: "<name>",
			desc:  "Scaffold a sensor (automated check). id form: '<name>'.",
			extras: []generatorFlag{
				{name: "kind", flag: "--kind", desc: "Sensor kind: computational | inferential. Default: computational."},
			},
		},
		{
			tool:  "keystone_new_action",
			verb:  "action",
			idArg: "<name>",
			desc:  "Scaffold a lifecycle action.",
		},
		{
			tool:  "keystone_new_playbook",
			verb:  "playbook",
			idArg: "<name>",
			desc:  "Scaffold a playbook (ordered sequence of actions).",
		},
		{
			tool:  "keystone_new_rule",
			verb:  "rule",
			idArg: "<id>",
			desc:  "Scaffold an agent-side rule (host-native flavor; framework default is `guide`).",
		},
		{
			tool:  "keystone_new_skill",
			verb:  "skill",
			idArg: "<id>",
			desc:  "Scaffold a Claude Code skill. id may use the colon-namespaced form (e.g. 'keystone:demo').",
		},
		{
			tool:  "keystone_new_subagent",
			verb:  "subagent",
			idArg: "<id>",
			desc:  "Scaffold a subagent definition.",
		},
		{
			tool:  "keystone_new_command",
			verb:  "command",
			idArg: "<id>",
			desc:  "Scaffold a slash command.",
		},
		{
			tool:  "keystone_new_adapter",
			verb:  "adapter",
			idArg: "<agent>",
			desc:  "Scaffold the per-agent adapter triple (activation, lifecycle, sensors) for a new host.",
		},
		{
			tool:  "keystone_new_policy",
			verb:  "policy",
			idArg: "<name>",
			desc:  "Scaffold a new policy repo skeleton (separate dir; publish to git afterward).",
		},
	}

	for _, g := range gens {
		registerGenerator(s, projectDir, g.tool, g.verb, g.idArg, g.desc, g.extras)
	}

	s.AddTool(
		mcp.NewTool("keystone_harness_bootstrap",
			mcp.WithDescription("Scaffold the harness into the project directory (non-interactive equivalent of `keystone init`). Use this when the project has no `.keystone/` yet."),
			mcp.WithString("agent",
				mcp.Required(),
				mcp.Description("Target agent to bind (e.g. claude-code, codex, cursor, _generic)."),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			agent, err := req.RequireString("agent")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out, err := execKeystone(ctx, "init", projectDir, "--agent", agent)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("bootstrap failed: %v\n%s", err, out)), nil
			}
			return mcp.NewToolResultText(out), nil
		},
	)

	s.AddTool(
		mcp.NewTool("keystone_target_add",
			mcp.WithDescription("Add another agent target to an existing harness. Equivalent to `keystone target add <agent>`."),
			mcp.WithString("agent",
				mcp.Required(),
				mcp.Description("Agent target to install (e.g. claude-code, codex, cursor)."),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			agent, err := req.RequireString("agent")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out, err := execKeystone(ctx, "target", "add", agent, "--dir", projectDir)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("target add failed: %v\n%s", err, out)), nil
			}
			return mcp.NewToolResultText(out), nil
		},
	)

	// Two maintenance tools — index + project. Cheap, idempotent. The
	// MCP agent should call these after any keystone_new_* tool so the
	// INDEX.json and host projections stay current.
	s.AddTool(
		mcp.NewTool("keystone_index_refresh",
			mcp.WithDescription("Regenerate .keystone/INDEX.json from the harness. Call after any keystone_new_* tool."),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			out, err := execKeystone(ctx, "index", "--dir", projectDir)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("index refresh failed: %v\n%s", err, out)), nil
			}
			return mcp.NewToolResultText(out), nil
		},
	)

	s.AddTool(
		mcp.NewTool("keystone_project_refresh",
			mcp.WithDescription("Regenerate host-native projections (.claude/...) from canonical sources. Call after any keystone_new_* tool that touches skill/subagent/command."),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			out, err := execKeystone(ctx, "project", "--dir", projectDir)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("project refresh failed: %v\n%s", err, out)), nil
			}
			return mcp.NewToolResultText(out), nil
		},
	)
}

// generatorFlag describes an optional named flag for a new-* tool that
// gets passed through to the CLI's flag parser when present.
type generatorFlag struct {
	name, flag, desc string
}

// registerGenerator wires one `keystone new <verb> <id>` tool into the
// MCP server. The handler runs the keystone binary as a child process,
// captures combined output, and returns it as the tool result.
func registerGenerator(
	s *server.MCPServer,
	projectDir, toolName, verb, idArg, desc string,
	extras []generatorFlag,
) {
	opts := []mcp.ToolOption{
		mcp.WithDescription(desc),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Primitive id, form: "+idArg+"."),
		),
	}
	for _, e := range extras {
		opts = append(opts, mcp.WithString(e.name, mcp.Description(e.desc)))
	}
	s.AddTool(
		mcp.NewTool(toolName, opts...),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id, err := req.RequireString("id")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			args := []string{"new", verb, "--dir", projectDir, id}
			for _, e := range extras {
				if val := req.GetString(e.name, ""); val != "" {
					args = append(args, e.flag, val)
				}
			}
			out, err := execKeystone(ctx, args...)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("%s failed: %v\n%s", toolName, err, out)), nil
			}
			body, _ := json.MarshalIndent(map[string]any{
				"tool":     toolName,
				"verb":     verb,
				"id":       id,
				"output":   strings.TrimSpace(out),
				"project":  projectDir,
				"refresh_hint": "Call keystone_index_refresh next to rebuild .keystone/INDEX.json.",
			}, "", "  ")
			return mcp.NewToolResultText(string(body)), nil
		},
	)
}

// execKeystone runs the running keystone binary as a child process
// with the given args. Uses os.Executable() so the spawned child is
// the same binary the host launched.
func execKeystone(ctx context.Context, args ...string) (string, error) {
	self, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locate self: %w", err)
	}
	cmd := exec.CommandContext(ctx, self, args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err = cmd.Run()
	return buf.String(), err
}
