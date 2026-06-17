---
kind: guide
id: process/ci-failure
description: 'What to do when CI fails — on a PR, after a merge, or post-deploy.'
---
# CI Failure Remediation

What to do when CI fails — on a PR, after a merge, or post-deploy. The release phase covers the happy path; this covers everything else.

## Entry conditions

Any of:

- CI reports a red check on an open PR.
- CI fails on `main` after a merge.
- A deploy fails or rolls back.
- A scheduled job (cron, nightly) starts failing.

## Activities

### 1. Fetch the failure

Pull the failing log from CI directly — never guess from the job name. Bootstrap records the command in `corpus/state/CODEBASE_STATE.md`:

- GitHub Actions: `gh run view <id> --log-failed`
- GitLab: `glab ci view <id>`
- CircleCI: `circleci runs logs <id>`
- Buildkite: `bk job logs <id>`
- Jenkins: project-specific (recorded by bootstrap)

### 2. Reproduce locally

Run the same sensor command locally that CI ran. If the failure does not reproduce:

- Compare environment (Node/Python/Go version, OS, env vars).
- Check for non-determinism — see [[determinism]]. A test that passes locally and fails in CI is usually non-deterministic, not "CI is broken."
- Check for ordering — does CI run tests in parallel? In a different order?

A failure that only reproduces in CI is still a failure. Do not merge around it.

### 3. Fix at the root

Diagnose the *root cause*, not the surface. A test that passes after a retry has not been fixed; it has been ignored.

- **Real failure** → fix the code or the test. Re-invoke the **verify** action.
- **Flaky test** → quarantine with an explicit tracker card for the fix; do not silently `pytest --rerun`. See [[determinism]].
- **Infra/runner issue** → file a tracker card; do not silently rerun until green.

### 4. Post-merge failure on `main`

`main` going red is incident-shaped. Default response:

- **Revert first, debug second** if a fix is not obvious in <10 minutes. A reverted main is a green main. See [[rollback]].
- The revert PR carries a short body: what failed, why reverting, link to a follow-up card.

### 5. Deploy failure

If CD ran and a deploy failed:

- Check the deploy log via the recorded command.
- Roll back via the recorded rollback procedure. See [[rollback]].
- Open a postmortem-tracking artifact only if user impact occurred.

## GOLDEN PATH

- **Aim never to merge around a failing check.** Disabling a check, marking it non-required, or using admin-merge to bypass CI is a [[dangerous-actions|dangerous action]] requiring explicit confirmation. Builds on the existing IRON LAW "no commits with failing sensors" — CI is the same sensors at the pipeline level.
- **Aim to revert before debugging on `main`.** A reverted main lets the team keep working while the cause is investigated.

## RULES

- **Do not retry until green.** Each retry without a diagnosis is evidence dressed as confirmation. If a job needs to retry, the *underlying flake* is the bug.
- **Stale log = stale evidence.** A log from before the latest push does not count. Fetch fresh.
- **Carry the tracker card forward.** The remediation links to the original spec's tracker card *and* its own follow-up card if the fix introduces tech debt.

## Anti-patterns

- "Rerunning until it passes."
- "CI was flaky" without an open tracker card for the flake.
- Reverting silently — a revert is a release event; communicate it.
- Touching `main` directly to fix a `main` failure. Revert via PR.

---

See also: [[determinism]], [[rollback]], [[dangerous-actions]], `release.md`.
