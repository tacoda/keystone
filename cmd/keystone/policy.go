package main

import (
	"fmt"
	"os"
)

// runPolicy dispatches `keystone policy <subcommand> ...`.
func runPolicy(args []string) error {
	if len(args) == 0 {
		printPolicyUsage(os.Stderr)
		return fmt.Errorf("policy requires a subcommand")
	}
	switch args[0] {
	case "add":
		return runPolicyAdd(args[1:])
	case "update":
		return runPolicyUpdate(args[1:])
	case "remove", "rm":
		return runPolicyRemove(args[1:])
	case "help", "--help", "-h":
		printPolicyUsage(os.Stdout)
		return nil
	default:
		return fmt.Errorf("unknown policy subcommand %q (try: add, update, remove)", args[0])
	}
}

func printPolicyUsage(w *os.File) {
	fmt.Fprint(w, `keystone policy — manage installed policies

Usage:
  keystone policy add <shorthand> [--name <n>] [--dir <path>] [--harness-root <name>]
  keystone policy update <name> [@<new-version>] [--dir <path>] [--harness-root <name>]
  keystone policy remove <name> [--dir <path>] [--harness-root <name>]
  keystone policy help

Commands:
  add       Append a policy to keystone.json and install it.
  update    Bump a policy's pinned version (or re-fetch the recorded ref)
            and reinstall.
  remove    Remove a policy from keystone.json and delete its vendor
            directory.

Shorthand source format:
  [<host>/]<owner>/<repo>[@<version>]

Examples:
  keystone policy add tacoda/tacoda-org@0.2.0
  keystone policy add gitlab.com/acme/policies@main
  keystone policy update tacoda-org @0.3.0
  keystone policy remove tacoda-org
`)
}
