# Port: Flywheel sink

**Activation:** Write-only. Lifecycle actions write into these directories — `learn` files candidates into `learning/inbox/`, `audit` moves retired content into `archive/`.
**Purpose:** Storage destinations for the Learning and Pruning flywheels. The Learning flywheel files new rule candidates; the Pruning flywheel removes retired guides and corpus.

## Path convention

```
harness/learning/inbox/<date>-<slug>.md
harness/learning/triage.md
harness/archive/<date>-<slug>.md
```

These directories are **always project-owned**. Plugins cannot ship flywheel sinks — sinks are project state, not policy.

## Required shape

- **`learning/inbox/<entry>.md`** — written by `learn`.
  ```markdown
  ---
  status: new | reviewed | accepted | rejected
  ---

  # Candidate rule: <one-line summary>

  <proposed rule>

  ## Evidence
  <observations that prompted the candidate>
  ```

- **`learning/triage.md`** — index. Lists inbox entries by status with one-line summaries. Maintained by the agent during `learn`.

- **`archive/<entry>.md`** — written by `audit` when a guide or corpus file is retired.
  ```markdown
  ---
  archived_from: <original-path>
  reason: <one-sentence reason for retirement>
  ---

  <the original content, preserved>
  ```

## Cascade behavior

**Not part of the cascade.** These are write-only sinks; nothing reads them at cascade-resolution time. They exist to capture the flywheel's output for human review.

## Drift behavior

`keystone doctor` reports:
- inbox entries older than a configurable threshold still in `status: new` (triage backlog).
- archive entries whose `archived_from` path still exists elsewhere in the cascade (incomplete retirement).

## Example

```markdown
---
status: new
---

# Candidate rule: branch names follow `<type>/<ticket>-<slug>`

Git branches use a prefix (`feat/`, `fix/`, `chore/`) followed by a tracker ID
and a kebab-case slug. Observed three times this week; no existing guide.

## Evidence
- PR #412 — branch `add-thing` (no prefix); reviewer requested rename.
- PR #418 — branch `johnson/login` (author-prefixed); inconsistent with #412.
- PR #421 — followed the proposed pattern; merged without comment.
```

## Authoring

Sinks are written by actions, not authored by users directly. No `keystone new` generator.
