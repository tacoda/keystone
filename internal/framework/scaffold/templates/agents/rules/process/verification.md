---
kind: rule
id: process/verification
description: 'The phase where claims are paired with evidence before commit.'
---
# Verification

The phase where claims are paired with evidence before commit.

## Entry condition

The implementation phase gate has passed. There is a working change in the working tree. The agent or user has signaled "ready to commit".

## IRON LAW

**NO COMPLETION CLAIMS WITHOUT FRESH VERIFICATION EVIDENCE.**

If you have not run the verification command in this turn, you cannot claim it passes. "Should pass" / "I think it works" / "the linter looked happy earlier" are not evidence. Skip any step → you are guessing, not verifying.

The verification phase exists to make this IRON LAW operable.

## Activities

### 1. Invoke the verify command

The **verify** command runs every verification hook (see Hooks). It also reads the diff and proposes updates to `state/` for any code regions the diff touched.

The agent invokes the command, then reads its full output, not the summary. Stale evidence is not evidence.

`AGENTS.md` says which hosts run hooks. On a host that does not, run the checks in `.agents/hooks/` yourself, or surface the commands to the human.

### 2. Confirm or fix

- **Hooks pass** → proceed to the gate.
- **Hooks fail** → return to the implementation phase. Do not "fix forward" inside verification.

Semantic review (spec adherence, review-agent findings, PR comments) moves to the **review phase** next. Verification is now mechanical only.

### 3. State updates

Where the **verify** command proposed state-layer diffs, the user reviews and accepts or edits per the scaffolding safety contract. State updates land in this phase, not after the commit.

## Hooks

Verification hooks gate the commit:

| Hook | What it proves |
|---|---|
| Lint | Surface-level style and pattern violations are clean. |
| Type-check | Signatures and contracts are consistent. |
| Test runner | All tests pass; 0 failures. |
| Build | The change compiles / packages. |
| Drift | No loaded rule is violated by the diff. |
| State proposal | Hooks that update state produced their diff for human review. |
| Commit-message | The proposed commit message is conventional and free of AI attribution. |

Full contracts (triggers, inputs, exit conditions, outputs, state writes) live in each hook file under `.agents/hooks/`.

A hook that *runs* and *exits successfully* is evidence. A hook that was run before any edit since is **stale evidence** and does not count.

## Claim → required evidence

| Claim | Required evidence |
|---|---|
| Tests pass | Test runner exits 0, 0 failures, fresh this turn |
| Linter clean | Lint exits 0, 0 errors, fresh this turn |
| Build succeeds | Build exits 0, fresh this turn |
| Bug fixed | The original-symptom test passes (and previously failed) |
| Regression test works | RED → GREEN cycle observed in this turn |
| Subagent completed | `git diff` shows the changes — not the agent's self-report |
| Requirements met | Line-by-line walk through the spec or plan |

## Gate condition

To exit verification:

1. Every applicable hook exited successfully **in this turn**.
2. State-layer diffs proposed by the **verify** command were accepted or edited.

Only then: proceed to the **review** command.

## Artifacts

| Kind | Location |
|---|---|
| State updates | `.agents/references/state/CODEBASE_STATE.md`, `.agents/references/state/migrations/active/...` |
| PR review responses | Comments on the PR |

## Anti-patterns

- **Stale evidence.** A test run from before the last edit does not count.
- **Partial evidence.** Linter clean does not imply tests pass.
- **Inferred evidence.** "The change looks small, so the build is fine" is a guess.
- **Self-reported subagent success.** Check the diff yourself.
- **Verifying with the same tool the agent wrote.** If the agent generated the test *and* claimed it passes without running it, both halves are unverified.
