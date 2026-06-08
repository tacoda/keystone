package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// runNew dispatches `keystone new <port> <name>` to the matching scaffold.
// Generators emit skeleton markdown that already conforms to the path
// convention (harness-root-relative inter-harness links, no ../) and the
// port contract (frontmatter, required sections). The author fills in
// the body.
func runNew(args []string) error {
	if len(args) == 0 {
		printNewUsage(os.Stderr)
		return fmt.Errorf("`keystone new` requires a port and name")
	}
	switch args[0] {
	case "help", "--help", "-h":
		printNewUsage(os.Stdout)
		return nil
	case "guide":
		return runNewGuide(args[1:])
	case "corpus":
		return runNewCorpus(args[1:])
	case "sensor":
		return runNewSensor(args[1:])
	case "action":
		return runNewAction(args[1:])
	case "playbook":
		return runNewPlaybook(args[1:])
	case "adapter":
		return runNewAdapter(args[1:])
	case "plugin":
		return runNewPlugin(args[1:])
	default:
		return fmt.Errorf("unknown port %q (try: guide, corpus, sensor, action, playbook, adapter, plugin)", args[0])
	}
}

func printNewUsage(w *os.File) {
	fmt.Fprint(w, `keystone new — scaffold a new file at the conventional path

Usage:
  keystone new guide <topic>/<name>     [--dir <path>] [--harness-root <name>]
  keystone new corpus <topic>/<name>    [--dir <path>] [--harness-root <name>]
  keystone new sensor <name>            [--dir <path>] [--harness-root <name>] [--kind <k>]
  keystone new action <name>            [--dir <path>] [--harness-root <name>]
  keystone new playbook <name>          [--dir <path>] [--harness-root <name>]
  keystone new adapter <agent>          [--dir <path>] [--harness-root <name>]
  keystone new plugin <name>            [--dir <path>]

Each generator drops a skeleton at the conventional path with the right
frontmatter, sections, and harness-root-relative cross-references.

Examples:
  keystone new guide process/release            # writes harness/guides/process/release.md
                                                # + harness/corpus/process/release.md
  keystone new sensor lint --kind computational # writes harness/sensors/lint.md
  keystone new playbook ship                    # writes harness/playbooks/ship.md
  keystone new plugin acme-policies             # scaffolds ./acme-policies/ as a plugin repo
`)
}

// parseDirAndHarnessRoot is the shared flag parser for in-harness
// generators. Returns (projectDir, harnessRoot, remaining-args).
func parseDirAndHarnessRoot(args []string) (projectDir, harnessRoot string, remaining []string, err error) {
	flagValue, args, err := extractHarnessRootFlag(args)
	if err != nil {
		return "", "", nil, err
	}
	dir := "."
	remaining = make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--dir":
			if i+1 >= len(args) {
				return "", "", nil, fmt.Errorf("flag %s requires a value", a)
			}
			dir = args[i+1]
			i++
		case strings.HasPrefix(a, "--dir="):
			dir = strings.TrimPrefix(a, "--dir=")
		default:
			remaining = append(remaining, a)
		}
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", "", nil, fmt.Errorf("resolve dir: %w", err)
	}
	harnessRoot, err = resolveHarnessRoot(absDir, flagValue)
	if err != nil {
		return "", "", nil, err
	}
	return absDir, harnessRoot, remaining, nil
}

// writeSkeleton writes content to path, refusing to overwrite an existing
// file. Creates parent directories as needed. Logs a "wrote:" line.
func writeSkeleton(path, content string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s already exists — refusing to overwrite (edit the file directly or move it aside)", path)
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "  wrote: %s\n", path)
	return nil
}
