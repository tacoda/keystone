package primitive

import "strings"

// FrameworkEvents is the closed set of keystone-internal workflow events a
// `hook` may bind to via `event:`. These fire from keystone's own subcommands
// (and from `keystone hook fire`) — the host cannot see them, which is why the
// hook layer exists. Host-phase events (PreToolUse, PostToolUse, …) are the
// other `event:` namespace; anything not in this set is treated as a host
// phase.
var FrameworkEvents = []string{
	"pre-command", "post-command",
	"pre-playbook", "post-playbook",
	"on-gate",
	"pre-verify", "post-verify",
	"on-phase-enter", "on-phase-exit",
}

// IsFrameworkEvent reports whether s is a keystone framework event (vs a
// host-phase event fired by the host).
func IsFrameworkEvent(s string) bool {
	for _, e := range FrameworkEvents {
		if e == s {
			return true
		}
	}
	return false
}

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
