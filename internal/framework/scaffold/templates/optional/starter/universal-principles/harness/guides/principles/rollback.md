# Rollback — rules

The non-negotiable and strongly-preferred rules from [`corpus/principles/rollback.md`](../../corpus/principles/rollback.md). Loaded ambient; enforced at planning and release.

## GOLDEN PATH

- **Aim for every change to have a written rollback path before it merges.** "Revert the PR" is a valid rollback for pure-code changes — write it down anyway. For changes involving migrations, infrastructure, external services, or data, "revert the PR" is **not** sufficient. The rollback path must name the exact steps and the expected blast radius.
- **Aim to ship behind a flag** for risky changes. A change behind an off-by-default flag is a change that is not yet live. See [[modern-software-engineering]] on decoupling deployment from release.
- **Aim for the rollback path to be the same as the roll-forward path, but in reverse.** Symmetric procedures are practiced procedures.
- **Aim to test the rollback in staging.** A rollback path that has never been exercised is a hope.
- **Aim for changes small enough that revert is an option.** A 3000-line PR is one that the team cannot revert under pressure. See [[scoping]].

## RULES

- **Migrations follow the [[migrations]] expand-contract pattern.** The rollback for an expand step is `git revert` of the expand commit; for a contract step it is restore-from-backup. Plan accordingly.
- **Feature flags are documented in the spec.** Flag name, default state, who can toggle it, what it gates, when it will be removed.
- **Long-lived flags are tech debt.** A flag that has been on for 100% of traffic for 30 days gets removed in a dedicated PR. The harness's audit action may flag candidates.
- **Kill switches are not feature flags.** A kill switch turns *off* an existing capability for incident response; a feature flag *exposes* a new capability for rollout. The risk profile is different; document each clearly.

## PR body must include (when applicable)

A `## Rollback` section naming:

1. The condition that would trigger a rollback (what does "this went wrong" look like).
2. The exact rollback command(s) — `git revert <sha>`, `kubectl rollout undo`, `flipper disable <flag>`, etc. Bootstrap records the project's tooling.
3. The expected user impact during rollback.

## Sensors

- The **commit-message sensor** does not enforce a rollback section directly. The review phase confirms it for changes that touch listed risky paths.
- **Bootstrap** detects the feature-flag system (LaunchDarkly, GrowthBook, Unleash, Flipper, Statsig, custom config) and records the toggle command in `corpus/state/CODEBASE_STATE.md`. If none is wired, the rollback rule notes "no flag system available — risky changes ship in their own deploy with a documented revert."

---

Traces to: [`corpus/principles/rollback.md`](../../corpus/principles/rollback.md). See also: [[migrations]], [[modern-software-engineering]], [[scoping]], [[ci-failure]].
