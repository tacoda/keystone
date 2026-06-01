package main

import (
	"embed"
	"fmt"
	"os"
)

//go:embed all:harness all:targets all:optional
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
  keystone options
  keystone version
  keystone help

Commands:
  init      Scaffold harness/ and the agent menu file(s) into <dir> (default: .)
  options   Print the allowed labels for every option flag
  version   Print the binary version
  help      Print this message

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

Categories the bootstrap action infers from the codebase (not asked here):
language, database, ci-platform, deployment target.
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
