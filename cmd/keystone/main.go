// Package main is the keystone CLI binary entry point.
//
// Command tree wiring lives in root.go (the Cobra root + every
// subcommand). Per-command handlers stay in their per-verb files
// (init.go, doctor.go, mcp.go, …). Most handlers still accept a raw
// args slice for back-compat with the pre-Cobra dispatch; a later
// pass will lift their flag parsing into Cobra-native declarations.
package main

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/tacoda/keystone/internal/framework/scaffold"
)

// assets is the embedded scaffold template tree, rooted at templates/.
var assets fs.FS = scaffold.Templates

// version is stamped by the build (goreleaser). The fallback is the
// development sentinel.
var version = "dev"

func main() {
	SetAssets(assets)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "keystone: %v\n", err)
		os.Exit(1)
	}
}

// printOptionLabels writes the full catalog of allowed labels per category.
// Used by `keystone options`.
func printOptionLabels(w *os.File) {
	for _, c := range categories {
		fmt.Fprintf(w, "--%s\n", c.ID)
		if c.MultiSelect {
			fmt.Fprintf(w, "  (multi-select; comma-separated)\n")
		}
		for _, v := range c.Values {
			fmt.Fprintf(w, "  %-20s  %s\n", v.ID, v.Description)
		}
		fmt.Fprintln(w)
	}
}
