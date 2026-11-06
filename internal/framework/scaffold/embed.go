// Package scaffold owns the embedded template tree that `keystone init`
// writes into a consumer's repo. The templates are organized as they appear
// on disk after install:
//
//	templates/
//	├── agents/    — installed as .agents/ (rules, references, hooks,
//	│                skills, commands, agents, evals, …)
//	├── keystone/  — installed as .keystone/ (learning, archive, hosts)
//	└── optional/  — opt-in content (architecture/, compliance/, starter/)
//	                 pulled in by --architecture / --compliance / --starter
//
// Host files (AGENTS.md, .claude/, .cursor/, …) are not templates:
// `keystone project` generates them. The Templates fs.FS is rooted at
// templates/, so callers see "agents", "keystone", and "optional". (The 0.x patches subsystem
// was retired in 2.1; `keystone migrate` is the upgrade path.)
package scaffold

import (
	"embed"
	"io/fs"
)

//go:embed all:templates
var rawTemplates embed.FS

// Templates is the embedded scaffold tree, rooted at the templates directory.
// Top-level entries: agents/, keystone/, optional/.
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
