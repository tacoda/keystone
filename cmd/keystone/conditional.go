package main

import (
	"fmt"
	"io/fs"
	"os"
)

// installConditional copies opinion-specific files from the embedded `optional/`
// tree into destDir based on the user's selections.
//
// Layout convention: optional/<category-id>/<label-id>/<...mirrors destDir...>
// e.g. optional/architecture/hexagonal/harness/corpus/principles/hexagonal.md →
//      <destDir>/harness/corpus/principles/hexagonal.md
// (Paired guide files at optional/.../harness/guides/principles/hexagonal.md
// land at <destDir>/harness/guides/principles/hexagonal.md.)
//
// `agent` is excluded because the agent-specific bundle is already handled by
// the `targets/` copy in init.go.
//
// Missing optional roots are silently skipped — the directory is only present
// for categories the project chose to maintain content for.
func installConditional(assets fs.FS, destDir string, sel Selections) error {
	for catID, values := range sel {
		if catID == "agent" {
			continue
		}
		for _, label := range values {
			root := fmt.Sprintf("optional/%s/%s", catID, label)
			if _, err := fs.Stat(assets, root); err != nil {
				continue // no content shipped for this label — that's fine
			}
			fmt.Fprintf(os.Stdout, "▸ installing optional content for %s=%s\n", catID, label)
			if err := copyTree(assets, root, destDir, skipIfExists); err != nil {
				return fmt.Errorf("copy optional %s/%s: %w", catID, label, err)
			}
		}
	}
	return nil
}
