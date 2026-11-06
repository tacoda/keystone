package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tacoda/keystone/internal/framework/projection"
)

// runNewAdapter handles `keystone new adapter <name>`: a host keystone has
// no built-in profile for, defined in .agents/hosts/<name>.json. Each kind
// starts indexed in AGENTS.md; the author sets native, link, or copy (with
// the host's folder) where the host reads it.
func runNewAdapter(args []string) error {
	projectDir, _, remaining, err := parseDirAndCharterRoot(args)
	if err != nil {
		return err
	}
	if len(remaining) != 1 {
		return fmt.Errorf("`keystone new adapter` requires exactly one argument: <name>")
	}
	name := remaining[0]
	body, err := projection.CustomHostSkeleton(name)
	if err != nil {
		return err
	}
	if err := writeSkeleton(filepath.Join(projectDir, projection.CustomHostsDir, name+".json"), string(body)); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, `
Next:
  1. For each kind the host reads, set "strategy": "native" (it reads .agents/),
     "link" or "copy" (add "dir": the host's folder), or leave "index".
  2. Run: keystone target add %s
     keystone projects to it from then on. Hooks and permissions are reported
     as gaps: a custom host gets no bridge or translation.
`, name)
	return nil
}
