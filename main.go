package main

import (
	"embed"
	"fmt"
	"os"
)

//go:embed all:harness all:targets
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
  keystone init [<dir>] [--agent <name>] [--force]
  keystone version
  keystone help

Commands:
  init      Scaffold harness/ and the agent menu file(s) into <dir> (default: .)
  version   Print the binary version
  help      Print this message

Flags for init:
  --agent <name>   Agent target to install. One of:
                   claude-code, codex, pi, cursor, aider,
                   github-copilot-cli, continue, cline, goose, _generic
                   If omitted, keystone attempts detection from existing files
                   in <dir>; if detection fails, init exits with an error.
  --force          Overwrite an existing harness/ in <dir> without prompting.
`)
}
