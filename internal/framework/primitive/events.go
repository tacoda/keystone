package primitive

import "strings"

// HostPhases is the CLOSED set of host-native hook phases — the lifecycle
// points the coding-agent host (Claude Code, …) fires itself and can see.
// A `hook` bound to one of these is bridged into the host's settings.json.
// This set is closed because it mirrors what the host actually emits.
//
// Everything else a hook binds to is a SIGNAL (see IsSignal): a keystone
// framework event the host cannot see, which is why the hook layer exists.
// Signals are OPEN — the set below is only the built-ins keystone fires; a
// project may declare and fire its own (keystone.json `signals:` +
// `keystone signal fire <name>`).
var HostPhases = []string{
	"PreToolUse", "PostToolUse", "UserPromptSubmit", "Notification",
	"Stop", "SubagentStop", "PreCompact", "SessionStart", "SessionEnd",
}

// BuiltinSignals are the keystone framework events fired from keystone's own
// subcommands (and via `keystone signal fire`). Projects extend this with
// custom signals; any non-host-phase event name is treated as a signal, so
// custom names work without registration — declaration is for lint/discovery.
var BuiltinSignals = []string{
	"pre-command", "post-command",
	"pre-playbook", "post-playbook",
	"pre-task", "post-task",
	"on-gate",
	"pre-verify", "post-verify",
	"on-phase-enter", "on-phase-exit",
	"on-spec", "on-plan", "on-review", "on-commit",
	"on-learn", "on-synthesize", "on-audit", "on-mode-change",
}

// IsHostPhase reports whether s is a host-native hook phase (bridged into
// the host) rather than a keystone signal.
func IsHostPhase(s string) bool {
	for _, p := range HostPhases {
		if p == s {
			return true
		}
	}
	return false
}

// IsSignal reports whether s is a keystone signal — a framework event the
// host cannot see. The set is OPEN: any non-empty event that is not a host
// phase is a signal, so projects can define their own.
func IsSignal(s string) bool {
	return strings.TrimSpace(s) != "" && !IsHostPhase(s)
}

// IsBuiltinSignal reports whether s is one of keystone's shipped signals
// (vs a project-defined custom one). Used for discovery, not gating.
func IsBuiltinSignal(s string) bool {
	for _, e := range BuiltinSignals {
		if e == s {
			return true
		}
	}
	return false
}

// IsFrameworkEvent is a deprecated alias for IsSignal, kept so existing
// callers keep compiling during the 4.0 rename.
//
// Deprecated: use IsSignal.
func IsFrameworkEvent(s string) bool { return IsSignal(s) }

// HookFire classifies how a primitive deterministically fires at its `event:`,
// unifying hooks and computational guides/sensors into one reliable layer.
// Anything with an event + run is a computational fire (a shell command);
// anything with an event + an agent target is an inferential dispatch. A
// `sensor` with an event but no run is an inferential review dispatched as its
// own projected agent. ok=false for primitives that don't event-fire (e.g. an
// inferential guide, which is a glob-activated rule shim, not a hook).
func HookFire(p Primitive) (event string, computational, ok bool) {
	event = strings.TrimSpace(p.Event)
	if event == "" {
		return "", false, false
	}
	if strings.TrimSpace(p.Run) != "" {
		return event, true, true
	}
	if strings.TrimSpace(p.Agent) != "" || Kind(p.Kind) == KindSensor {
		return event, false, true
	}
	return "", false, false
}
