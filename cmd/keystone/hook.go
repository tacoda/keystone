package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

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
	event, phase, command, typ, dir, host string
	dirSet                                bool // --dir given: no upward search
}

// runHookFire selects every primitive subscribed (via `on:`) to the given
// event and acts on it: computational subscribers run their `run:` script
// (in parallel; any non-zero exit blocks), inferential ones are emitted as a
// dispatch manifest for the host orchestrator to spawn. keystone cannot
// invoke an LLM itself, so the inferential side surfaces *what* to dispatch,
// not the result. Shared by `keystone signal fire` and the `hook fire` alias.
func runHookFire(args []string) error {
	err := fireHooks(args)
	if err != nil && hostPhaseArg(args) {
		return exitError{code: 2, err: err} // the host blocks only on exit 2: fail closed
	}
	return err
}

// hostPhaseArg reports whether any argument names a host phase, so an error
// raised before the event is parsed still blocks the host.
func hostPhaseArg(args []string) bool {
	for _, a := range args {
		if primitive.IsHostPhase(a) {
			return true
		}
	}
	return false
}

func fireHooks(args []string) error {
	opts, err := parseHookFire(args)
	if err != nil {
		return err
	}
	absDir, err := charterDir(opts)
	if err != nil {
		return err
	}
	primitives, _, err := primitive.Walk(absDir, config.DefaultAgentsRoot)
	if err != nil {
		return err
	}
	composed, _ := primitive.Compose(primitives)

	comp, inf := selectHooks(composed, opts)
	emitInferentialManifest(inf, opts)
	failures := runComputationalHooks(absDir, comp, opts, hostPayload(opts))
	if failures == 0 {
		printHostResponse(os.Stdout, opts)
		return nil
	}
	return fmt.Errorf("hook fire %s: %d hook(s) failed", opts.event, failures)
}

// charterDir is --dir, or else the nearest directory at or above the cwd
// that holds the charter: a host may run the bridge from a subdirectory.
// A host phase with no charter found is an error (so it blocks), never a
// silent zero-hook pass.
func charterDir(opts hookFireOpts) (string, error) {
	start, err := filepath.Abs(opts.dir)
	if err != nil {
		return "", fmt.Errorf("resolve dir: %w", err)
	}
	if opts.dirSet {
		return start, nil
	}
	for d := start; ; d = filepath.Dir(d) {
		if info, err := os.Stat(filepath.Join(d, config.DefaultAgentsRoot)); err == nil && info.IsDir() {
			return d, nil
		}
		if filepath.Dir(d) == d {
			break
		}
	}
	if primitive.IsHostPhase(opts.event) {
		return "", fmt.Errorf("no %s/ found at or above %s", config.DefaultAgentsRoot, start)
	}
	return start, nil
}

// hostPayload reads the host's event JSON from stdin (tool name, tool input,
// …) when the event is a host phase and stdin is piped, so each hook sees
// what triggered it. Signals fired by keystone itself get no payload.
func hostPayload(opts hookFireOpts) []byte {
	if !primitive.IsHostPhase(opts.event) {
		return nil
	}
	info, err := os.Stdin.Stat()
	if err != nil || info.Mode()&os.ModeCharDevice != 0 {
		return nil
	}
	data, _ := io.ReadAll(os.Stdin)
	return data
}

