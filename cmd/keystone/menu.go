package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// agentMenuFile maps each agent to the relative path of its single "menu file" —
// the discovery file the agent reads on session start (CLAUDE.md, CONVENTIONS.md,
// .github/copilot-instructions.md, etc.).
//
// Cursor is intentionally absent: it has no single menu, only an array of
// independent .cursor/rules/*.mdc files. Each .mdc installs via skipIfExists.
var agentMenuFile = map[string]string{
	"_generic":       "AGENTS.md",
	"aider":          "CONVENTIONS.md",
	"claude-code":    "CLAUDE.md",
	"cline":          "cline-instructions.md",
	"codex":          "AGENTS.md",
	"continue":       ".continuerules",
	"github-copilot": ".github/copilot-instructions.md",
	"goose":          ".goosehints",
	"pi":             "AGENTS.md",
}

const (
	menuStartMarker = "<!-- keystone:start -->"
	menuEndMarker   = "<!-- keystone:end -->"
)

// installMenuFile installs the agent's menu file into destDir with merge semantics:
//
//   - Destination missing → write the bracketed section as the new file.
//   - Destination has keystone markers → replace everything between them (idempotent).
//   - Destination has a first-line H1 (`# `) → insert the bracketed section right
//     after the H1 (and any blank line following it). Preserves user content.
//   - Destination has no H1 → prepend the bracketed section at the top.
//
// Returns the relative path (or "") to signal callers what to skip in the regular
// target copy. An agent without a registered menu file is a no-op.
func installMenuFile(assets embed.FS, agent, destDir string) (string, error) {
	rel, ok := agentMenuFile[agent]
	if !ok {
		return "", nil
	}

	srcPath := filepath.Join("targets", agentTargetDir(agent), rel)
	insert, err := fs.ReadFile(assets, srcPath)
	if err != nil {
		// Target has no menu file shipped — nothing to install.
		return rel, nil
	}
	insertStr := strings.TrimSpace(string(insert))

	bracketed := fmt.Sprintf("%s\n%s\n%s", menuStartMarker, insertStr, menuEndMarker)

	destPath := filepath.Join(destDir, rel)
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return rel, err
	}

	existing, err := os.ReadFile(destPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return rel, err
		}
		// Fresh install — write the bracketed section as the entire file.
		// No auto-generated H1: it would be decorative for .md menus and ugly
		// for dotfiles. The user adds their own H1 when they're ready.
		newContent := bracketed + "\n"
		if err := os.WriteFile(destPath, []byte(newContent), 0o644); err != nil {
			return rel, err
		}
		fmt.Fprintf(os.Stdout, "  wrote: %s\n", destPath)
		return rel, nil
	}

	merged := mergeMenuSection(string(existing), bracketed)
	if merged == string(existing) {
		fmt.Fprintf(os.Stdout, "  unchanged: %s\n", destPath)
		return rel, nil
	}
	if err := os.WriteFile(destPath, []byte(merged), 0o644); err != nil {
		return rel, err
	}
	fmt.Fprintf(os.Stdout, "  merged: %s\n", destPath)
	return rel, nil
}

// mergeMenuSection returns the existing file content with the bracketed harness
// block inserted (or refreshed). The function is pure — no I/O — so it's
// straightforward to unit-test.
func mergeMenuSection(existing, bracketed string) string {
	// Idempotent path: markers already present → replace what's between them.
	startIdx := strings.Index(existing, menuStartMarker)
	endIdx := strings.Index(existing, menuEndMarker)
	if startIdx >= 0 && endIdx > startIdx {
		endIdx += len(menuEndMarker)
		return existing[:startIdx] + bracketed + existing[endIdx:]
	}

	// First-time insert: place the bracketed block after the first H1, or at the
	// very top if no H1 exists. Normalize blank lines so exactly one separates
	// our block from each neighbor, regardless of the original file's spacing.
	lines := strings.Split(existing, "\n")
	insertAt := 0
	for i, line := range lines {
		if strings.HasPrefix(line, "# ") {
			insertAt = i + 1
			break
		}
	}

	before := strings.TrimRight(strings.Join(lines[:insertAt], "\n"), "\n")
	after := strings.TrimLeft(strings.Join(lines[insertAt:], "\n"), "\n")

	switch {
	case before == "" && after == "":
		return bracketed + "\n"
	case before == "":
		return bracketed + "\n\n" + after
	case after == "":
		return before + "\n\n" + bracketed + "\n"
	default:
		return before + "\n\n" + bracketed + "\n\n" + after
	}
}
