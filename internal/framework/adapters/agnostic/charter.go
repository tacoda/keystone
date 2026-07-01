package agnostic

import (
	"fmt"
	"path/filepath"
	"strings"
)

// CharterMDRelPath is the single canonical agent-entrypoint file at the
// repo root. `keystone project` writes it once; every host's own file
// (CLAUDE.md, AGENTS.md, CONVENTIONS.md, a Cursor/Continue rule) is a
// thin pointer that directs the agent to read it. The full orientation
// — how to read the index, the activation table, the iron laws, the
// lifecycle — lives here and nowhere else, so there is one source of
// truth for what governs the repo.
const CharterMDRelPath = "CHARTER.md"

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

// Host profiles, one per supported target. Kept as constructors so a
// caller reads `agnostic.ClaudeCodeProfile()` at the call site rather
// than threading a registry.
func ClaudeCodeProfile() HostProfile {
	return HostProfile{Name: "Claude Code", Import: "@CHARTER.md", Subagents: true, SlashCommands: true, SkillsAutoActivate: true, Hooks: true}
}

func OpencodeProfile() HostProfile {
	return HostProfile{Name: "opencode", Subagents: true, SlashCommands: true, SkillsAutoActivate: true}
}

func CursorProfile() HostProfile {
	return HostProfile{Name: "Cursor", SkillsAutoActivate: true}
}

func ContinueProfile() HostProfile {
	return HostProfile{Name: "Continue", SkillsAutoActivate: true}
}

func AiderProfile() HostProfile {
	return HostProfile{Name: "Aider"}
}

// GenericProfile is the conservative default for the shared root
// AGENTS.md, which many hosts read. It claims no host-specific
// capability; a host that has more (subagents, commands) discovers them
// through its own projected dirs regardless. Empty Name → the pointer
// renders a plain "On this host" heading.
func GenericProfile() HostProfile {
	return HostProfile{}
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
		lines = append(lines, "- **Subagents** — spawn charter agents (`.charter/agents/`) as subagents for review/scout work.")
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

// ProjectCharterMD writes the canonical CHARTER.md at the repo root.
// Idempotent: re-running with unchanged content is a no-op.
func ProjectCharterMD(projectDir string) (ProjectionResult, error) {
	abs := filepath.Join(projectDir, CharterMDRelPath)
	body := []byte(CharterBody())
	res, err := writeIfChanged(abs, CharterMDRelPath, body)
	return res, err
}

// CharterBody is the canonical, host-neutral orientation written to
// CHARTER.md. It is the one place the full orientation lives.
func CharterBody() string {
	return `# Charter

This repo is governed by a **keystone charter** — the authored standards
that constrain whatever coding-agent harness runs here (Claude Code,
Cursor, Codex, opencode, …). A *harness* is the engine that runs the
model; the *charter* is what you author to make its output reliable.
Author the spec → charter. Be the engine → harness.

The charter is a tree of typed primitives under ` + "`.charter/`" + `. Every
host-native surface (` + "`.claude/`" + `, ` + "`.cursor/`" + `, this file) is projected
from it — do not hand-edit projected files.

## Read first

` + "`.charter/INDEX.lite.json`" + ` — cheap discovery (kind + id + description
per primitive). Browse it to pick what you need; open the full
` + "`.charter/INDEX.json`" + ` only when you need a primitive's path, globs, or
triggers, and open a body only when its activation condition matches.

## Activate by kind

- **guide** — touched files match the entry's ` + "`globs:`" + ` (or it declares none).
- **corpus** — a guide's ` + "`corpus:`" + ` (or a prose forward-link) points at it — the *why*, on demand.
- **command** — user intent matches ` + "`description`" + ` + ` + "`phase`" + `; the body is a unit of work.
- **playbook** — a composed sequence of commands with human ` + "`gates:`" + `.
- **sensor** — a check that reacts to a signal or host phase (` + "`on:`" + `); computational (` + "`run:`" + ` → status verdict) or inferential (agent review → schema); gates.
- **tool** — an author-defined external callable; on-demand, or a side-effect when it declares ` + "`on:`" + `.
- **skill** — auto-activates by ` + "`triggers:`" + ` match.
- **agent** — a role spawned as a subagent by ` + "`id`" + `; the body is its system prompt.
- **pattern** — a reusable documentation pattern; apply when writing docs.

A **signal** is a keystone framework event (` + "`on:`" + `); host phases are bridged, any other ` + "`on:`" + ` value is a signal (extensible; ` + "`keystone signal list`" + `).

## Iron laws (non-negotiable, every phase)

- No proceeding without explicit acceptance criteria.
- No completion claims without fresh verification — checks run this turn, against post-edit state, with cited output.
- No commits with failing checks. Never ` + "`--no-verify`" + `.
- No AI attribution in commits, PRs, or tracker comments.
- No silent overwrites of state files.
- No reading or writing sensitive files (` + "`.env*`" + `, ` + "`*.pem`" + `, ` + "`*.key`" + `, ` + "`credentials.json`" + `, secrets dirs) — ask out-of-band.
- No dangerous action without explicit in-turn confirmation (` + "`rm -rf`" + `, force-push, ` + "`reset --hard`" + `, prod DB writes, external comms, installs).
- No invented imports, methods, config keys, or CLI flags — grep / read the manifest / check ` + "`--help`" + ` first.
- No "while I'm here" cleanups — every changed line traces to the request.
- No accepting a subagent's "done" report as evidence — read the diff; re-run checks in the parent turn.

## Lifecycle

To start a unit of work, say **"run task on ` + "`<ticket-id>`" + `"** — runs the
` + "`task`" + ` playbook. For one command, ask in natural language ("run verify",
"do a review pass"); its body lives at its INDEX ` + "`path`" + `.

## Override

Project files at ` + "`.charter/<kind>/<id>.md`" + ` always win. Among installed
policies, deeper nesting refines outer policies; a ` + "`strict:`" + ` item is
absolute — nothing else overrides it.
`
}
