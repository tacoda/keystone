package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tacoda/keystone/internal/framework/config"
)

// runNew dispatches `keystone new <port> <name>` to the matching scaffold.
// Generators emit skeleton markdown that already conforms to the path
// convention (harness-root-relative inter-harness links, no ../) and the
// port contract (frontmatter, required sections). The author fills in
// the body.
// newDispatch maps each `keystone new <verb>` to its scaffold generator.
var newDispatch = map[string]func([]string) error{
	"rule":     runNewRule,
	"hook":     runNewHook,
	"command":  runNewCommand,
	"skill":    runNewSkill,
	"agent":    runNewAgent,
	"pattern":  runNewPattern,
	"posture":  runNewPosture,
	"tool":     runNewTool,
	"document": runNewDocument,
	"corpus":   runNewCorpus,
	"eval":     runNewEval,
	"adapter":  runNewAdapter,
	"policy":   runNewPolicy,
}

func runNew(args []string) error {
	if len(args) == 0 {
		printNewUsage(os.Stderr)
		return fmt.Errorf("`keystone new` requires a port and name")
	}
	switch args[0] {
	case "help", "--help", "-h":
		printNewUsage(os.Stdout)
		return nil
	}
	gen, ok := newDispatch[args[0]]
	if !ok {
		return fmt.Errorf("unknown kind %q (use: rule, hook, command, skill, agent, pattern, posture, tool, document, corpus, eval, adapter, policy)", args[0])
	}
	return gen(args[1:])
}

func printNewUsage(w *os.File) {
	fmt.Fprint(w, `keystone new — scaffold a new file at the conventional path

Usage:

    keystone new rule <topic>/<name>      [--dir <path>]   # glob-scoped directive + paired corpus
    keystone new hook <name>              [--dir <path>]   # automated check (projects to host hook)
    keystone new command <id>             [--dir <path>]   # a unit of work / lifecycle step
    keystone new skill <id>               [--dir <path>]   # a composed capability
    keystone new agent <id>               [--dir <path>]   # a role spawned as a subagent
    keystone new pattern <id>             [--dir <path>]   # prose documentation pattern (tutorial, how-to, reference, explanation)
    keystone new posture <id>             [--dir <path>]   # tool/permission posture (allow/ask/deny)
    keystone new tool <id>                [--dir <path>]   # author-defined callable (transport: cli | mcp | plugin)
    keystone new document <id>            [--dir <path>]   # governed output template (plan/review/adr/...)
    keystone new corpus <topic>/<name>    [--dir <path>]   # on-demand reasoning
    keystone new eval <id>                [--dir <path>]
    keystone new adapter <agent>          [--dir <path>]
    keystone new policy <name>            [--dir <path>]

Each generator drops a skeleton at the conventional path with the right
frontmatter, sections, and harness-root-relative cross-references. The
harness root is .keystone/harness/.

Ids may use the colon-namespaced form (e.g. keystone:index); the disk
filename normalizes : to -.

Examples:
  keystone new rule process/release        # .keystone/harness/rules/process/release.md
                                           # + paired .keystone/harness/corpus/process/release.md
  keystone new hook lint                    # .keystone/harness/hooks/lint.md
  keystone new skill keystone:index         # .keystone/harness/skills/keystone-index/SKILL.md
  keystone new agent security-reviewer      # .keystone/harness/agents/security-reviewer.md
  keystone new command review               # .keystone/harness/commands/review.md
  keystone new document adr                 # .keystone/harness/documents/adr.md
  keystone new policy acme-policies         # ./acme-policies/ (policy repo skeleton)
`)
}

// parseDirAndHarnessRoot is the shared flag parser for in-harness
// generators. Returns (projectDir, harnessRoot, remaining-args). The
// harness root is fixed at 2.0; only --dir is honored as a flag.
func parseDirAndHarnessRoot(args []string) (projectDir, harnessRoot string, remaining []string, err error) {
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
	return absDir, config.DefaultHarnessRoot, remaining, nil
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
