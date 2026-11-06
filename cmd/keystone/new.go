package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tacoda/keystone/internal/framework/config"
)

// runNew dispatches `keystone new <kind> <name>` to the matching scaffold.
// Generators emit skeleton markdown that already conforms to the path
// convention (charter-root-relative links, no ../) and the kind's contract
// (frontmatter, required sections). The author fills in the body.
var newDispatch = map[string]func([]string) error{
	// constraints
	"rule":       runNewRule,
	"reference":  runNewReference,
	"hook":       runNewHook,
	"permission": runNewPermission,
	"eval":       runNewEval,
	// capabilities
	"skill":    runNewSkill,
	"command":  runNewCommand,
	"subagent": runNewSubagent,
	"tool":     runNewTool,
	"document": runNewDocument,
	// keystone itself
	"adapter": runNewAdapter,
	"policy":  runNewPolicy,
}

// newRenamed maps each 4.x `keystone new` kind to its 5.0 name.
var newRenamed = map[string]string{
	"guide": "rule", "corpus": "reference", "sensor": "hook", "posture": "permission",
	"agent": "subagent", "playbook": "skill", "pattern": "skill",
}

const newKinds = "rule, reference, hook, permission, eval, skill, command, subagent, tool, document, adapter, policy"

func runNew(args []string) error {
	if len(args) == 0 {
		printNewUsage(os.Stderr)
		return fmt.Errorf("`keystone new` requires a kind and name")
	}
	switch args[0] {
	case "help", "--help", "-h":
		printNewUsage(os.Stdout)
		return nil
	}
	if to, ok := newRenamed[args[0]]; ok {
		return fmt.Errorf("`keystone new %s` is now `keystone new %s` (5.0 names)", args[0], to)
	}
	gen, ok := newDispatch[args[0]]
	if !ok {
		return fmt.Errorf("unknown kind %q (use: %s)", args[0], newKinds)
	}
	return gen(args[1:])
}

func printNewUsage(w *os.File) {
	fmt.Fprint(w, `keystone new — scaffold a new file at the conventional path

Usage:

  Constraints — limit, gate, or check what the agent does:
    keystone new rule <topic>/<name>      [--dir <path>]   # a rule (+ paired reference); paths: scopes it
    keystone new reference <topic>/<name> [--dir <path>]   # the reasoning behind a rule
    keystone new hook <id>                [--dir <path>]   # a check that runs on an event and can block
    keystone new permission <id>          [--dir <path>]   # allow / ask / deny tool rules
    keystone new eval <id>                [--dir <path>]   # a test that the charter changes behavior

  Capabilities — add things the agent can do:
    keystone new skill <id>               [--dir <path>]   # Agent Skills SKILL.md
    keystone new command <id>             [--dir <path>]   # a unit of work the user asks for by name
    keystone new subagent <id>            [--dir <path>]   # a role the agent delegates to
    keystone new tool <id>                [--dir <path>]   # a callable (transport: mcp | cli | plugin)
    keystone new document <id>            [--dir <path>]   # a governed output template

  Keystone itself:
    keystone new adapter <name>           [--dir <path>]   # a host keystone has no profile for (.agents/hosts/<name>.json)
    keystone new policy <name>            [--dir <path>]

Each generator drops a skeleton at the conventional path with the right
frontmatter, sections, and charter-root-relative cross-references. The
charter root is .agents/.

Ids may use the colon-namespaced form (e.g. keystone:index); the disk
filename normalizes : to -.

Examples:
  keystone new rule process/release        # .agents/rules/process/release.md
                                           # + .agents/references/process/release.md
  keystone new skill keystone:index         # .agents/skills/keystone-index/SKILL.md
  keystone new subagent security-reviewer   # .agents/agents/security-reviewer.md
  keystone new hook go-vet                  # .agents/hooks/go-vet.md
  keystone new command review               # .agents/commands/review.md
  keystone new document adr                 # .agents/documents/adr.md
  keystone new policy acme-policies         # ./acme-policies/ (policy repo skeleton)
`)
}

// parseDirAndCharterRoot is the shared flag parser for in-charter
// generators. Returns (projectDir, charterRoot, remaining-args). The
// charter root is fixed at 2.0; only --dir is honored as a flag.
func parseDirAndCharterRoot(args []string) (projectDir, charterRoot string, remaining []string, err error) {
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
	return absDir, config.DefaultAgentsRoot, remaining, nil
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
