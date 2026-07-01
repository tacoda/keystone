package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// runSignal dispatches `keystone signal <sub>`: `fire` (dispatch every hook
// bound to a signal) and `list` (show built-in + project-declared signals).
//
// A signal is a keystone framework event the host cannot see — the higher-
// level, extensible counterpart to a host hook phase. The set is open: any
// non-host-phase event a hook binds to is a signal, so projects define their
// own by declaring them in keystone.json and firing `keystone signal fire`.
func runSignal(args []string) error {
	if len(args) == 0 {
		printSignalUsage(os.Stderr)
		return fmt.Errorf("`keystone signal` requires a subcommand (fire | list)")
	}
	switch args[0] {
	case "help", "--help", "-h":
		printSignalUsage(os.Stdout)
		return nil
	case "fire":
		if len(args) >= 2 && isHelpFlag(args[1]) {
			printSignalUsage(os.Stdout)
			return nil
		}
		// Dispatch is identical to a framework-hook fire; signals ARE the
		// framework events hooks bind to.
		return runHookFire(args[1:])
	case "list":
		return runSignalList(args[1:])
	default:
		return fmt.Errorf("unknown signal subcommand %q (use: fire | list)", args[0])
	}
}

// runSignalList prints the signals available in this project: keystone's
// built-ins plus any declared in keystone.json `signals:`. Host phases are
// listed separately for reference — they are NOT signals (the host fires
// them; keystone bridges them).
func runSignalList(args []string) error {
	absDir, err := filepath.Abs(dirArg(args))
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}

	fmt.Fprintln(os.Stdout, "Built-in signals (keystone-fired):")
	printSorted(primitive.BuiltinSignals)

	cfg, cfgErr := config.ReadProjectConfig(absDir)
	if cfgErr != nil && !errors.Is(cfgErr, os.ErrNotExist) {
		return fmt.Errorf("read keystone.json: %w", cfgErr)
	}
	if cfg != nil && len(cfg.Signals) > 0 {
		fmt.Fprintln(os.Stdout, "\nProject signals (keystone.json):")
		printSorted(cfg.Signals)
	}

	fmt.Fprintln(os.Stdout, "\nHost phases (host-fired, bridged — not signals):")
	printSorted(primitive.HostPhases)
	return nil
}

// dirArg returns the value of a `--dir` flag in args, or "." if absent.
func dirArg(args []string) string {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == "--dir" {
			return args[i+1]
		}
	}
	return "."
}

func printSorted(items []string) {
	cp := append([]string(nil), items...)
	sort.Strings(cp)
	for _, s := range cp {
		fmt.Fprintf(os.Stdout, "  %s\n", s)
	}
}

func printSignalUsage(w *os.File) {
	fmt.Fprint(w, `keystone signal — fire or list keystone signals

A signal is a keystone framework event the host cannot see — the extensible,
higher-level counterpart to a host hook phase. Any non-host-phase event a
primitive subscribes to via `+"`on:`"+` is a signal; declare custom ones in
keystone.json `+"`signals:`"+` for lint + discovery.

Usage:

    keystone signal fire <name> [--phase X] [--command Y] [--type Z] [--dir D]
    keystone signal list [--dir D]

`+"`fire`"+` dispatches every hook bound to <name>: computational hooks run their
`+"`run:`"+` script (parallel; non-zero exit blocks); inferential hooks are listed
as a dispatch manifest for the host to spawn. `+"`list`"+` shows built-in +
project-declared signals (and host phases, for reference).
`)
}
