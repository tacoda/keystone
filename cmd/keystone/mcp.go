package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/mcp"
)

// mcpCmd is the parent of `keystone mcp <sub>` — exposes the MCP
// server itself plus utilities to register it with a host agent.
func mcpCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "mcp",
		Short: "Run the keystone MCP server, or register it with a host agent",
		Long: `keystone mcp — model-context-protocol server for the charter.

The server reads .charter/ and .charter/INDEX.json and
exposes them as MCP tools and resources so host agents (Claude Code,
Cursor, Codex, …) can consult the charter at runtime without reading
every markdown file. Same source of truth as the CLI; no version skew.

Subcommands:
  serve     Run the MCP server over stdio (host-launched).
  install   Write a host-agent config blob so the agent auto-launches the server.
  status    Print server health (project root, primitive count, server version).
  show      Print a JSON snippet you can paste into the host agent's config.
`,
	}
	c.AddCommand(mcpServeCmd())
	c.AddCommand(mcpInstallCmd())
	c.AddCommand(mcpStatusCmd())
	c.AddCommand(mcpShowCmd())
	return c
}

func mcpServeCmd() *cobra.Command {
	var dir string
	c := &cobra.Command{
		Use:   "serve",
		Short: "Run the MCP server over stdio (host-launched)",
		Long: `Run the MCP server over stdio. Intended to be invoked by a host
agent (Claude Code, Cursor, etc.); the host writes the launch command
into its MCP config. To register the server, run:

  keystone mcp install --agent claude-code
  keystone mcp install --agent cursor

The server reads from the project's .charter/ tree. Pass
--dir to point at a project other than the current directory.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer cancel()
			return mcp.Serve(ctx, mcp.Options{ProjectDir: dir})
		},
	}
	c.Flags().StringVar(&dir, "dir", "", "Project root (defaults to cwd).")
	return c
}

func mcpInstallCmd() *cobra.Command {
	var (
		agent    string
		dir      string
		stdout   bool
		serverNm string
	)
	c := &cobra.Command{
		Use:   "install",
		Short: "Write a host-agent MCP config so the agent auto-launches the server",
		Long: `Write the appropriate .mcp.json (or per-host equivalent) so the
agent auto-launches the keystone MCP server on session start.

Supported agents:
  claude-code   → <dir>/.mcp.json
  cursor        → <dir>/.cursor/mcp.json
  vscode        → <dir>/.vscode/mcp.json
  codex         → <dir>/.codex/mcp.json
  opencode      → <dir>/opencode.json (mcp key, type: local)

Pass --stdout to print the config JSON instead of writing it; useful
for piping into another tool or pasting into an agent's session.

The server entry runs ` + "`keystone mcp serve`" + ` against the project's
.charter/. Re-installing is idempotent.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if agent == "" {
				return fmt.Errorf("--agent is required (try: claude-code, cursor, vscode, codex, opencode)")
			}
			absDir, err := filepath.Abs(dirOr(dir, "."))
			if err != nil {
				return err
			}
			entry := mcpServerEntry(serverNm, absDir)
			path, body, err := agentMCPConfigPath(agent, absDir, entry)
			if err != nil {
				return err
			}
			if stdout {
				fmt.Println(string(body))
				return nil
			}
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(path, body, 0o644); err != nil {
				return err
			}
			rel, _ := filepath.Rel(absDir, path)
			fmt.Fprintf(os.Stdout, "✓ wrote %s — restart the agent to pick up the keystone MCP server\n", rel)
			return nil
		},
	}
	c.Flags().StringVar(&agent, "agent", "", "Target agent (claude-code | cursor | vscode | codex | opencode).")
	c.Flags().StringVar(&dir, "dir", ".", "Project root (defaults to cwd).")
	c.Flags().BoolVar(&stdout, "stdout", false, "Print the config JSON to stdout instead of writing the file.")
	c.Flags().StringVar(&serverNm, "name", "keystone", "Server name to record in the agent's config.")
	return c
}

func mcpStatusCmd() *cobra.Command {
	var dir string
	c := &cobra.Command{
		Use:   "status",
		Short: "Print MCP server diagnostics for a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			absDir, err := filepath.Abs(dirOr(dir, "."))
			if err != nil {
				return err
			}
			indexPath := filepath.Join(absDir, config.DefaultCharterRoot, config.IndexName)
			indexExists := fileExists(indexPath)
			fmt.Fprintf(os.Stdout, "keystone mcp status\n")
			fmt.Fprintf(os.Stdout, "  project_dir:  %s\n", absDir)
			fmt.Fprintf(os.Stdout, "  index:        %s (%s)\n", indexPath, presentStr(indexExists))
			fmt.Fprintf(os.Stdout, "  server:       %s (version %s)\n", "keystone-mcp", mcp.Version)
			bin, _ := exec.LookPath("keystone")
			if bin == "" {
				bin = "(keystone binary not on PATH — host agents won't find it)"
			}
			fmt.Fprintf(os.Stdout, "  launch:       %s mcp serve --dir %s\n", bin, absDir)
			return nil
		},
	}
	c.Flags().StringVar(&dir, "dir", ".", "Project root (defaults to cwd).")
	return c
}

