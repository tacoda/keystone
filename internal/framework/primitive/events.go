package primitive

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
