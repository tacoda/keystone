package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// runNewHook handles `keystone new hook <id>`. Scaffolds
// <harness-root>/hooks/<id>.md — an automated check that projects to a
// host hook via `host_triggers:`. The 2.x `sensor` kind split in 3.0:
// computational checks are `hook`; inferential review folds into `agent`.
func runNewHook(args []string) error {
	projectDir, harnessRoot, remaining, err := parseDirAndHarnessRoot(args)
	if err != nil {
		return err
	}
	if len(remaining) != 1 {
		return fmt.Errorf("`keystone new hook` requires exactly one argument: <id>")
	}
	id := remaining[0]
	if err := validatePrimitiveID(id); err != nil {
		return fmt.Errorf("hook id: %w", err)
	}
	diskName := strings.ReplaceAll(id, ":", "-")
	path := filepath.Join(projectDir, harnessRoot, "hooks", diskName+".md")
	body := fmt.Sprintf(`---
kind: hook
id: %s
description: TODO — one-line description of what this hook checks.
fix: ""
host_triggers:
  - phase: PreToolUse
    command: TODO — the shell command that runs the check.
---

# %s

What this hook verifies, and why. Document the exact command, what a
pass / fail looks like, and what the agent should do on failure.

## Remediation

What fixes a failure. If it is mechanical, set the fix: field to the
command and keystone verify --fix will run it.
`, id, id)
	return writeSkeleton(path, body)
}

// runNewDocument handles `keystone new document <id>`. Scaffolds
// <harness-root>/documents/<id>.md — a template for a governed output
// document (plan, review, ADR, retro, feature). Instances are written
// under .keystone/work/ and advanced through `gates:` by `keystone
// document promote`.
func runNewDocument(args []string) error {
	projectDir, harnessRoot, remaining, err := parseDirAndHarnessRoot(args)
	if err != nil {
		return err
	}
	if len(remaining) != 1 {
		return fmt.Errorf("`keystone new document` requires exactly one argument: <id>")
	}
	id := remaining[0]
	if err := validatePrimitiveID(id); err != nil {
		return fmt.Errorf("document id: %w", err)
	}
	diskName := strings.ReplaceAll(id, ":", "-")
	path := filepath.Join(projectDir, harnessRoot, "documents", diskName+".md")
	body := fmt.Sprintf(`---
kind: document
id: %s
description: TODO — one-line description of what this document captures.
type: ""
produced_by: ""
gates:
  - draft
  - in-review
  - approved
  - executed
  - done
---

# %s

The template another operator fills in. Each section is a heading the
author completes; the filled instance lands in .keystone/work/<task-id>/.

## Sections

Describe the sections this document must carry.
`, id, id)
	return writeSkeleton(path, body)
}
