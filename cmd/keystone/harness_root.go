package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/tacoda/keystone/internal/framework/config"
)

// extractHarnessRootFlag pulls --harness-root <name> (and --harness-root=<name>)
// out of args, returning the explicit flag value (or "" when not passed) plus
// args minus those tokens. Used by every subcommand other than `init` to
// strip the flag before its own parser sees it.
//
// `init` parses --harness-root inside parseInitArgs alongside its other
// catalog-driven flags.
func extractHarnessRootFlag(args []string) (flagValue string, remaining []string, err error) {
	remaining = make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--harness-root":
			if i+1 >= len(args) {
				return "", nil, fmt.Errorf("flag %s requires a value", a)
			}
			flagValue = args[i+1]
			i++
		case strings.HasPrefix(a, "--harness-root="):
			flagValue = strings.TrimPrefix(a, "--harness-root=")
		default:
			remaining = append(remaining, a)
		}
	}
	return flagValue, remaining, nil
}

// resolveHarnessRoot returns the harness folder name for a non-init command.
// Precedence:
//  1. flagValue if non-empty (--harness-root <name> on the command line)
//  2. keystone.json's harness_root field, if a project config exists at projectDir
//  3. config.DefaultHarnessRoot
//
// A malformed keystone.json (existing file that fails to parse) returns an
// error so the user fixes the file rather than silently falling back to the
// default.
func resolveHarnessRoot(projectDir, flagValue string) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}
	cfg, err := config.ReadProjectConfig(projectDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return config.DefaultHarnessRoot, nil
		}
		return "", fmt.Errorf("read %s: %w", config.ProjectConfigFile, err)
	}
	return cfg.ResolvedHarnessRoot(), nil
}
