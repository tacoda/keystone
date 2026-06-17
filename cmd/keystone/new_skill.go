package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// runNewEval handles `keystone new eval <id>`. Scaffolds
// <harness>/evals/<id>/EVAL.md + sibling expected.json stub. The
// `id` becomes both the disk dir name and the frontmatter id.
func runNewEval(args []string) error {
	projectDir, harnessRoot, remaining, err := parseDirAndHarnessRoot(args)
	if err != nil {
		return err
	}
	if len(remaining) != 1 {
		return fmt.Errorf("`keystone new eval` requires exactly one argument: <id>")
	}
	id := remaining[0]
	if err := validatePrimitiveID(id); err != nil {
		return fmt.Errorf("eval id: %w", err)
	}
	diskName := strings.ReplaceAll(id, ":", "-")
	dir := filepath.Join(projectDir, harnessRoot, "evals", diskName)
	evalPath := filepath.Join(dir, "EVAL.md")
	expPath := filepath.Join(dir, "expected.json")

	evalBody := fmt.Sprintf(`---
kind: eval
id: %s
description: TODO — what this eval measures and what counts as a pass.
level: static
---

# %s

Scenario prose. Describe the change the eval is measuring against —
the touched-files context, the expected agent behavior, the assertion
the run will check.
`, id, id)

	expBody := `{
  "touched_files": [
    "src/example.go"
  ],
  "static": {
    "rules_fired":  [],
    "rules_silent": []
  },
  "sensors": []
}
`
	if err := writeSkeleton(evalPath, evalBody); err != nil {
		return err
	}
	if err := writeSkeleton(expPath, expBody); err != nil {
		return err
	}
	return nil
}

// runNewPersona handles `keystone new persona <id>`. Scaffolds
// <harness>/personas/<id>.md with canonical primitive frontmatter.
//
// Persona is the framework wrapper for subagent: it projects to
// .claude/agents/<id>.md via `keystone project`. Same id as a
// hand-written subagent under harness/agents/ is a lint error
// (projection collision).
func runNewPersona(args []string) error {
	projectDir, harnessRoot, remaining, err := parseDirAndHarnessRoot(args)
	if err != nil {
		return err
	}
	if len(remaining) != 1 {
		return fmt.Errorf("`keystone new persona` requires exactly one argument: <id>")
	}
	id := remaining[0]
	if err := validatePrimitiveID(id); err != nil {
		return fmt.Errorf("persona id: %w", err)
	}
	diskName := strings.ReplaceAll(id, ":", "-")
	path := filepath.Join(projectDir, harnessRoot, "personas", diskName+".md")
	body := fmt.Sprintf(`---
kind: persona
id: %s
description: TODO — one-line description of what this persona reviews / does.
tools:
  - Read
  - Grep
---

# %s

System prompt this persona runs under as a delegated subagent. Describe
posture, what it prioritizes, what it ignores, and what it returns.
Keep it focused — one job per persona.

## Output

What the persona returns to its caller. Be explicit about format
(JSON, markdown table, plain text) and length expectations.
`, id, id)
	return writeSkeleton(path, body)
}

// runNewSource handles `keystone new source <id>`. Scaffolds
// <harness>/sources/<id>.md declaring an external source for the
// stage-3 resolution flow.
func runNewSource(args []string) error {
	projectDir, harnessRoot, remaining, err := parseDirAndHarnessRoot(args)
	if err != nil {
		return err
	}
	if len(remaining) != 1 {
		return fmt.Errorf("`keystone new source` requires exactly one argument: <id>")
	}
	id := remaining[0]
	if err := validatePrimitiveID(id); err != nil {
		return fmt.Errorf("source id: %w", err)
	}
	path := filepath.Join(projectDir, harnessRoot, "sources", id+".md")
	body := fmt.Sprintf(`---
kind: source
id: %s
description: TODO — what's behind this source and when to query it.
type: folder
settings:
  path: ./docs
---

# %s

Prose describing ownership, auth, and the kinds of queries this source
answers well. Stage 3 of the runtime resolution flow reaches sources
when in-harness rules + corpus aren't enough.
`, id, id)
	return writeSkeleton(path, body)
}

// runNewRule handles `keystone new rule <id>`. Scaffolds
// <harness-root>/rules/<id>.md with canonical primitive frontmatter for
// an agent-side rule — the host-native flavor (Cursor-style, plain
// CLAUDE.md directive) that lives alongside, not instead of, framework
// guides.
//
// Reach for `keystone new rule` only when extending what the host
// already understands. The default authoring path is `keystone new
// guide`, which produces a framework guide with tiers, paired corpus,
// and severity.
func runNewRule(args []string) error {
	projectDir, harnessRoot, remaining, err := parseDirAndHarnessRoot(args)
	if err != nil {
		return err
	}
	if len(remaining) != 1 {
		return fmt.Errorf("`keystone new rule` requires exactly one argument: <id>")
	}
	id := remaining[0]
	if err := validatePrimitiveID(id); err != nil {
		return fmt.Errorf("rule id: %w", err)
	}
	diskName := strings.ReplaceAll(id, ":", "-")
	path := filepath.Join(projectDir, harnessRoot, "rules", diskName+".md")
	body := fmt.Sprintf(`---
kind: rule
id: %s
description: TODO — one-line description of what this rule directs.
---

# %s

Plain-rule body. One-paragraph framing followed by directives.

- Do this.
- Don't do that.
`, id, id)
	return writeSkeleton(path, body)
}

