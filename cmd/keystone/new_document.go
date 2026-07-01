package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// runNewDocument handles `keystone new document <id>`. Scaffolds
// <charter-root>/documents/<id>.md — a template for a governed output
// document (plan, review, ADR, retro, feature). Instances are written
// under .charter/work/ and advanced through `gates:` by `keystone
// document promote`.
func runNewDocument(args []string) error {
	projectDir, charterRoot, remaining, err := parseDirAndCharterRoot(args)
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
	path := filepath.Join(projectDir, charterRoot, "documents", diskName+".md")
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
author completes; the filled instance lands in .charter/work/<task-id>/.

## Sections

Describe the sections this document must carry.
`, id, id)
	return writeSkeleton(path, body)
}
