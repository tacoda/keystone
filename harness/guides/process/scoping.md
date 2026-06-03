# Scoping

How big any single commit, PR, or implementation pass is allowed to get before it must split.

## RULES

- **One concern per commit.** A commit fixes one bug, adds one feature, or refactors one shape — not two.
- **Target a diff a reviewer can read in one sitting.** Default ceiling: **≤500 changed lines (excluding generated files and lockfiles), ≤10 source files**. Bootstrap may tighten these from project history.
- **Split when the ceiling is reached.** If the change needs more, ship it as a sequence of dependent commits or PRs, each independently reviewable.
- **Generated artifacts (lockfiles, snapshots, build output) sit in their own commit** when they cross 50 lines. Mixing them with logic hides the logic.
- **Refactor and behavior change never share a commit.** See [[refactoring]] and `release.md` GOLDEN RULES.

## GOLDEN RULES

- **Aim to keep the diff readable in one screen of `git diff --stat`.** If the stats themselves do not fit, the PR is too big.
- **Aim to make every commit independently revertable.** Two intertwined changes in one commit make the revert worse than the original problem. See [[rollback]].
- **Aim to push pre-work up to its own PR.** Renames, type cleanups, dependency bumps that *enable* the real change should land first, alone, and green.

## Sensors

- The **commit-message sensor** does not enforce scope (it cannot tell intent from size). Scope is enforced at the review phase by reviewers reading the diff, and by the agent self-checking before staging.

## Pacing modes

- **paired** — agent asks before committing anything that exceeds the ceiling.
- **solo** — agent splits autonomously when the ceiling is exceeded; reports the split.
- **autopilot** — same as solo. The split itself is documented in the assumption log.

## Anti-patterns

- A 2000-line "fix typo + reorganize + add feature" megaPR.
- Hiding a behavior change inside a refactor commit so the diff looks small per-file.
- "I'll squash it before merge." Squashing erases the audit trail; ship atomic from the start.
- Renaming half the call sites of a function and the implementation in the same commit. Rename, then change.

---

See also: [[refactoring]], `release.md`, [[rollback]].
