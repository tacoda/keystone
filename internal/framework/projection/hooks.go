package projection

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tacoda/keystone/internal/framework/adapters/claudecode"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// A hook bridge is one host config entry per host event, each running
// `keystone hook fire <Event> --host <host>`; keystone then runs every hook
// subscribed to that event and exits 2 when one fails (the host's "block").
// Events keep the Claude Code names; each bridge maps them to its host's.

// bridgeCmd is the command every host bridge runs.
func bridgeCmd(phase, host string) string {
	return "keystone hook fire " + phase + " --host " + host
}

// shellBridge is bridgeCmd for hosts that run it through a shell: any other
// failure (binary missing: 127; an older binary rejecting a flag: 1) still
// exits 2, the code hosts block on.
func shellBridge(phase, host string) string { return bridgeCmd(phase, host) + " || exit 2" }

// HookBridge builds a host's bridge output for the subscribed phases, and
// lists the phases the host has no event for (those hooks do not run there).
type HookBridge func(phases []string) (Output, []string)

// hookOutputs plans each host's bridge. A host without a bridge gets a
// Report output, so `keystone project` and AGENTS.md say its hooks do not run.
func hookOutputs(primitives []primitive.Primitive, hosts []Profile) ([]Output, []Warning) {
	phases := claudecode.HostPhases(primitives)
	if len(phases) == 0 {
		return staleBridgeCleanup(hosts), nil
	}
	var outs []Output
	var warns []Warning
	for _, h := range hosts {
		if h.Hooks == nil {
			outs = append(outs, Output{Host: h.Host, Kind: string(hookK), Strategy: Report})
			continue
		}
		o, missing := h.Hooks(phases)
		o.Host, o.Kind = h.Host, string(hookK)
		outs = append(outs, o)
		if len(missing) > 0 {
			warns = append(warns, Warning{Host: h.Host, Message: "no host event for " + strings.Join(missing, ", ") + " — hooks on it do not run here"})
		}
	}
	return outs, append(warns, doubleFireWarning(hosts)...)
}

// staleBridgeCleanup keeps merging with no phases when no hook binds a host
// event, so keystone's old entries leave shared config files (owned bridge
// files are pruned instead). An absent file is left absent.
func staleBridgeCleanup(hosts []Profile) []Output {
	var outs []Output
	for _, h := range hosts {
		if h.Hooks == nil {
			continue
		}
		o, _ := h.Hooks(nil)
		if o.Strategy != Merge {
			continue
		}
		merge := o.MergeFn
		o.MergeFn = func(prev []byte) ([]byte, error) {
			if len(prev) == 0 {
				return prev, nil
			}
			return merge(prev)
		}
		o.Host, o.Kind = h.Host, string(hookK)
		outs = append(outs, o)
	}
	return outs
}

// doubleFireWarning: Copilot also reads .claude/settings.json hooks.
func doubleFireWarning(hosts []Profile) []Warning {
	claude, copilot := false, false
	for _, h := range hosts {
		claude = claude || h.Host == "claude-code"
		copilot = copilot || h.Host == "github-copilot"
	}
	if !claude || !copilot {
		return nil
	}
	return []Warning{{Host: "github-copilot", Message: "Copilot CLI also runs the hooks in .claude/settings.json, so each check runs twice there (harmless for checks, slower)"}}
}

// claudeHooks merges the bridges into .claude/settings.json, keeping every
// user key and hook.
func claudeHooks(phases []string) (Output, []string) {
	return Output{Dest: ".claude/settings.json", Strategy: Merge, MergeFn: func(prev []byte) ([]byte, error) {
		return claudecode.MergeHooks(prev, phases)
	}}, nil
}

// codexEvents are the Claude Code events Codex also fires (developers.openai.com/codex/hooks).
var codexEvents = map[string]string{
	"PreToolUse": "PreToolUse", "PostToolUse": "PostToolUse", "UserPromptSubmit": "UserPromptSubmit",
	"Stop": "Stop", "SubagentStop": "SubagentStop", "PreCompact": "PreCompact",
	"SessionStart": "SessionStart", "SessionEnd": "SessionEnd",
}

// codexHooks merges the bridges into .codex/hooks.json, which has the Claude
// Code shape. Codex runs project hooks only once the user trusts them
// (/hooks), and only when the project .codex/ layer is trusted.
func codexHooks(phases []string) (Output, []string) {
	var supported []string
	for _, p := range phases {
		if _, ok := codexEvents[p]; ok {
			supported = append(supported, p)
		}
	}
	return Output{Dest: ".codex/hooks.json", Strategy: Merge, MergeFn: func(prev []byte) ([]byte, error) {
		return claudecode.MergeHooksFor(prev, supported, "codex")
	}}, missingEvents(phases, codexEvents)
}

var cursorEvents = map[string]string{
	"PreToolUse": "preToolUse", "PostToolUse": "postToolUse", "UserPromptSubmit": "beforeSubmitPrompt",
	"Stop": "stop", "SubagentStop": "subagentStop", "PreCompact": "preCompact",
	"SessionStart": "sessionStart", "SessionEnd": "sessionEnd",
}

