package main

import (
	"embed"
	"fmt"
	"os"
)

//go:embed all:harness all:targets all:optional all:migrations
var assets embed.FS

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		printUsage(os.Stderr)
		os.Exit(2)
	}

	switch os.Args[1] {
	case "init":
		if err := runInit(os.Args[2:], assets); err != nil {
			fmt.Fprintf(os.Stderr, "keystone: %v\n", err)
			os.Exit(1)
		}
	case "add-target":
		if err := runAddTarget(os.Args[2:], assets); err != nil {
			fmt.Fprintf(os.Stderr, "keystone: %v\n", err)
			os.Exit(1)
		}
	case "policy":
		if err := runPolicy(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "keystone: %v\n", err)
			os.Exit(1)
		}
	case "migrate":
		if err := runMigrate(os.Args[2:], assets); err != nil {
			fmt.Fprintf(os.Stderr, "keystone: %v\n", err)
			os.Exit(1)
		}
	case "options":
		printOptionLabels(os.Stdout)
	case "version", "--version", "-v":
		fmt.Println(version)
	case "help", "--help", "-h":
		printUsage(os.Stdout)
	default:
		fmt.Fprintf(os.Stderr, "keystone: unknown command %q\n\n", os.Args[1])
		printUsage(os.Stderr)
		os.Exit(2)
	}
}

func printUsage(w *os.File) {
	fmt.Fprint(w, `keystone — install the project harness into a directory

Usage:
  keystone init [<dir>] [flags]
  keystone add-target <agent>[,<agent>...] [<dir>]
  keystone policy update <name> [<new-ref>] [--dir <path>] [--force]
  keystone migrate [<dir>] [--apply|-y] [--dry-run] [--from <version>]
  keystone options
  keystone version
  keystone help

Commands:
  init         Scaffold harness/ and the agent menu file(s) into <dir> (default: .)
  add-target   Install another agent target bundle into an existing harness
  policy       Manage org policies installed under harness/policies/ (see 'policy help')
  migrate      Apply pending harness migrations to an existing install
  options      Print the allowed labels for every option flag
  version      Print the binary version
  help         Print this message

Behavior:
  When run in a TTY with options unset, keystone prompts interactively
  (via huh) for each missing option. When stdin is not a TTY, the agent
  must be supplied via --agent (or detected); other options stay unset.

Flags for init:
  --force          Overwrite an existing harness/ in <dir> without prompting.

  --agent <label>             Agent target to install. Detected from marker
                              files if unset.
  --app-type <label>          Application type.
  --architecture <a,b,...>    Architecture preference(s) — comma-separated.
  --testing <a,b,...>         Testing approach(es) — comma-separated.
  --compliance <a,...>        Compliance scope — comma-separated.
  --policy <ref>              Install an org policy into harness/policies/. Repeatable.
                              v1 supports git+<url>[#<rev>]:
                                --policy git+https://github.com/acme/policy.git#v1.2.0

Categories the bootstrap action infers from the codebase (not asked here):
language, frameworks, libraries, database, ci-platform.
`)
}

// printOptionLabels writes the full catalog of allowed labels per category.
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
