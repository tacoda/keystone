package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// runNewPermission handles `keystone new permission <id>`: allow/ask/deny
// rules in Claude Code syntax, translated to every host that can enforce them.
func runNewPermission(args []string) error {
	id, projectDir, charterRoot, err := newOneArg(args, "permission")
	if err != nil {
		return err
	}
	path := filepath.Join(projectDir, charterRoot, "permissions", strings.ReplaceAll(id, ":", "-")+".md")
	body := fmt.Sprintf(`---
kind: permission
id: %s
description: TODO — what these permission rules protect.
allow:
  - "Bash(go test:*)"
ask:
  - "Bash(git push:*)"
deny:
  - "Read(./.env*)"
---

# %s

Why these grants are set this way — what the allow/ask/deny lists protect and
the trust boundary they encode.
`, id, id)
	return writeSkeleton(path, body)
}

// runNewHook handles `keystone new hook <id>`: a computational check that
// runs on a host event and blocks the agent when it exits non-zero.
func runNewHook(args []string) error {
	id, projectDir, charterRoot, err := newOneArg(args, "hook")
	if err != nil {
		return err
	}
	path := filepath.Join(projectDir, charterRoot, "hooks", strings.ReplaceAll(id, ":", "-")+".md")
	body := fmt.Sprintf(`---
kind: hook
id: %s
description: TODO — what this check catches.
mode: computational
event: Stop        # a host event (PreToolUse, PostToolUse, Stop, …) or a keystone signal
run: "true"        # TODO — the check; a non-zero exit blocks
---

# %s

What the check verifies, what a failure means, and how to fix one.
`, id, id)
	return writeSkeleton(path, body)
}

// runNewTool handles `keystone new tool <id>`. A tool is an author-defined
// callable the agent invokes on demand (Claude's generic sense of "tool").
// Its transport is one of: an MCP server keystone registers it on, a plugin,
// or a plain CLI. The scaffold declares the `transport:`, a `run:` handler,
// and the `args:` input schema.
func runNewTool(args []string) error {
	id, projectDir, charterRoot, err := newOneArg(args, "tool")
	if err != nil {
		return err
	}
	diskName := strings.ReplaceAll(id, ":", "-")
	path := filepath.Join(projectDir, charterRoot, "tools", diskName+".md")
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
// require exactly one id, validate it. Returns (id, projectDir, charterRoot).
func newOneArg(args []string, kind string) (id, projectDir, charterRoot string, err error) {
	projectDir, charterRoot, remaining, err := parseDirAndCharterRoot(args)
	if err != nil {
		return "", "", "", err
	}
	if len(remaining) != 1 {
		return "", "", "", fmt.Errorf("`keystone new %s` requires exactly one argument: <id>", kind)
	}
	if err := validatePrimitiveID(remaining[0]); err != nil {
		return "", "", "", fmt.Errorf("%s id: %w", kind, err)
	}
	return remaining[0], projectDir, charterRoot, nil
}
