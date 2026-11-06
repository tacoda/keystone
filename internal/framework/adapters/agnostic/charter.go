package agnostic

import (
	"fmt"
	"strings"
)

// HostProfile describes what a given coding-agent host can do with the
// charter. The thin per-host pointer renders a capability delta from it
// — e.g. Claude Code runs subagents and slash commands; Cursor does
// not. Capabilities never drop the projected primitive files (skills,
// commands, agents still land in each host's native dirs); the profile
// only tells the agent which of them this host can actually invoke.
type HostProfile struct {
	// Name is the human host label shown in the pointer ("Claude Code").
	Name string
	// Import is the host's native file-import directive for CLAUDE.md-
	// style entrypoints (e.g. "@CHARTER.md"). Empty when the host has no
	// import mechanism — the pointer then falls back to an imperative
	// "read ./CHARTER.md" instruction.
	Import string
	// Subagents: the host can spawn charter agents as subagents.
	Subagents bool
	// SlashCommands: charter commands surface as /keystone-<id>.
	SlashCommands bool
	// SkillsAutoActivate: the host auto-activates skills by trigger.
	SkillsAutoActivate bool
	// Hooks: charter hooks fire on this host's lifecycle events.
	Hooks bool
}

// Host profiles kept for the frozen 4.x Cursor/Continue pointer rules.
func CursorProfile() HostProfile {
	return HostProfile{Name: "Cursor", SkillsAutoActivate: true}
}

func ContinueProfile() HostProfile {
	return HostProfile{Name: "Continue", SkillsAutoActivate: true}
}

func AiderProfile() HostProfile {
	return HostProfile{Name: "Aider"}
}

// RenderPointer returns the thin per-host entrypoint body: an imperative
// instruction to load CHARTER.md (so the iron laws and ambient rules
// apply on every host, including those with no file-import), followed by
// the host's capability delta. It carries none of the orientation prose
// itself — that lives only in CHARTER.md.
func RenderPointer(p HostProfile) string {
	var b strings.Builder
	b.WriteString("# Charter\n\n")
	if p.Import != "" {
		// Native import: the host inlines CHARTER.md automatically, but
		// still state the imperative for agents that skim.
		fmt.Fprintf(&b, "%s\n\n", p.Import)
		b.WriteString("You **must** read [`CHARTER.md`](CHARTER.md) before doing anything in this repo — it carries the iron laws and the ambient rules that govern the charter. The import above loads it; do not proceed without it.\n")
	} else {
		b.WriteString("You **must** read [`CHARTER.md`](CHARTER.md) now, before doing anything in this repo. It carries the iron laws and the ambient rules that govern the charter; they apply whether or not this file restates them. Do not proceed without loading it.\n")
	}
	if p.Name == "" {
		b.WriteString("\n## On this host\n\n")
	} else {
		fmt.Fprintf(&b, "\n## On this host — %s\n\n", p.Name)
	}
	b.WriteString(renderCapabilities(p))
	return b.String()
}

// renderCapabilities turns a HostProfile into the capability-delta bullet
// list. Every line is phrased so the agent knows what it can and cannot
// reach here — the charter primitives are always projected; the profile
// only scopes which are invocable.
func renderCapabilities(p HostProfile) string {
	var lines []string
	if p.Subagents {
		lines = append(lines, "- **Subagents** — spawn charter agents (`.agents/agents/`) as subagents for review/scout work.")
	} else {
		lines = append(lines, "- **No subagents** — run agent-defined reviews inline yourself; the `agents/` bodies are still the instructions.")
	}
	if p.SlashCommands {
		lines = append(lines, "- **Slash commands** — charter commands and playbooks surface as `/keystone-<id>`.")
	} else {
		lines = append(lines, "- **No slash commands** — invoke a command/playbook by opening its body and following the steps.")
	}
	if p.SkillsAutoActivate {
		lines = append(lines, "- **Skills** — auto-activate by their `triggers:`.")
	}
	if p.Hooks {
		lines = append(lines, "- **Hooks** — charter hooks fire automatically on this host's lifecycle events.")
	} else {
		lines = append(lines, "- **No automatic hooks** — run the checks a hook would fire (lint, type-check, test, build) yourself before claiming done.")
	}
	return strings.Join(lines, "\n") + "\n"
}