// cursorHooks merges into .cursor/hooks.json. failClosed is set because a
// crashing Cursor hook otherwise fails open (the action proceeds).
func cursorHooks(phases []string) (Output, []string) {
	return Output{Dest: ".cursor/hooks.json", Strategy: Merge, MergeFn: func(prev []byte) ([]byte, error) {
		return mergeVersionedHooks(prev, phases, cursorEvents, func(phase string) map[string]any {
			return map[string]any{"command": bridgeCmd(phase, "cursor"), "failClosed": true}
		}, "command")
	}}, missingEvents(phases, cursorEvents)
}

var copilotEvents = map[string]string{
	"PreToolUse": "preToolUse", "PostToolUse": "postToolUse", "UserPromptSubmit": "userPromptSubmitted",
	"Notification": "notification", "Stop": "agentStop", "SubagentStop": "subagentStop",
	"PreCompact": "preCompact", "SessionStart": "sessionStart", "SessionEnd": "sessionEnd",
}

// copilotHooks owns .github/hooks/keystone.json (Copilot reads every file
// in .github/hooks/). A non-zero preToolUse command denies the call.
func copilotHooks(phases []string) (Output, []string) {
	out, _ := mergeVersionedHooks(nil, phases, copilotEvents, func(phase string) map[string]any {
		return map[string]any{"type": "command", "bash": shellBridge(phase, "github-copilot"), "timeoutSec": 120}
	}, "bash")
	return Output{Dest: ".github/hooks/keystone.json", Strategy: Shim, Content: out}, missingEvents(phases, copilotEvents)
}

var opencodeEvents = map[string]string{
	"PreToolUse": "tool.execute.before", "PostToolUse": "tool.execute.after",
	"SessionStart": "session.created", "Stop": "session.idle",
}

// opencodeHooks owns a plugin. Bun's $ throws on any non-zero exit (a
// missing binary too), and a throw in tool.execute.before blocks the tool.
// The event input goes to the bridge on stdin, as the other hosts send it.
func opencodeHooks(phases []string) (Output, []string) {
	var b strings.Builder
	b.WriteString("// Generated by keystone project. Edit .agents/hooks/, not this file.\n")
	b.WriteString("export const KeystoneHooks = async ({ $ }) => ({\n")
	for _, phase := range phases {
		if ev, ok := opencodeEvents[phase]; ok {
			fmt.Fprintf(&b, "  %q: async (input) => { await $`%s < ${new Response(JSON.stringify(input))}` },\n", ev, bridgeCmd(phase, "opencode"))
		}
	}
	b.WriteString("});\n")
	return Output{Dest: ".opencode/plugins/keystone.js", Strategy: Shim, Content: []byte(b.String())}, missingEvents(phases, opencodeEvents)
}

func missingEvents(phases []string, events map[string]string) []string {
	var out []string
	for _, p := range phases {
		if _, ok := events[p]; !ok {
			out = append(out, p)
		}
	}
	return out
}

// mergeVersionedHooks merges bridge entries into a {"version":1,"hooks":
// {event:[entry…]}} file (Cursor, Copilot). Entries whose cmdKey starts
// with "keystone hook fire " are keystone's and are replaced; every other
// entry and key is kept.
func mergeVersionedHooks(prev []byte, phases []string, events map[string]string, entry func(string) map[string]any, cmdKey string) ([]byte, error) {
	doc, err := readJSONObject(prev)
	if err != nil {
		return nil, err
	}
	hooks, _ := doc["hooks"].(map[string]any)
	if hooks == nil {
		hooks = map[string]any{}
	}
	stripKeystoneEntries(hooks, cmdKey)
	appendBridges(hooks, phases, events, entry)
	doc["version"] = 1
	doc["hooks"] = hooks
	return marshalJSON(doc)
}

// readJSONObject parses prev as a JSON object; empty input is an empty one.
func readJSONObject(prev []byte) (map[string]any, error) {
	doc := map[string]any{}
	if len(bytes.TrimSpace(prev)) == 0 {
		return doc, nil
	}
	if err := json.Unmarshal(prev, &doc); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	return doc, nil
}

// appendBridges adds one entry per phase the host has an event for.
func appendBridges(hooks map[string]any, phases []string, events map[string]string, entry func(string) map[string]any) {
	for _, phase := range phases {
		if ev, ok := events[phase]; ok {
			list, _ := hooks[ev].([]any)
			hooks[ev] = append(list, entry(phase))
		}
	}
}

func stripKeystoneEntries(hooks map[string]any, cmdKey string) {
	for ev, listAny := range hooks {
		list, _ := listAny.([]any)
		kept := list[:0]
		for _, e := range list {
			if m, ok := e.(map[string]any); ok && strings.HasPrefix(fmt.Sprint(m[cmdKey]), "keystone hook fire ") {
				continue
			}
			kept = append(kept, e)
		}
		if len(kept) == 0 {
			delete(hooks, ev)
		} else {
			hooks[ev] = kept
		}
	}
}

func marshalJSON(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	err := enc.Encode(v)
	return buf.Bytes(), err
}
