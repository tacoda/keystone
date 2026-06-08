package main

import (
	"fmt"
	"os"
)

// runPlugin dispatches `keystone plugin <subcommand> ...`.
func runPlugin(args []string) error {
	if len(args) == 0 {
		printPluginUsage(os.Stderr)
		return fmt.Errorf("plugin requires a subcommand")
	}
	switch args[0] {
	case "add":
		return runPluginAdd(args[1:])
	case "update":
		return runPluginUpdate(args[1:])
	case "remove", "rm":
		return runPluginRemove(args[1:])
	case "help", "--help", "-h":
		printPluginUsage(os.Stdout)
		return nil
	default:
		return fmt.Errorf("unknown plugin subcommand %q (try: add, update, remove)", args[0])
	}
}

func printPluginUsage(w *os.File) {
	fmt.Fprint(w, `keystone plugin — manage installed plugins

Usage:
  keystone plugin add <shorthand> [--name <n>] [--dir <path>] [--harness-root <name>]
  keystone plugin update <name> [@<new-version>] [--dir <path>] [--harness-root <name>]
  keystone plugin remove <name> [--dir <path>] [--harness-root <name>]
  keystone plugin help

Commands:
  add       Append a plugin to keystone.json and install it.
  update    Bump a plugin's pinned version (or re-fetch the recorded ref)
            and reinstall.
  remove    Remove a plugin from keystone.json and delete its vendor
            directory.

Shorthand source format:
  [<host>/]<owner>/<repo>[@<version>]

Examples:
  keystone plugin add tacoda/tacoda-org@0.2.0
  keystone plugin add gitlab.com/acme/policies@main
  keystone plugin update tacoda-org @0.3.0
  keystone plugin remove tacoda-org
`)
}
