# Flywheel sinks

<!-- voice-register: informational -->
<!-- voice-english: ste -->

The flywheel sinks are where the two charter flywheels store their output. The learning flywheel files rule candidates in `.keystone/learning/`, and the pruning flywheel moves retired primitives to `.keystone/archive/`. Only commands write to these folders, and the agent doesn't read them during normal work. You review what lands there.

## Path convention

```
.keystone/learning/inbox/<timestamp>-<slug>.md
.keystone/learning/promoted/
.keystone/learning/rejected/
.keystone/learning/wishlist.md
.keystone/archive/<original path under .agents/>
```

Your project always owns these folders. A policy can't ship a sink, because a sink is project state.

## Who writes what

- **`learn`** writes a candidate to `learning/inbox/`.
- **`synthesize`** classifies each candidate as a rule, a reference, or both. It promotes an accepted candidate into `.agents/rules/` or `.agents/references/`, then moves the inbox file to `learning/promoted/`. A rejected candidate moves to `learning/rejected/` with the reason.
- **`audit`** walks the inbox and finds stale rules and references. It moves each retired file to `archive/`, then runs `keystone index`.
- **`wishlist.md`** is a list of known gaps that the team keeps by hand.

## Required shape

An inbox entry, as `learn` writes it:

```markdown
---
captured_at: <ISO datetime>
captured_by: <command / agent>
confidence: <high | medium | low>
candidate_kind: <rule | information | both>
suggested_target: <rules/... | references/...>
---

# <One-line title>

## What happened
## The judgment call
## Suggested rule (if any)
## Suggested information (if any)
```

An archived file keeps its original content under a header that `audit` adds:

```markdown
---
archived_at: <ISO date>
archived_by: <command / human>
reason: <factually-wrong | aspirationally-stale | domain-stale | process-stale | strategy-shift>
replaced_by: <path to new file, or "(none)">
---
```

The archive mirrors the original path under `.agents/`. For the full format, read `.keystone/learning/README.md` and `.keystone/archive/README.md`.

## Cascade behavior

The sinks aren't part of the cascade, so nothing reads them when keystone resolves the charter. Keystone doesn't index them either, because they sit outside `.agents/`.

## Archive written by migration

`keystone migrate` also writes to the archive. Its `v5_0` step moves each unedited 4.x scaffold skill to `.keystone/archive/5.0/skills/`. Those skills are `keystone-new-guide`, `-corpus`, `-sensor`, `-playbook`, `-command`, and `-action`. If you edited one, it stays in `.agents/skills/`. To move the archived skills back, run `keystone migrate down`.

## Authoring

Let the commands write the sinks, and don't write them by hand. The exception is `wishlist.md`, which you keep yourself. There is no `keystone new` generator for a sink.
