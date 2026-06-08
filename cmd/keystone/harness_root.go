package main

import (
	"fmt"
	"strings"

	"github.com/tacoda/keystone/internal/framework/config"
)

// extractHarnessRoot pulls --harness-root <name> (and --harness-root=<name>)
// out of args, returning the resolved root (defaulting to
// config.DefaultHarnessRoot) plus args minus those tokens. Used by every
// subcommand other than `init` so the flag works uniformly across the CLI.
//
// `init` parses --harness-root inside parseInitArgs alongside its other
// catalog-driven flags.
func extractHarnessRoot(args []string) (root string, remaining []string, err error) {
	root = config.DefaultHarnessRoot
	remaining = make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--harness-root":
			if i+1 >= len(args) {
				return "", nil, fmt.Errorf("flag %s requires a value", a)
			}
			root = args[i+1]
			i++
		case strings.HasPrefix(a, "--harness-root="):
			root = strings.TrimPrefix(a, "--harness-root=")
		default:
			remaining = append(remaining, a)
		}
	}
	if root == "" {
		return "", nil, fmt.Errorf("--harness-root cannot be empty")
	}
	return root, remaining, nil
}
