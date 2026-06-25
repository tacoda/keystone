package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// runNewPattern handles `keystone new pattern <id>`. A pattern is prose — a
// reusable documentation pattern (the Diátaxis modes: tutorial, how-to,
// reference, explanation). The scaffold generates a documentation skeleton the
// author fills in; it is not a code generator and has no projection.
func runNewPattern(args []string) error {
	id, projectDir, harnessRoot, err := newOneArg(args, "pattern")
	if err != nil {
		return err
	}
	path := filepath.Join(projectDir, harnessRoot, "patterns", strings.ReplaceAll(id, ":", "-")+".md")
	body := fmt.Sprintf(`---
kind: pattern
id: %s
description: TODO — the documentation pattern this prescribes and when to use it.
---

# %s

## Purpose

The reader's need this document type serves (Diátaxis: learning / task /
information / understanding).

## Audience & context

Who reads a document of this type and what they already know.

## Structure

The section skeleton a document following this pattern uses.

## Conventions

Voice, length, what to include and what to leave out.

## Example

A short model document following the pattern.
`, id, id)
	return writeSkeleton(path, body)
}

// runNewPosture handles `keystone new posture <id>`. A posture declares
// allow/ask/deny tool-permission lists that project to the host's permissions
// block (Claude Code: .claude/settings.json).
func runNewPosture(args []string) error {
	id, projectDir, harnessRoot, err := newOneArg(args, "posture")
	if err != nil {
		return err
	}
	path := filepath.Join(projectDir, harnessRoot, "posture", strings.ReplaceAll(id, ":", "-")+".md")
	body := fmt.Sprintf(`---
kind: posture
id: %s
description: TODO — the tool/permission posture this declares.
allow:
  - "Bash(go test:*)"
ask:
  - "Bash(git push:*)"
deny:
  - "Read(.env*)"
---

# %s

Why these grants are set this way — what the allow/ask/deny lists protect and
the trust boundary they encode.
`, id, id)
	return writeSkeleton(path, body)
}

// runNewTool handles `keystone new tool <id>`. A tool is an author-defined
// callable the agent invokes on demand (Claude's generic sense of "tool").
// Its transport is one of: an MCP server keystone registers it on, a plugin,
// or a plain CLI. The scaffold declares the `transport:`, a `run:` handler,
// and the `args:` input schema.
func runNewTool(args []string) error {
	id, projectDir, harnessRoot, err := newOneArg(args, "tool")
	if err != nil {
		return err
	}
	diskName := strings.ReplaceAll(id, ":", "-")
	path := filepath.Join(projectDir, harnessRoot, "tools", diskName+".md")
	body := fmt.Sprintf(`---
kind: tool
id: %s
description: TODO — what the agent invokes this for.
transport: cli   # cli | mcp | plugin
run: ./scripts/%s.sh
args:
  - name: target
    type: string
    required: true
    description: TODO — the input this callable expects.
---

# %s

What this callable does, its inputs and outputs, and when the agent should
reach for it instead of a built-in tool.
`, id, diskName, id)
	return writeSkeleton(path, body)
}

// newOneArg is the shared front-half for single-id scaffolds: parse flags,
// require exactly one id, validate it. Returns (id, projectDir, harnessRoot).
func newOneArg(args []string, kind string) (id, projectDir, harnessRoot string, err error) {
	projectDir, harnessRoot, remaining, err := parseDirAndHarnessRoot(args)
	if err != nil {
		return "", "", "", err
	}
	if len(remaining) != 1 {
		return "", "", "", fmt.Errorf("`keystone new %s` requires exactly one argument: <id>", kind)
	}
	if err := validatePrimitiveID(remaining[0]); err != nil {
		return "", "", "", fmt.Errorf("%s id: %w", kind, err)
	}
	return remaining[0], projectDir, harnessRoot, nil
}
