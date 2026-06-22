// Package agnostic projects the agent-agnostic surface — the files
// every non-Claude coding agent reads regardless of host (Aider,
// Cline, Cursor, OpenCode, Roo, generic AGENTS.md readers).
//
// Single output: AGENTS.md at the repo root. Content mirrors the
// keystone-managed section of CLAUDE.md so a user's hand-edited
// CLAUDE.md still works for Claude Code while AGENTS.md serves
// every other host.
//
// AGENTS.md is unconditional — every `keystone project` run writes
// it. Unlike the per-host adapter outputs (.cursor/, .aider.*) it
// doesn't require `adapters:` opt-in.
package agnostic

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

// AgentsMDRelPath is the root-level file generic agents read.
const AgentsMDRelPath = "AGENTS.md"

// ProjectionResult records what happened during ProjectAgentsMD.
type ProjectionResult struct {
	Path  string
	Wrote bool
}

// ProjectAgentsMD writes AGENTS.md at the repo root. content is the
// body — typically the same paragraph block CLAUDE.md uses, plus the
// keystone harness pointer. The projector is idempotent: re-running
// with the same content is a no-op.
func ProjectAgentsMD(projectDir, content string) (ProjectionResult, error) {
	abs := filepath.Join(projectDir, AgentsMDRelPath)
	body := []byte(content)
	if len(body) == 0 || body[len(body)-1] != '\n' {
		body = append(body, '\n')
	}
	prev, _ := os.ReadFile(abs)
	if bytes.Equal(prev, body) {
		return ProjectionResult{Path: AgentsMDRelPath, Wrote: false}, nil
	}
	if err := atomicWrite(abs, body); err != nil {
		return ProjectionResult{Path: AgentsMDRelPath}, err
	}
	return ProjectionResult{Path: AgentsMDRelPath, Wrote: true}, nil
}

// DefaultBody returns the canonical AGENTS.md content. Mirrors the
// keystone-managed CLAUDE.md section so a generic agent gets the same
// orientation a Claude Code agent does.
func DefaultBody() string {
	return `# Agent orientation

This project uses a **keystone harness** — an agent-agnostic framework
for guides, sensors, actions, playbooks, and personas. Every
host-native surface (Claude Code skills, Cursor rules, Aider
conventions) is projected from the canonical sources under
` + "`.keystone/harness/`" + `.

## Read first

` + "`.keystone/INDEX.lite.json`" + ` — cheap discovery (kind + id +
description per primitive). Browse this to pick what you need; open
the full ` + "`.keystone/INDEX.json`" + ` only when you need a primitive's
path, globs, or triggers.

## Activate by kind

- **guide** — touched files match the entry's ` + "`globs`" + ` (or no globs declared).
- **corpus** — a guide's ` + "`traces:`" + ` (or a prose forward-link) points at it.
- **action** — user intent matches ` + "`description`" + `; the body is the playbook.
- **playbook** — composed sequence of actions.
- **sensor** — fires under an action's phase, narrowed by ` + "`globs`" + `.
- **persona** — spawned as a subagent for narrow review/scout work.

## Iron laws (non-negotiable)

- No proceeding without explicit acceptance criteria.
- No completion claims without fresh verification — sensors run this turn.
- No commits with failing sensors. Never ` + "`--no-verify`" + `.
- No AI attribution in commits, PRs, or tracker comments.
- No silent overwrites of state files.
- No reading or writing sensitive files (` + "`.env*`" + `, ` + "`*.pem`" + `, ` + "`credentials.json`" + `, secrets dirs).
- No dangerous action without explicit in-turn confirmation (` + "`rm -rf`" + `, force-push, prod DB writes, external comms).
- No invented imports, methods, config keys, or CLI flags — grep first.
- No "while I'm here" cleanups — every changed line traces to the request.
- No accepting a subagent's "done" report as evidence — read the diff.

## Lifecycle

To start a unit of work, say **"run task on ` + "`<ticket-id>`" + `"** — runs the
` + "`task`" + ` playbook. For one action, ask in natural language ("run verify",
"do a review pass"). The action's body lives at its INDEX ` + "`path`" + `.

## Override

Project files at ` + "`.keystone/harness/<kind>/<id>.md`" + ` always win. Among
installed policies, deeper nesting refines outer policies; ` + "`strict:`" + ` is
absolute.
`
}

func atomicWrite(destAbs string, contents []byte) error {
	if err := os.MkdirAll(filepath.Dir(destAbs), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(destAbs), err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(destAbs), ".keystone-agnostic.*")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(contents); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("close temp: %w", err)
	}
	if err := os.Rename(tmpName, destAbs); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("rename %s -> %s: %w", tmpName, destAbs, err)
	}
	return nil
}
