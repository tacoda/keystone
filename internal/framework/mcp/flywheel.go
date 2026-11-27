package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerFlywheelTools exposes the charter-hygiene flywheels over MCP:
// learn (capture), synthesize (promote), audit (report), mode (pacing).
// Each shells out to the keystone binary (the write-tools pattern), so
// CLI and MCP run the exact same code path.
func registerFlywheelTools(s *server.MCPServer, projectDir string) {
	s.AddTool(
		mcp.NewTool("keystone_learn",
			mcp.WithDescription("Capture a learning candidate into .keystone/learning/inbox/ (the Learning flywheel's additive step). A later keystone_synthesize promotes accepted candidates. `body` is what happened; `layer` optionally proposes the destination (references/principles | rules/<id> | hook); `globs` is a comma-separated list proposed as the promoted rule's paths:."),
			mcp.WithString("slug", mcp.Required(), mcp.Description("Short slug for the candidate file, e.g. no-silent-overwrites.")),
			mcp.WithString("body", mcp.Required(), mcp.Description("What happened, in a few lines.")),
			mcp.WithString("source", mcp.Description("Where the candidate came from (review, incident, ...). Defaults to keystone_learn.")),
			mcp.WithString("layer", mcp.Description("Proposed destination layer: references/principles, rules/<id>, or hook.")),
			mcp.WithString("globs", mcp.Description("Comma-separated globs proposed as the promoted rule's paths:.")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			slug, err := req.RequireString("slug")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			body, err := req.RequireString("body")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			args := learnArgs(projectDir, slug, body,
				req.GetString("source", ""), req.GetString("layer", ""), req.GetString("globs", ""))
			out, err := execKeystone(ctx, args...)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("keystone_learn failed: %v\n%s", err, out)), nil
			}
			return mcp.NewToolResultText(out), nil
		},
	)
	s.AddTool(
		mcp.NewTool("keystone_synthesize",
			mcp.WithDescription("Promote accepted learning-inbox candidates to their proposed layer (rule / reference / hook), file them under learning/promoted/, and re-index. Only candidates marked status: accepted are touched. Set dry_run to preview without writing."),
			mcp.WithBoolean("dry_run", mcp.Description("Report what would be promoted without writing anything.")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			out, err := execKeystone(ctx, synthesizeArgs(projectDir, req.GetBool("dry_run", false))...)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("keystone_synthesize failed: %v\n%s", err, out)), nil
			}
			return mcp.NewToolResultText(out), nil
		},
	)
	s.AddTool(
		mcp.NewTool("keystone_audit",
			mcp.WithDescription("Dual-flywheel audit report as JSON: the learning inbox (pending/accepted candidates) and pruning candidates (project primitives that are uncited, unscoped, and untriggered). Data only — acting on a candidate is keystone_charter_move / keystone_charter_demote or an edit."),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			out, err := execKeystone(ctx, auditArgs(projectDir)...)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("keystone_audit failed: %v\n%s", err, out)), nil
			}
			return mcp.NewToolResultText(out), nil
		},
	)
	s.AddTool(
		mcp.NewTool("keystone_drift_report",
			mcp.WithDescription("Context-drift report as JSON: dead relative links, rules whose paths: match zero files, and primitives whose verified: date is older than the drift.days window (default 90). Warnings only — keystone names stale context, it does not judge the diff. `files` optionally scopes the report to comma-separated repo-relative files."),
			mcp.WithString("files", mcp.Description("Comma-separated repo-relative files to scope the scan to. Omit for the whole charter.")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			out, err := execKeystone(ctx, driftArgs(projectDir, req.GetString("files", ""))...)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("keystone_drift_report failed: %v\n%s", err, out)), nil
			}
			return mcp.NewToolResultText(out), nil
		},
	)
	s.AddTool(
		mcp.NewTool("keystone_mode",
			mcp.WithDescription("Read or set the charter pacing mode. Omit `mode` to read the active mode; pass paired, solo, or autopilot to switch it (rewrites the **Active mode:** line of rules/process/modes.md and re-indexes). Returns JSON."),
			mcp.WithString("mode", mcp.Description("paired | solo | autopilot. Omit to read the current mode.")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			out, err := execKeystone(ctx, modeArgs(projectDir, req.GetString("mode", ""))...)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("keystone_mode failed: %v\n%s", err, out)), nil
			}
			return mcp.NewToolResultText(out), nil
		},
	)
}

// The arg builders are pure so tests can assert the exact CLI argv each
// tool execs — MCP must never diverge from the CLI's flags.

func learnArgs(projectDir, slug, body, source, layer, globs string) []string {
	args := []string{"learn", slug, "--body", body, "--dir", projectDir}
	if source != "" {
		args = append(args, "--source", source)
	}
	if layer != "" {
		args = append(args, "--layer", layer)
	}
	if globs != "" {
		args = append(args, "--globs")
		for _, g := range strings.Split(globs, ",") {
			if g = strings.TrimSpace(g); g != "" {
				args = append(args, g)
			}
		}
	}
	return args
}

func synthesizeArgs(projectDir string, dryRun bool) []string {
	args := []string{"synthesize", "--dir", projectDir, "--json"}
	if dryRun {
		args = append(args, "--dry-run")
	}
	return args
}

func auditArgs(projectDir string) []string {
	return []string{"audit", "--dir", projectDir, "--json"}
}

func modeArgs(projectDir, mode string) []string {
	args := []string{"mode"}
	if mode != "" {
		args = append(args, mode)
	}
	return append(args, "--dir", projectDir, "--json")
}

func driftArgs(projectDir, files string) []string {
	args := []string{"drift", "--dir", projectDir, "--json"}
	if files != "" {
		for _, f := range strings.Split(files, ",") {
			if f = strings.TrimSpace(f); f != "" {
				args = append(args, f)
			}
		}
	}
	return args
}
