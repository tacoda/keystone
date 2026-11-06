package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tacoda/keystone/internal/framework/primitive"
	"github.com/tacoda/keystone/internal/framework/projection"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/loader"
	"github.com/tacoda/keystone/internal/framework/lockfile"
	"github.com/tacoda/keystone/internal/framework/policies"
	"github.com/tacoda/keystone/internal/framework/sensors"
)

// runVerify handles `keystone verify [--dir <path>] [--sensor <id>]`.
//
// Default mode: walks the policy tree from keystone.json, checks each
// vendored policy for drift against the lockfile, and reports any
// strict-cascade violations from project files shadowing locked items.
// Exits non-zero on any violation; drift alone exits zero but is
// reported and reset.
//
// --sensor <id> mode: bypasses the policy cascade entirely and runs
// one keystone-owned sensor. The sensor reads Claude Code's hook
// protocol JSON from stdin when present (tool_input.file_path,
// tool_input.content, command). Exit codes follow the hook protocol:
//
//	0  pass / advisory
//	2  block with the sensor's Message rendered to stderr
//	1  internal error
func runVerify(args []string) error {
	dir, sensorID, help, err := parseVerifyArgs(args)
	if err != nil || help {
		return err
	}

	if sensorID != "" {
		return runOneSensor(dir, sensorID)
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}
	cfg, err := config.ReadProjectConfig(absDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no %s at %s — run `keystone init` first", config.ProjectConfigFile, absDir)
		}
		return err
	}

	lf, err := lockfile.Read(absDir)
	if err != nil {
		return err
	}
	expected := map[string]map[string]string{}
	for name, lock := range lf.Policies {
		expected[name] = lock.Files
	}

	res, err := loader.Verify(absDir, cfg, expected)
	if err != nil {
		return err
	}

	if res.HasDrift() {
		fmt.Fprintf(os.Stdout, "▸ drift detected — resetting %d policy(s)\n", len(res.Drift))
		for _, d := range res.Drift {
			fmt.Fprintf(os.Stdout, "  • %s: %d drifted file(s)\n", d.Policy, len(d.Files))
			for _, f := range d.Files {
				fmt.Fprintf(os.Stdout, "      - %s (%s)\n", f.Path, f.Kind)
			}
			if err := policies.Reset(d.Policy, absDir); err != nil {
				return fmt.Errorf("reset %s: %w", d.Policy, err)
			}
		}
		fmt.Fprintln(os.Stdout, "  re-run `keystone install` to repopulate from cache")
	}

	if res.HasErrors() {
		fmt.Fprintf(os.Stdout, "✗ keystone verify found %d strict violation(s) in %s\n\n", len(res.Violations), absDir)
		for _, v := range res.Violations {
			fmt.Fprintln(os.Stdout, "  "+v.String())
			fmt.Fprintln(os.Stdout)
		}
		fmt.Fprintln(os.Stdout, "Strict policy items cannot be overridden by the project layer. Remove the offending file(s) or take it up with the policy author.")
		return fmt.Errorf("strict cascade is violated")
	}

	if err := checkProjectionDrift(absDir); err != nil {
		return err
	}
	if !res.HasDrift() {
		fmt.Fprintf(os.Stdout, "✓ keystone verify clean — no drift, no strict violations (%s)\n", absDir)
	}
	return nil
}

// parseVerifyArgs reads `keystone verify` flags; help=true means usage was printed.
func parseVerifyArgs(args []string) (dir, sensorID string, help bool, err error) {
	dir = "."
	help, err = parseFlags(args, map[string]*string{"--dir": &dir, "--sensor": &sensorID}, nil,
		func() { printVerifyUsage(os.Stdout) })
	return dir, sensorID, help, err
}

// checkProjectionDrift fails when a host projection no longer matches the
// charter: a missing output, a link pointing elsewhere, or a shim or copy
// edited by hand. `keystone project` repairs all three.
func checkProjectionDrift(absDir string) error {
	prims, _, err := primitive.Walk(absDir, config.DefaultAgentsRoot)
	if err != nil {
		return err
	}
	composed, _ := primitive.Compose(prims)
	lf, err := lockfile.Read(absDir)
	if err != nil {
		return err
	}
	pr, mode := planHosts(absDir, composed, lf.ManagedEntries)
	drifts := projection.Check(absDir, pr.Outputs, mode)
	if len(drifts) == 0 {
		return nil
	}
	fmt.Fprintf(os.Stdout, "✗ %d host projection(s) drifted from the charter:\n", len(drifts))
	for _, d := range drifts {
		fmt.Fprintf(os.Stdout, "  • %s: %s\n", d.Dest, d.Problem)
	}
	return fmt.Errorf("projection drift — run `keystone project`")
}

