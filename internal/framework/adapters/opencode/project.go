// Package opencode projects keystone primitives into opencode's
// host-native layout. opencode discovers skills, subagents, and commands
// from `.opencode/` (plural directory names) — the same agent-kind →
// native-path mapping Claude Code uses, only the root differs:
//
//	.claude/skills/<id>/SKILL.md  → .opencode/skills/<id>/SKILL.md
//	.claude/agents/<id>.md        → .opencode/agents/<id>.md
//	.claude/commands/<id>.md      → .opencode/commands/<id>.md
//
// The projected bytes are byte-identical to the Claude Code projection —
// opencode reads the same Markdown-with-frontmatter shape — so this
// adapter reuses primitive.RenderForHost and only rewrites the root.
//
// Globbed guides also project, as rule-shims under `.opencode/rules/`, and
// the glob is registered in opencode.json's `instructions` array. opencode
// loads `instructions` files always (combined with AGENTS.md) — it has no
// per-edited-file gating like Cursor, so these are always-on. The Keystone
// MCP server is wired separately via `keystone mcp install --agent opencode`.
package opencode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

// Root is the relative directory opencode reads project primitives from.
const Root = ".opencode"

// claudeRoot is the projection root primitive.ProjectionRelPath emits; we
// rewrite it to Root for opencode.
const claudeRoot = ".claude"

// RulesGlob is the instructions-array entry registered in opencode.json so
// opencode loads the projected guide rule-shims. opencode's `instructions`
// field is a glob-selectable file list combined with AGENTS.md — unlike
// Cursor, it has no per-edited-file gating, so these load on every turn.
const RulesGlob = ".opencode/rules/*.md"

// ProjectionResult records what ProjectAgents wrote.
type ProjectionResult struct {
	Wrote     int // files newly written or content-changed
	Unchanged int // files whose content was already correct
	Rules     int // of Wrote+Unchanged, how many were guide rule-shims
}

// ProjectAgents mirrors every skill, subagent, command, and globbed guide
// into `.opencode/`. Guides become rule-shims under `.opencode/rules/` and
// the glob is registered in opencode.json's `instructions` array so opencode
// loads them. Re-runs are idempotent; kinds with no host file are skipped.
func ProjectAgents(projectDir string, primitives []primitive.Primitive) (ProjectionResult, error) {
	var out ProjectionResult
	for _, p := range primitives {
		sub, isRule, ok := opencodeSubpath(primitive.ProjectionRelPath(p))
		if !ok {
			continue
		}
		content, ok, err := primitive.RenderForHost(projectDir, p)
		if err != nil {
			return out, fmt.Errorf("render %s/%s: %w", p.Kind, p.ID, err)
		}
		if !ok {
			continue
		}
		if isRule {
			out.Rules++
		}
		dest := filepath.Join(projectDir, Root, sub)
		if prev, _ := os.ReadFile(dest); bytes.Equal(prev, content) {
			out.Unchanged++
			continue
		}
		if err := atomicWrite(dest, content); err != nil {
			return out, fmt.Errorf("write %s: %w", dest, err)
		}
		out.Wrote++
	}
	if out.Rules > 0 {
		if err := ensureInstructions(projectDir, RulesGlob); err != nil {
			return out, fmt.Errorf("wire opencode.json instructions: %w", err)
		}
	}
	return out, nil
}

// opencodeSubpath maps a `.claude/...` projection path to its `.opencode/`
// sub-path and reports whether it is a guide rule-shim. The
// skills/agents/commands/rules subtrees project to opencode; every other
// target returns ok=false.
func opencodeSubpath(claudeRel string) (sub string, isRule, ok bool) {
	if claudeRel == "" {
		return "", false, false
	}
	rest := strings.TrimPrefix(claudeRel, claudeRoot+string(filepath.Separator))
	if rest == claudeRel {
		return "", false, false // not under .claude
	}
	sep := string(filepath.Separator)
	switch {
	case strings.HasPrefix(rest, "rules"+sep):
		return rest, true, true
	case strings.HasPrefix(rest, "skills"+sep),
		strings.HasPrefix(rest, "agents"+sep),
		strings.HasPrefix(rest, "commands"+sep):
		return rest, false, true
	}
	return "", false, false
}

// ensureInstructions adds glob to opencode.json's `instructions` array,
// creating the file (with `$schema`) if absent. Idempotent — an existing
// entry is left untouched; other top-level keys (mcp, model, …) are
// preserved. Writes only when the array actually changed.
func ensureInstructions(projectDir, glob string) error {
	path := filepath.Join(projectDir, "opencode.json")

	doc := map[string]any{}
	if raw, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(raw, &doc) // best-effort; overwrite on parse failure
	}
	if _, ok := doc["$schema"]; !ok {
		doc["$schema"] = "https://opencode.ai/config.json"
	}

	var list []any
	if existing, ok := doc["instructions"].([]any); ok {
		for _, v := range existing {
			if s, ok := v.(string); ok && s == glob {
				return nil // already present — nothing to write
			}
		}
		list = existing
	}
	doc["instructions"] = append(list, glob)

	body, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	return atomicWrite(path, append(body, '\n'))
}

// atomicWrite — same temp+rename shape every keystone adapter uses.
func atomicWrite(destAbs string, contents []byte) error {
	if err := os.MkdirAll(filepath.Dir(destAbs), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(destAbs), err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(destAbs), ".keystone-opencode.*")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(contents); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("close temp: %w", err)
	}
	if err := os.Rename(tmpName, destAbs); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("rename %s -> %s: %w", tmpName, destAbs, err)
	}
	return nil
}
