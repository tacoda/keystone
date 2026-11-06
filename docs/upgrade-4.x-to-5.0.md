# Upgrading from 4.x to 5.0

<!-- voice-register: informational -->
<!-- voice-english: ste -->

5.0 splits the charter into two folders and renames its kinds to the names
the rest of the ecosystem uses. Your sources move from `.charter/` to
`.agents/`, keystone's own state moves to `.keystone/`, and `CHARTER.md`
becomes a managed block in `AGENTS.md`. The migration only moves files: it
never rewrites what you wrote. The cost is one conflict that can stop it,
which section 3 shows you how to clear, and a lint pass to adopt the new
spellings when you're ready.

```bash
git add -A && git commit -m "pre-5.0"      # your restore point
go install github.com/tacoda/keystone/cmd/keystone@latest   # or the release binary
keystone migrate up --dry-run              # print every step, write nothing
keystone migrate up
keystone index && keystone project
keystone verify
```

Install 5.0 first: a 4.x binary doesn't know the 5.0 migration. Run the
migration right after, because the 5.0 binary reads `.agents/` and finds
nothing until the files move.

## 1. What moves

The migration moves each `.charter/` entry to its 5.0 home. File content
stays byte for byte the same.

| 4.x                                  | 5.0                          |
| ------------------------------------ | ---------------------------- |
| `.charter/guides/`, `.charter/rules/` | `.agents/rules/`            |
| `.charter/corpus/`                   | `.agents/references/`        |
| `.charter/sensors/`                  | `.agents/hooks/`             |
| `.charter/posture/`                  | `.agents/permissions/`       |
| `.charter/playbooks/<id>.md`         | `.agents/skills/<id>/SKILL.md` |
| `.charter/patterns/<id>.md`          | `.agents/skills/<id>/SKILL.md` |
| `.charter/agents/`                   | `.agents/agents/`            |
| `.charter/skills/`, `commands/`, `tools/`, `documents/`, `evals/` | the same name under `.agents/` |
| `.charter/concerns/`                 | `.agents/concerns/` (retired; see section 4) |
| `.charter/adapters/`                 | `.keystone/hosts/`           |
| `.charter/learning/`, `archive/`, `state/`, `work/`, `sources/`, `policies/` | the same name under `.keystone/` |
| `.charter/lockfile.json`, `INDEX.json`, `INDEX.lite.json` | `.keystone/`   |

It also does four things that aren't moves:

- **Removes the 4.x entrypoints.** It deletes `CHARTER.md`, and an `AGENTS.md`
  or `CONVENTIONS.md` that only points at it. It checks that the file starts
  with the exact text 4.x generated first. A file with your own text in it
  stays.
- **Takes over the 4.x host files.** It records the `keystone-*` files under
  `.claude/` and the other host folders as keystone's. The next `keystone
  project` replaces them. 5.0 projects each primitive under its own name, so
  `.claude/agents/keystone-reviewer.md` becomes `.claude/agents/reviewer.md`.
- **Archives the old scaffold skills.** 5.0 replaces `keystone-new-guide`,
  `-corpus`, `-sensor`, `-playbook`, `-command`, and `-action` with one
  `keystone-new` skill. If you never edited one, it moves to
  `.keystone/archive/5.0/skills/`. If you did, it stays where it is.
- **Carries your `.gitignore`.** Where your `.gitignore` ignored
  `.charter/policies/`, it adds `.keystone/policies/`.

`keystone migrate down` reverses every step.

## 2. Keep one binary on every machine

Host hook configs call the `keystone` on your `PATH`. A 4.x binary looks
for `.charter/` and finds nothing, so your checks don't run. It also rejects
the `--host` flag that the 5.0 bridges pass. Install 5.0 everywhere the repo
runs agents, including CI, before you merge the migration.

The 5.0 bridges fail closed. If `keystone` is missing or fails before the
checks run, Claude Code and Copilot block the tool call instead of letting it
through.

## 3. The conflict that stops the migration

The migration stops before it moves anything when two files map to the same
place, or when `.charter/` holds an entry 5.0 has no home for. It names each
one.

The common case is a repo with both `.charter/guides/README.md` and a 4.x
`.charter/rules/README.md`. Both map to `.agents/rules/README.md`. The
`rules/` one described the old split between guides and rules, which 5.0
removes.

1. Delete `.charter/rules/README.md`.
2. Run `keystone migrate up` again.

For any other conflict, move or merge the named file by hand, then run the
migration again.

## 4. Adopt the 5.0 names

The 4.x names keep working, permanently. I kept the migration to moves only
because a change to your frontmatter is one you should see and approve, and
`keystone lint --fix` shows you each diff first. A 4.x kind in frontmatter, a 4.x
folder in a policy, `globs:`, and `on:` all parse. `keystone lint` warns once
per file with the 5.0 spelling.

| 4.x                 | 5.0                       |
| ------------------- | ------------------------- |
| `kind: guide`       | `kind: rule`              |
| `kind: corpus`      | `kind: reference`         |
| `kind: sensor`      | `kind: hook`              |
| `kind: posture`     | `kind: permission`        |
| `kind: agent`       | `kind: subagent`          |
| `kind: playbook`, `kind: pattern` | `kind: skill` |
| `globs:`            | `paths:`                  |
| `on:`               | `event:`                  |

To apply them:

1. Run `keystone lint --fix`. It shows the diff for each file it changes.
2. Move any skill `triggers:` into its `description`, by hand. Agent Skills
   has no `triggers:` field, and hosts match on the description.

5.0 retires `concern`. A concern's fields merged into the primitives that
included it. Copy those fields into each primitive, then delete the concern.
Lint warns until you do.

## 5. Choose how hosts get skills

Claude Code reads the Agent Skills format but not `.agents/skills/`, so
keystone links each skill into `.claude/skills/`. Continue and Cline get
copies, because they may skip symlinked folders.

If symlinks are a problem for you (Windows without developer mode, a tool
that doesn't follow them), set copies everywhere:

```json
{ "link_mode": "copy" }
```

`keystone verify` then checks that each copy still matches its source.

## 6. What else changed

- **`keystone new`** takes the 5.0 kinds. `keystone new guide` and the other
  4.x names exit with an error that names the replacement.
- **`keystone init --reset`** no longer exists. It deleted `.agents/` before writing
  it again. `keystone init --overwrite` replaces keystone's files in place and
  deletes nothing.
- **Host files:** the lockfile records every file keystone writes. `keystone project` never
  replaces a file it didn't write, unless you pass `--overwrite`.
- **MCP:** `keystone_get_corpus` is now `keystone_get_reference`, and the
  `keystone_new_<kind>` tools use the 5.0 kinds.
- **`keystone show --json`** names the field `paths`, not `globs`.
- **Budgets:** `keystone.json` `budgets` keys are now `always_loaded` or a
  kind name. The 4.x keys (`guides`, `corpus`, …) still work. See `keystone
  budget --help`.

If something here doesn't match what you see, `keystone migrate status` shows
what ran. `keystone migrate down` reverses the migration, and your pre-5.0
commit is the restore point for everything else.
