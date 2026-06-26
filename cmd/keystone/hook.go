package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// runHook dispatches `keystone hook <sub>`. Today the only subcommand is
// `fire` — the single entry point that dispatches framework hooks.
func runHook(args []string) error {
	if len(args) == 0 {
		printHookUsage(os.Stderr)
		return fmt.Errorf("`keystone hook` requires a subcommand (fire)")
	}
	switch args[0] {
	case "help", "--help", "-h":
		printHookUsage(os.Stdout)
		return nil
	case "fire":
		if len(args) >= 2 && isHelpFlag(args[1]) {
			printHookUsage(os.Stdout)
			return nil
		}
		return runHookFire(args[1:])
	default:
		return fmt.Errorf("unknown hook subcommand %q (use: fire)", args[0])
	}
}

func isHelpFlag(s string) bool { return s == "--help" || s == "-h" }

// hookFireOpts is the parsed invocation: the event plus the optional context
// (phase / command / type) passed through to each hook as env vars.
type hookFireOpts struct {
	event, phase, command, typ, dir string
}

// runHookFire selects every `kind: hook` bound to the given event and acts on
// it: computational hooks run their `run:` script (in parallel; any non-zero
// exit blocks), inferential hooks are emitted as a dispatch manifest for the
// host orchestrator to spawn. keystone cannot invoke an LLM itself, so the
// inferential side surfaces *what* to dispatch, not the result.
func runHookFire(args []string) error {
	opts, err := parseHookFire(args)
	if err != nil {
		return err
	}
	absDir, err := filepath.Abs(opts.dir)
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}
	primitives, _, err := primitive.Walk(absDir, config.DefaultHarnessRoot)
	if err != nil {
		return err
	}
	composed, _ := primitive.Compose(primitives)

	comp, inf := selectHooks(composed, opts)
	emitInferentialManifest(inf, opts)
	failures := runComputationalHooks(absDir, comp, opts)
	if failures > 0 {
		return fmt.Errorf("hook fire %s: %d hook(s) failed", opts.event, failures)
	}
	return nil
}

func parseHookFire(args []string) (hookFireOpts, error) {
	opts := hookFireOpts{dir: "."}
	flags := map[string]*string{
		"--phase": &opts.phase, "--command": &opts.command,
		"--type": &opts.typ, "--dir": &opts.dir,
	}
	var positionals []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		dst, isFlag := flags[a]
		switch {
		case isFlag:
			if i+1 >= len(args) {
				return opts, fmt.Errorf("flag %s requires a value", a)
			}
			*dst = args[i+1]
			i++
		case strings.HasPrefix(a, "-"):
			return opts, fmt.Errorf("unknown flag %s", a)
		default:
			positionals = append(positionals, a)
		}
	}
	if len(positionals) != 1 {
		return opts, fmt.Errorf("`keystone hook fire` requires exactly one <event>")
	}
	opts.event = positionals[0]
	return opts, nil
}

// selectHooks splits the primitives that deterministically fire on the given
// event into computational (run a `run:` script) and inferential (dispatch an
// agent). Any kind that fires — a hook, a computational guide/sensor, an
// inferential sensor — is selected via primitive.HookFire.
func selectHooks(primitives []primitive.Primitive, opts hookFireOpts) (comp, inf []primitive.Primitive) {
	for _, p := range primitives {
		event, computational, ok := primitive.HookFire(p)
		if !ok || event != opts.event {
			continue
		}
		if computational {
			comp = append(comp, p)
		} else {
			inf = append(inf, p)
		}
	}
	return comp, inf
}

// runComputationalHooks runs every hook's `run:` script concurrently, passing
// the fire context as env vars. Returns the count of non-zero exits.
func runComputationalHooks(absDir string, hooks []primitive.Primitive, opts hookFireOpts) int {
	type result struct {
		id  string
		out string
		err error
	}
	results := make([]result, len(hooks))
	var wg sync.WaitGroup
	for i, h := range hooks {
		wg.Add(1)
		go func(i int, h primitive.Primitive) {
			defer wg.Done()
			cmd := exec.Command("bash", "-c", h.Run)
			cmd.Dir = absDir
			cmd.Env = append(os.Environ(), hookEnv(opts)...)
			out, err := cmd.CombinedOutput()
			results[i] = result{id: h.ID, out: strings.TrimSpace(string(out)), err: err}
		}(i, h)
	}
	wg.Wait()

	failures := 0
	for _, r := range results {
		if r.err != nil {
			failures++
			fmt.Fprintf(os.Stderr, "✗ hook %s: %v\n%s\n", r.id, r.err, r.out)
		} else {
			fmt.Fprintf(os.Stdout, "✓ hook %s\n", r.id)
		}
	}
	return failures
}

func hookEnv(opts hookFireOpts) []string {
	return []string{
		"KEYSTONE_EVENT=" + opts.event,
		"KEYSTONE_PHASE=" + opts.phase,
		"KEYSTONE_COMMAND=" + opts.command,
		"KEYSTONE_TYPE=" + opts.typ,
	}
}

// emitInferentialManifest prints the agents the host should spawn for this
// event, with the structured-result schema each must return. keystone selects
// and parallelizes the intent; the host runs the LLM.
func emitInferentialManifest(hooks []primitive.Primitive, opts hookFireOpts) {
	if len(hooks) == 0 {
		return
	}
	fmt.Fprintf(os.Stdout, "dispatch (%s) — spawn these agents in parallel:\n", opts.event)
	for _, h := range hooks {
		// A hook names its agent via `agent:`; an inferential sensor IS the
		// agent (its body projects to .claude/agents/<id>), so it dispatches
		// itself by id.
		agent := h.Agent
		if agent == "" {
			agent = h.ID
		}
		fmt.Fprintf(os.Stdout, "  - agent: %s  (%s %s, returns: %s)\n", agent, h.Kind, h.ID, h.Returns)
	}
}

func printHookUsage(w *os.File) {
	fmt.Fprint(w, `keystone hook — fire framework hooks

Usage:

    keystone hook fire <event> [--phase X] [--command Y] [--type Z] [--dir D]

Dispatches every `+"`kind: hook`"+` bound to <event>. Computational hooks run
their `+"`run:`"+` script in parallel (any non-zero exit blocks); inferential
hooks are listed as a dispatch manifest for the host to spawn.

Framework events: pre-command, post-command, pre-playbook, post-playbook,
on-gate, pre-verify, post-verify, on-phase-enter, on-phase-exit. Host phases
(PreToolUse, …) reach this command through the single settings.json bridge.
`)
}
