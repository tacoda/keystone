---
kind: guide
id: process/sensitive-files
description: 'Files the agent must never read into context, write to, or commit.'
---
# Sensitive Files

Files the agent must never read into context, write to, or commit. Enforced at every phase — sensitive-file access is never a side effect of the task.

## IRON LAW

**NEVER READ OR WRITE A FILE ON THE SENSITIVE LIST.**

If a file matches a pattern below, the agent does not open it, paste it, or pass it as an argument to any tool. Asking the user to provide the value out-of-band is the right escalation. Logs that may contain request bodies, headers, or query strings count as sensitive too.

## The list

Default patterns the harness treats as sensitive. The **bootstrap** action augments these by scanning `.gitignore` and `.git/info/exclude` for further candidates, and records the final list in `corpus/state/CODEBASE_STATE.md`.

- `.env`, `.env.*` — environment files
- `*.pem`, `*.key`, `*.p12`, `*.pfx` — private keys and certificates
- `id_rsa`, `id_ed25519`, `*.ppk` — SSH keys
- `*.gpg`, `*.asc` — encrypted artifacts
- `credentials.json`, `service-account*.json` — cloud credentials
- `*.kdbx` — password databases
- `.netrc`, `.npmrc`, `.pypirc` — when they contain tokens
- `secrets/`, `vault/` directories
- Any path matching `*secret*`, `*credential*`, `*password*` outside test fixtures

## RULES

- **No `cat`, `head`, `tail` on sensitive files.** No `grep` either — even searching leaks values into context.
- **Test fixtures that look like credentials must be obviously fake** (`fake-key-do-not-use`, `xxx`, `00000000`). Anything that *looks* real is treated as real.
- **If the agent needs a value from a sensitive file** (a hostname, a username, but not the secret), ask the user.
- **A `.env.example` ships with placeholders, never real values.** Bootstrap diffs `.env.example` against `.env` to confirm.

## GOLDEN PATH

- **Aim to keep secrets out of the repo entirely.** Workload identity (IAM roles, OIDC) beats long-lived keys. See [[secrets-management]].
- **Aim to fail loud on accidental sensitive reads.** If an agent action somehow pulls one in, the next sensor pass should flag it — drift sensor and commit-message sensor both consult this list.

## Sensors

- The [drift sensor](sensors/drift.md) reads this list and flags any diff that adds a path matching a sensitive pattern.
- The [commit-message sensor](sensors/commit-message.md) blocks commits that touch matching paths.

## Anti-patterns

- "Just this once" reading `.env` to confirm a variable exists. Ask the user.
- Copying a credential into a comment "for reference."
- Treating staging or test credentials as harmless. They aren't.
- Committing a `.env.example` with values that look real instead of placeholders.

---

See also: [[secrets-management]], [[dangerous-actions]].
