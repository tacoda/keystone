---
kind: guide
id: process/release
description: 'The phase where the change lands and is announced.'
---
# Release

The phase where the change lands and is announced.

## Entry condition

The verification phase gate has passed. All sensors are green with fresh evidence. The diff is ready to commit.

## IRON LAWS

**NO CO-AUTHORS. NO AI ATTRIBUTION.**

Commit messages, PR descriptions, and tracker tickets do not mention Claude, an AI agent, or any tool used to author the change. The author is the human running the charter.

**NO COMMITS WITH FAILING PRE-COMMIT CHECKS.**

If the verification phase reported any sensor failure, the commit does not happen. CI is a backstop, not a substitute.

## Activities

### 1. Stage with intent

Stage specific files, not `git add .` or `git add -A`. Accidental inclusion of generated artifacts, secrets, or unrelated changes is what `.gitignore` is for, but staging discipline is the last defense.

### 2. Write the commit message

Conventional commits: `<type>(<scope>): <subject>`.

| Type | When |
|---|---|
| `feat` | New user-visible behavior |
| `fix` | Defect correction |
| `refactor` | Behavior-preserving structural change |
| `test` | Adding or improving tests |
| `docs` | Documentation only |
| `chore` | Tooling, dependencies, non-code changes |

GOLDEN RULE:

- **Aim for one logical change per commit.** A reviewer should read the diff in one sitting and explain it back.
- **Aim for structural changes separated from behavioral changes.** A refactor commit and a feature commit on the same diff hide each other.
- **Aim for commit messages that explain *why*, not what.** The diff already says what.

Title under 70 characters. Detail in the body.

### 3. Commit

Run `git commit`. The HEREDOC form preserves multi-paragraph bodies:

```
git commit -m "$(cat <<'EOF'
<type>(<scope>): <subject>

<body explaining why>
EOF
)"
```

Never `--no-verify`. Never `--no-gpg-sign`. If a hook fails, the commit did not happen — fix the cause and re-stage; do not amend.

### 4. Pull request (when working on a branch)

Title under 70 characters. Body links to the tracker ticket and explains the **why**. Include a test plan when the change touches user-facing behavior. No mentions of AI in the description.

### 5. Capture learnings

Invoke the **learn** action. If during the work the agent encountered a situation not covered by the corpus and made a judgment call, the reasoning gets captured to `.charter/learning/inbox/` for the next **synthesize** cycle.

This is the **Learning flywheel** kicking. Skipping it means the corpus never grows from this commit's lessons.

### 6. Release artifacts (when cutting a release)

For a release commit:

- Semver bump (`major.minor.patch`).
- Changelog entry under "Unreleased" promoted to a new version section with date.
- Tag the commit (`git tag v<version>`).
- Push tags (`git push --tags`).
- A single `chore(release): <version>` commit holds only the version-bump file changes.

### 7. CD pipeline integration

The corpus assumes a CI pipeline runs on every PR (sensors as a backstop) and, ideally, a CD pipeline triggers deploys on merge or tag.

- **CI exists, CD exists** → merge / tag is the release. The agent confirms the pipeline ran and reports its status in the PR.
- **CI exists, CD does not** → merge gates the change; deploy is a separate step the user owns. The agent should call out which step deploys.
- **Neither exists** → the charter still works for code quality, but release surfaces a warning: shipping without CI is a known gap. Recommend wiring at least lint + test into CI before relying on the **verify** action's evidence.

The pipeline configuration files (`.github/workflows/*`, `.gitlab-ci.yml`, `.circleci/config.yml`, etc.) are project state. Reference them from `state/CODEBASE_STATE.md` so the agent knows where deploy logic lives.

### 8. Postmortem (when incident-driven)

If the change is incident-recovery work, a blameless postmortem follows:

- Timeline (what happened, when, with logs and links).
- Impact (who or what, for how long).
- Root cause analysis (5-whys or causal chain — name the cause, not the symptom).
- What went well / what went poorly.
- Action items (owner, deadline, exit criteria).

Saved to `docs/postmortems/YYYY-MM-DD-<slug>.md`.

## Sensors

Release runs final-pass sensors:

- **Commit message validator** — conventional format, no AI attribution.
- **Lint, test** — final pass on the staged change, not the working tree.

## Gate condition

To exit release:

1. The commit landed cleanly (no `--no-verify`).
2. The PR (if any) was opened with body and test plan.
3. The **learn** action ran and any captures landed in `learning/inbox/`.
4. State files updated by the **verify** action are committed.

## Artifacts

| Kind | Location |
|---|---|
| Commit | git history |
| PR | tracker / forge |
| Tag (release) | git tag |
| Changelog entry (release) | `CHANGELOG.md` |
| Postmortem (incident) | `docs/postmortems/YYYY-MM-DD-<slug>.md` |
| Learning capture | `.charter/learning/inbox/<timestamp>-<slug>.md` |

## Anti-patterns

- AI attribution in commits, PRs, or tracker comments.
- `git add -A` followed by a large commit message — staging discipline matters.
- `--no-verify` to bypass a failing hook. Fix the hook or fix the cause; never bypass.
- Force-push to the main branch.
- "Release" commits that bundle a version bump with feature work.
- Postmortem that names a person as the root cause. Blameless or it is not a postmortem.
