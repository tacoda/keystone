// Package scaffold owns the embedded template tree that `keystone init`
// writes into a consumer's repo. The templates are organized as they appear
// on disk after install:
//
//	templates/
//	├── harness/   — the harness layout (guides, corpus, sensors,
//	│                actions, playbooks, adapters, learning, archive)
//	├── targets/   — per-agent menu files (CLAUDE.md, AGENTS.md,
//	│                .cursor/rules/, etc.) installed at the project root
//	└── optional/  — opt-in content (architecture/, compliance/, starter/)
//	                 pulled in by --architecture / --compliance / --starter
//
// The Templates fs.FS is rooted at templates/, so callers see "harness",
// "targets", and "optional" at its top level. (The 0.x patches subsystem
// was retired in 2.1; `keystone migrate` is the upgrade path.)
package scaffold

import (
	"embed"
	"io/fs"
)

//go:embed all:templates
var rawTemplates embed.FS

// Templates is the embedded scaffold tree, rooted at the templates directory.
// Top-level entries: harness/, targets/, optional/.
var Templates fs.FS

func init() {
	sub, err := fs.Sub(rawTemplates, "templates")
	if err != nil {
		// Embed paths are verified at compile time; a Sub failure here is a
		// programmer error during package setup, not a runtime condition.
		panic(err)
	}
	Templates = sub
}
