package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// writeInstallProfile renders sel as harness/state/INSTALL_PROFILE.md under
// destDir. Overwrites any existing file (the file is install-scoped — re-running
// init should reset it).
func writeInstallProfile(destDir string, sel Selections) error {
	path := filepath.Join(destDir, "harness", "state", "INSTALL_PROFILE.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	var b strings.Builder
	fmt.Fprintf(&b, "---\n")
	fmt.Fprintf(&b, "created: %s\n", time.Now().UTC().Format("2006-01-02"))
	fmt.Fprintf(&b, "keystone_version: %s\n", version)
	fmt.Fprintf(&b, "---\n\n")
	fmt.Fprintf(&b, "# Install Profile\n\n")
	fmt.Fprintf(&b, "Selections captured by `keystone init`. Read by the **bootstrap** action; safe to edit by hand.\n\n")
	fmt.Fprintf(&b, "## Selections\n\n")
	fmt.Fprintf(&b, "| Category | Value(s) |\n")
	fmt.Fprintf(&b, "|---|---|\n")

	// Iterate in catalog order, not map order, so the file is stable.
	for _, c := range categories {
		values := sel[c.ID]
		if len(values) == 0 {
			fmt.Fprintf(&b, "| %s | _(unset)_ |\n", c.ID)
			continue
		}
		fmt.Fprintf(&b, "| %s | %s |\n", c.ID, strings.Join(values, ", "))
	}

	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "  wrote: %s\n", path)
	return nil
}
