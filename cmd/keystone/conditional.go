package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// installConditional copies opinion-specific files from the embedded `optional/`
// tree into destDir based on the user's selections.
//
// Layout convention: optional/<category-id>/<label-id>/charter/<...mirrors charter layout...>
// e.g. optional/architecture/hexagonal/charter/corpus/principles/hexagonal.md →
//
//	<destDir>/<charterRoot>/corpus/principles/hexagonal.md
//
// The embedded path always uses literal `charter/` as the second segment; the
// destination uses the consumer's configured charterRoot (default "charter").
//
// `agent` is excluded because the agent-specific bundle is already handled by
// the `targets/` copy in init.go.
//
// Missing optional roots are silently skipped — the directory is only present
// for categories the project chose to maintain content for.
func installConditional(assets fs.FS, destDir, charterRoot string, sel Selections) error {
	for catID, values := range sel {
		if catID == "agent" {
			continue
		}
		for _, label := range values {
			srcRoot := fmt.Sprintf("optional/%s/%s/charter", catID, label)
			if _, err := fs.Stat(assets, srcRoot); err != nil {
				continue // no content shipped for this label — that's fine
			}
			fmt.Fprintf(os.Stdout, "▸ installing optional content for %s=%s\n", catID, label)
			dest := filepath.Join(destDir, charterRoot)
			if err := copyTree(assets, srcRoot, dest, skipIfExists); err != nil {
				return fmt.Errorf("copy optional %s/%s: %w", catID, label, err)
			}
		}
	}
	return nil
}