func parseHookFire(args []string) (hookFireOpts, error) {
	opts := hookFireOpts{dir: "."}
	flags := map[string]*string{
		"--phase": &opts.phase, "--command": &opts.command,
		"--type": &opts.typ, "--dir": &opts.dir, "--host": &opts.host,
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
			opts.dirSet = opts.dirSet || a == "--dir"
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

// hookResult is one computational hook's outcome.
type hookResult struct {
	id  string
	out string
	err error
}

// runComputationalHooks runs every hook's `run:` script concurrently, passing
// the fire context as env vars and the host payload (if any) on stdin.
// Returns the count of non-zero exits.
func runComputationalHooks(absDir string, hooks []primitive.Primitive, opts hookFireOpts, payload []byte) int {
	results := make([]hookResult, len(hooks))
	var wg sync.WaitGroup
	for i, h := range hooks {
		wg.Add(1)
		go func(i int, h primitive.Primitive) {
			defer wg.Done()
			results[i] = runOneHook(absDir, h, opts, payload)
		}(i, h)
	}
	wg.Wait()

	failures := 0
	for _, r := range results {
		if r.err != nil {
			failures++
			fmt.Fprintf(os.Stderr, "✗ hook %s: %v\n%s\n", r.id, r.err, r.out)
		} else {
			fmt.Fprintf(humanOut(opts), "✓ hook %s\n", r.id)
		}
	}
	return failures
}

// runOneHook runs h's `run:` with the fire context in its env and the host
// payload on stdin.
func runOneHook(absDir string, h primitive.Primitive, opts hookFireOpts, payload []byte) hookResult {
	ctx, cancel := hookContext(opts)
	defer cancel()
	cmd := exec.CommandContext(ctx, "bash", "-c", h.Run)
	cmd.WaitDelay = time.Second // a killed bash's children may hold the output pipe open
	cmd.Dir = absDir
	cmd.Env = append(os.Environ(), hookEnv(opts)...)
	if payload != nil {
		cmd.Stdin = bytes.NewReader(payload)
	}
	out, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		err = fmt.Errorf("timed out after %s", hostHookTimeout)
	}
	return hookResult{id: h.ID, out: strings.TrimSpace(string(out)), err: err}
}

// hostHookTimeout ends a host-phase hook before the host's own timeout
// (Claude Code: 60s), which the host treats as a non-blocking error. A hit
// counts as a failure, so it blocks. ponytail: fixed; make it per-hook if a
// real check needs longer.
const hostHookTimeout = 50 * time.Second

// hookContext bounds a host-phase hook; keystone-fired signals run unbounded.
func hookContext(opts hookFireOpts) (context.Context, context.CancelFunc) {
	if primitive.IsHostPhase(opts.event) {
		return context.WithTimeout(context.Background(), hostHookTimeout)
	}
	return context.WithCancel(context.Background())
}

// humanOut is where hook fire writes for people. Called by a host bridge
// (--host), that is stderr: hosts read stdout as the hook's answer (Cursor
// and Codex parse it as JSON, Claude Code adds a prompt hook's stdout to
// the context).
func humanOut(opts hookFireOpts) io.Writer {
	if opts.host != "" {
		return os.Stderr
	}
	return os.Stdout
}

// hostAllow is the JSON a host expects on stdout from a passing hook, by
// host and Claude Code event name; "" means print nothing. Cursor blocks a
// permission event on missing or invalid JSON; Codex needs JSON from Stop.
func hostAllow(host, event string) string {
	switch host {
	case "cursor":
		switch event {
		case "PreToolUse":
			return `{"permission":"allow"}`
		case "UserPromptSubmit":
			return `{"continue":true}`
		}
		return `{}`
	case "codex":
		return `{}`
	}
	return ""
}

func printHostResponse(w io.Writer, opts hookFireOpts) {
	if out := hostAllow(opts.host, opts.event); out != "" {
		fmt.Fprintln(w, out)
	}
}

func hookEnv(opts hookFireOpts) []string {
	return []string{
		"KEYSTONE_EVENT=" + opts.event,
		"KEYSTONE_PHASE=" + opts.phase,
		"KEYSTONE_COMMAND=" + opts.command,
		"KEYSTONE_TYPE=" + opts.typ,
		"KEYSTONE_HOST=" + opts.host, // which host's payload shape is on stdin
	}
}

// emitInferentialManifest prints the agents the host should spawn for this
// event, with the structured-result schema each must return. keystone selects
// and parallelizes the intent; the host runs the LLM.
func emitInferentialManifest(hooks []primitive.Primitive, opts hookFireOpts) {
	if len(hooks) == 0 {
		return
	}
	w := humanOut(opts)
	fmt.Fprintf(w, "dispatch (%s) — spawn these agents in parallel:\n", opts.event)
	for _, h := range hooks {
		// A hook names its agent via `agent:`; an inferential sensor IS the
		// agent (its body projects to .claude/agents/<id>), so it dispatches
		// itself by id.
		agent := h.Agent
		if agent == "" {
			agent = h.ID
		}
		fmt.Fprintf(w, "  - agent: %s  (%s %s, returns: %s)\n", agent, h.Kind, h.ID, h.Returns)
	}
}

func printHookUsage(w *os.File) {
	fmt.Fprint(w, `keystone hook — fire framework hooks

Usage:

    keystone hook fire <event> [--phase X] [--command Y] [--type Z] [--dir D]

Dispatches every primitive subscribed (`+"`on:`"+`) to <event>. Computational
subscribers run their `+"`run:`"+` script in parallel (any non-zero exit blocks);
inferential ones are listed as a dispatch manifest for the host to spawn.

<event> is a keystone signal (framework event) or a host phase. Signals are
open + extensible — see `+"`keystone signal list`"+` for the built-ins and any
project-declared ones. `+"`keystone signal fire`"+` is the preferred spelling;
this verb is kept as an alias. Host phases (PreToolUse, …) reach this command
through the single settings.json bridge.
`)
}