func mcpShowCmd() *cobra.Command {
	var (
		dir      string
		serverNm string
	)
	c := &cobra.Command{
		Use:   "show",
		Short: "Print a generic MCP config snippet (host-agnostic)",
		Long: `Prints a minimal { "mcpServers": { ... } } block you can paste
into a host agent's config or chat session. The structure matches
Claude Code's .mcp.json and Cursor's .cursor/mcp.json; other agents
may use a slightly different envelope.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			absDir, err := filepath.Abs(dirOr(dir, "."))
			if err != nil {
				return err
			}
			entry := mcpServerEntry(serverNm, absDir)
			body, _ := json.MarshalIndent(map[string]any{
				"mcpServers": map[string]any{serverNm: entry},
			}, "", "  ")
			fmt.Println(string(body))
			return nil
		},
	}
	c.Flags().StringVar(&dir, "dir", ".", "Project root (defaults to cwd).")
	c.Flags().StringVar(&serverNm, "name", "keystone", "Server name to record in the snippet.")
	return c
}

// mcpServerEntry returns the {command, args} record host agents expect
// inside their mcpServers map.
func mcpServerEntry(name, projectDir string) map[string]any {
	return map[string]any{
		"command": "keystone",
		"args":    []string{"mcp", "serve", "--dir", projectDir},
	}
}

// agentMCPConfigPath returns where to write the host agent's config
// file and the body to write. The body is a complete JSON document
// (existing entries are preserved when present — re-installing
// `keystone` does not erase another server's registration).
func agentMCPConfigPath(agent, projectDir string, entry map[string]any) (string, []byte, error) {
	var rel string
	switch strings.ToLower(agent) {
	case "claude-code", "claude":
		rel = ".mcp.json"
	case "cursor":
		rel = filepath.Join(".cursor", "mcp.json")
	case "vscode", "vs-code":
		rel = filepath.Join(".vscode", "mcp.json")
	case "codex":
		rel = filepath.Join(".codex", "mcp.json")
	case "opencode":
		return opencodeMCPConfig(projectDir)
	default:
		return "", nil, fmt.Errorf("unsupported --agent %q (try: claude-code | cursor | vscode | codex | opencode)", agent)
	}
	path := filepath.Join(projectDir, rel)

	doc := map[string]any{"mcpServers": map[string]any{}}
	if raw, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(raw, &doc) // best-effort; overwrite on parse failure
		if _, ok := doc["mcpServers"]; !ok {
			doc["mcpServers"] = map[string]any{}
		}
	}
	servers, ok := doc["mcpServers"].(map[string]any)
	if !ok {
		servers = map[string]any{}
		doc["mcpServers"] = servers
	}
	servers["keystone"] = entry

	body, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", nil, err
	}
	body = append(body, '\n')
	return path, body, nil
}

// opencodeMCPConfig writes the keystone server into opencode.json's `mcp`
// key as a `local` (stdio) server. opencode's schema differs from the
// Claude Code envelope: a top-level `mcp` object (not `mcpServers`), and
// each server records `type: "local"`, a `command` array (binary + args
// in one list), and `enabled`. Other top-level keys (`$schema`, `model`,
// existing `mcp` entries) are preserved.
func opencodeMCPConfig(projectDir string) (string, []byte, error) {
	path := filepath.Join(projectDir, "opencode.json")

	doc := map[string]any{}
	if raw, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(raw, &doc) // best-effort; overwrite on parse failure
	}
	if _, ok := doc["$schema"]; !ok {
		doc["$schema"] = "https://opencode.ai/config.json"
	}
	servers, ok := doc["mcp"].(map[string]any)
	if !ok {
		servers = map[string]any{}
		doc["mcp"] = servers
	}
	servers["keystone"] = map[string]any{
		"type":    "local",
		"command": []string{"keystone", "mcp", "serve", "--dir", projectDir},
		"enabled": true,
	}

	body, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", nil, err
	}
	body = append(body, '\n')
	return path, body, nil
}

func dirOr(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func presentStr(b bool) string {
	if b {
		return "present"
	}
	return "missing — run `keystone index`"
}