// runNewSkill handles `keystone new skill <id>`. Scaffolds
// <harness-root>/skills/<dir>/SKILL.md with canonical primitive
// frontmatter.
//
// The id may use the canonical colon-namespaced form
// (`keystone:index`); on disk the dir-name normalizes to hyphens
// (`keystone-index`) for cross-platform safety. The frontmatter
// records the colon form as `id:` so INDEX.json reports it
// unchanged.
func runNewSkill(args []string) error {
	projectDir, harnessRoot, remaining, err := parseDirAndHarnessRoot(args)
	if err != nil {
		return err
	}
	if len(remaining) != 1 {
		return fmt.Errorf("`keystone new skill` requires exactly one argument: <id>")
	}
	id := remaining[0]
	if err := validatePrimitiveID(id); err != nil {
		return fmt.Errorf("skill id: %w", err)
	}
	diskName := strings.ReplaceAll(id, ":", "-")
	path := filepath.Join(projectDir, harnessRoot, "skills", diskName, "SKILL.md")
	body := fmt.Sprintf(`---
kind: skill
id: %s
description: TODO — one-line description; surfaced in INDEX.json.
triggers:
  - %s
  - /%s
---

# %s

One-paragraph framing. What does this skill teach the agent to do, and
when should the agent fire it?

## Run

If the skill wraps a shell command, show the exact invocation:

`+"```\n<command>\n```"+`

## When to trigger

- Bullet one.
- Bullet two.
`, id, id, id, id)
	return writeSkeleton(path, body)
}

// runNewSubagent handles `keystone new subagent <id>`. Scaffolds
// <harness-root>/agents/<id>.md with canonical primitive frontmatter.
func runNewSubagent(args []string) error {
	projectDir, harnessRoot, remaining, err := parseDirAndHarnessRoot(args)
	if err != nil {
		return err
	}
	if len(remaining) != 1 {
		return fmt.Errorf("`keystone new subagent` requires exactly one argument: <id>")
	}
	id := remaining[0]
	if err := validatePrimitiveID(id); err != nil {
		return fmt.Errorf("subagent id: %w", err)
	}
	diskName := strings.ReplaceAll(id, ":", "-")
	path := filepath.Join(projectDir, harnessRoot, "agents", diskName+".md")
	body := fmt.Sprintf(`---
kind: subagent
id: %s
description: TODO — one-line description of what this subagent does.
tools:
  - Read
  - Grep
---

# %s

The system prompt the subagent runs under. Describe its job, its
constraints, what it returns. Keep it focused — one job per subagent.

## Output

What the subagent returns to its caller. Be explicit about format
(JSON, markdown table, plain text) and length expectations.
`, id, id)
	return writeSkeleton(path, body)
}

// runNewCommand handles `keystone new command <id>`. Scaffolds
// <harness-root>/commands/<id>.md with canonical primitive frontmatter.
func runNewCommand(args []string) error {
	projectDir, harnessRoot, remaining, err := parseDirAndHarnessRoot(args)
	if err != nil {
		return err
	}
	if len(remaining) != 1 {
		return fmt.Errorf("`keystone new command` requires exactly one argument: <id>")
	}
	id := remaining[0]
	if err := validatePrimitiveID(id); err != nil {
		return fmt.Errorf("command id: %w", err)
	}
	diskName := strings.ReplaceAll(id, ":", "-")
	path := filepath.Join(projectDir, harnessRoot, "commands", diskName+".md")
	body := fmt.Sprintf(`---
kind: command
id: %s
description: TODO — one-line description; surfaced when the user types /%s.
args:
  - name: target
    type: string
    required: false
    description: TODO — what this argument controls.
---

# /%s

What the command does when invoked. The body becomes the prompt the
host injects into the agent's session.
`, id, id, id)
	return writeSkeleton(path, body)
}

// validatePrimitiveID applies the shared id constraints: non-empty,
// lowercase, no whitespace, no path separators, max 64 chars. Permits
// `:` (namespace) and `-` (kebab); the disk-name normalizer rewrites
// `:` to `-`.
func validatePrimitiveID(id string) error {
	if id == "" {
		return fmt.Errorf("id is empty")
	}
	if len(id) > 64 {
		return fmt.Errorf("id %q exceeds 64 chars", id)
	}
	if strings.ContainsAny(id, "/\\ \t") {
		return fmt.Errorf("id %q must not contain spaces or path separators", id)
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case r == '-' || r == ':' || r == '_':
		default:
			return fmt.Errorf("id %q contains invalid character %q (allowed: a-z 0-9 - : _)", id, r)
		}
	}
	return nil
}