func printVerifyUsage(w *os.File) {
	ids := strings.Join(sensors.IDs(), ", ")
	if ids == "" {
		ids = "(none registered)"
	}
	fmt.Fprintf(w, `keystone verify — check vendored policies, the strict cascade, or one sensor

Usage:
  keystone verify [--dir <path>]
  keystone verify --sensor <id> [--dir <path>]

Default mode: reads keystone.json + the lockfile, then:
  - Walks every vendored policy and compares per-file hashes to the
    lockfile. Any drift triggers an immediate policies.Reset (run
    `+"`keystone install`"+` afterward to repopulate).
  - Walks each policy's strict items and reports project-layer files
    that shadow them.

--sensor mode: runs one keystone-owned sensor and exits with a Claude
Code hook protocol code. Reads tool_input fields (file_path, content,
command) from stdin JSON when present. Sensors that wrap external
tools (build, test, lint, vuln-scan) aren't routed here — wire those
to the external tool directly in your keystone.json hooks block.

Currently registered sensors: %s

Exit codes (default mode):
  0  clean (no drift after reset, no violations)
  0  drift only — drifted policies were reset; user re-installs to recover
  1  any strict violation

Exit codes (--sensor mode):
  0  pass / advisory
  2  block with sensor message on stderr (Claude Code hook block)
  1  internal error or unknown sensor

Flags:
  --dir <path>      Project root (defaults to cwd).
  --sensor <id>     Run one sensor instead of the policy cascade.
`, ids)
}

// runOneSensor dispatches a single sensor by id. Stdin is parsed as
// the Claude Code hook protocol payload (tool_input.file_path,
// tool_input.content, tool_input.command); missing fields degrade to
// no-op input — the sensor decides what to do with empty context.
//
// Exit codes mirror Claude Code's hook protocol so the verify command
// can be wired directly as a hook command in .claude/settings.json.
func runOneSensor(projectDir, sensorID string) error {
	absDir, err := filepath.Abs(projectDir)
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}

	ctx := sensors.Context{ProjectDir: absDir}
	if stdinInfo, err := os.Stdin.Stat(); err == nil && (stdinInfo.Mode()&os.ModeCharDevice) == 0 {
		// Stdin is piped — Claude Code hook protocol payload.
		raw, err := io.ReadAll(os.Stdin)
		if err == nil && len(raw) > 0 {
			parseHookStdin(raw, &ctx)
		}
	}

	res, err := sensors.Run(sensorID, ctx, os.Stdout)
	if err != nil {
		var unknown sensors.ErrUnknownSensor
		if errors.As(err, &unknown) {
			// Unknown sensor is a SOFT skip, not a hook failure.
			// Many sensors declare host_triggers without keystone-owned
			// backing logic — they're meant to wire to external tools
			// directly (go test, gofmt, govulncheck) or are LLM-judgment
			// sensors invoked through an action. The hook should pass
			// quietly so the user isn't spammed with non-blocking warnings.
			fmt.Fprintf(os.Stderr, "keystone verify --sensor %q: no keystone-owned runner (registered: %s). Skipping.\n",
				sensorID, strings.Join(sensors.IDs(), ", "))
			return nil
		}
		return fmt.Errorf("sensor %s: %w", sensorID, err)
	}
	if res.Block {
		if res.Message != "" {
			fmt.Fprintln(os.Stderr, res.Message)
		}
		os.Exit(2)
	}
	return nil
}

// parseHookStdin extracts the fields keystone sensors actually consume
// from Claude Code's hook protocol payload. Unknown fields are
// silently ignored — the protocol is permissive on the host side and
// keystone is permissive on the consumer side.
func parseHookStdin(raw []byte, ctx *sensors.Context) {
	var payload struct {
		ToolInput struct {
			FilePath string `json:"file_path"`
			Content  string `json:"content"`
			Command  string `json:"command"`
		} `json:"tool_input"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return
	}
	ctx.FilePath = payload.ToolInput.FilePath
	if payload.ToolInput.Content != "" {
		ctx.FileContent = []byte(payload.ToolInput.Content)
	}
	ctx.BashCommand = payload.ToolInput.Command
}
