package main

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/tacoda/keystone/internal/framework/scaffold"
)

// assets is the embedded scaffold template tree, rooted at templates/ so
// callers see harness/, targets/, optional/, patches/ at its top level.
// scaffold.Templates is an fs.FS; embed.FS lives inside the scaffold package.
var assets fs.FS = scaffold.Templates

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
	case "target":
		if err := runTarget(os.Args[2:], assets); err != nil {
			fmt.Fprintf(os.Stderr, "keystone: %v\n", err)
			os.Exit(1)
		}
	case "install":
		if err := runInstall(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "keystone: %v\n", err)
			os.Exit(1)
		}
	case "plugin":
		if err := runPlugin(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "keystone: %v\n", err)
			os.Exit(1)
		}
	case "verify":
		if err := runVerify(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "keystone: %v\n", err)
			os.Exit(1)
		}
	case "new":
		if err := runNew(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "keystone: %v\n", err)
			os.Exit(1)
		}
	case "policy":
		fmt.Fprintln(os.Stderr, "keystone: the policy command was removed in 1.0; use `keystone plugin add|update|remove` against keystone.json instead.")
		os.Exit(2)
	case "patch":
		if err := runPatch(os.Args[2:], assets); err != nil {
			fmt.Fprintf(os.Stderr, "keystone: %v\n", err)
			os.Exit(1)
		}
	case "migrate":
		fmt.Fprintln(os.Stderr, "keystone: `migrate` was renamed to `patch` in 1.0. Run `keystone patch ...` instead.")
		os.Exit(2)
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
  keystone install [--dir <path>] [--harness-root <name>]
  keystone plugin add <shorthand> [--name <n>] [--dir <path>] [--harness-root <name>]
  keystone plugin update <name> [@<new-version>] [--dir <path>] [--harness-root <name>]
  keystone plugin remove <name> [--dir <path>] [--harness-root <name>]
  keystone verify [--dir <path>] [--harness-root <name>]
  keystone new <port> <name> [flags]                          (see 'new help')
  keystone target add <agent>[,<agent>...] [--dir <path>] [--harness-root <name>]
  keystone patch [<dir>] [--apply|-y] [--dry-run] [--from <version>] [--harness-root <name>]
  keystone options
  keystone version
  keystone help

Commands:
  init       Scaffold the harness folder and the agent menu file(s) into <dir> (default: .)
  install    Materialize every plugin declared in keystone.json
  plugin     Manage installed plugins (add, update, remove) — see 'plugin help'
  verify     Check vendored plugins for drift and the strict cascade for violations
  new        Scaffold a new file at the conventional path — see 'new help'
  target     Manage agent targets installed under <harness-root>/adapters/ (see 'target help')
  patch      Apply pending framework patches to an existing install
  options    Print the allowed labels for every option flag
  version    Print the binary version
  help       Print this message

Behavior:
  When run in a TTY with options unset, keystone prompts interactively
  (via huh) for each missing option. When stdin is not a TTY, the agent
  must be supplied via --agent (or detected); other options stay unset.

  Existing harness folders are kept intact by default — only new files are
  written. Pass --reset --i-understand-this-is-destructive to wipe and
  rewrite an existing harness.

Flags for init:
  --reset                              Wipe the existing harness and rewrite from
                                       templates. Requires the confirm flag below.
  --i-understand-this-is-destructive   Pair with --reset to confirm the wipe.
  --harness-root <name>                Folder name for the harness (default: harness).

  --agent <label>             Agent target to install. Detected from marker
                              files if unset.
  --app-type <label>          Application type.
  --architecture <a,b,...>    Architecture preference(s) — comma-separated.
  --testing <a,b,...>         Testing approach(es) — comma-separated.
  --compliance <a,...>        Compliance scope — comma-separated.
  --starter <a,b,...>         Starter content pack(s) — e.g. universal-principles.
  --policy <ref>              Install an org policy. Repeatable.
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
