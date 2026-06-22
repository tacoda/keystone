---
kind: sensor
id: secret-scan
description: 'Scans the diff (or repo, on audit) for committed secrets — API keys, tokens, private keys, credentials.'
host_triggers:
  - phase: PreToolUse
    matcher: "Edit|Write|MultiEdit"
    command: keystone verify --sensor secret-scan
    timeout: 5
---
# Sensor: secret-scan

Scans the diff (or repo, on **audit**) for committed secrets — API keys, tokens, private keys, credentials.

- **Trigger** — **verify** (diff-scoped, gate) and **audit** (history-scoped, advisory).
- **Inputs** — the secret-scan command from `corpus/state/CODEBASE_STATE.md` (e.g., `gitleaks detect`, `trufflehog filesystem`, `detect-secrets scan`). Tool choice is the project's; the sensor describes the contract.
- **Exit condition** — exit code 0 and no findings.
- **Output** — pass/fail. On fail: file, line, and the secret type (the tool's classification, not the secret itself).
- **State writes** — none. Findings are not recorded in state — that would itself leak.

## Configuration

Tool selection lives in `CODEBASE_STATE.md`. The **bootstrap** action proposes a default based on detected stack (`gitleaks` if no other tool is configured). Allowlists for known-safe patterns (test fixtures, public keys) live wherever the chosen tool reads them (`.gitleaks.toml`, `.secrets.baseline`, etc.) — not in the harness.

## On finding a secret

Failing this sensor is not the end of the work. The committer still needs to:

1. Rotate the secret at its origin (the secret is now in git history; assume it's compromised).
2. Remove the secret from working tree.
3. Re-commit.
4. Decide whether history rewrite is warranted (usually yes for production secrets, usually no for low-value test fixtures).

The sensor does not do any of this. It just blocks the commit.
