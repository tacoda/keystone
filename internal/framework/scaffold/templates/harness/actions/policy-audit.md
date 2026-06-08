# policy-audit

**Audit the codebase against every installed policy.** Read [`harness/policies/README.md`](policies/README.md) and walk the installed policies recorded in `harness/.keystone.lock`.

This is the **compliance** half of policy auditing — does the code actually do what the policies require, and is the strict cascade intact? Lockfile integrity (hashes, local edits, ref freshness) is out of scope here.

## Activities

1. **Run `keystone policy verify`.** Reports two categories:
   - **Strict violations** — items a policy declared `strict` that are overridden by a lower tier (team policy or project file). Hard findings; the strict contract has been broken. The user must remove the shadowing file or raise the change with the policy author before the rest of this action's findings can be trusted.
   - **Required gaps** — items a policy declared `required` that no tier has defined. Advisory findings; the project should define them at `harness/<kind>/<name>.md`.
2. **List installed policies.** Read `harness/.keystone.lock` and enumerate every entry under `policies:`. For each policy:
   - Name, tier (`org` or `team`), resolved ref.
   - Path: `harness/policies/<name>/guides/` (rules to enforce), `harness/policies/<name>/playbooks/` and `harness/policies/<name>/actions/` (workflows and units of work), and `harness/policies/<name>/corpus/` (reasoning context, on-demand).
3. **Load each policy's guides, playbooks, and actions.** For each `harness/policies/<name>/{guides,playbooks,actions}/**/*.md`, read the file. Treat every rule as a claim the codebase should be checkable against.
4. **Check the codebase against each rule.** For each rule, decide one of:
   - **compliant** — the codebase satisfies the rule. Cite the evidence (file:line or pattern).
   - **violation** — the codebase contradicts the rule. Cite the offending file:line.
   - **inapplicable** — the rule's preconditions don't apply to this project (e.g., a Python rule in a Go-only repo). Say why.
   - **uncheckable** — the rule is prose without a verifiable claim. Say what would be needed to make it checkable.
5. **Report findings** grouped by policy, then by rule. One bullet per finding with the cited evidence. Strict-cascade violations are reported first.

## Output

One report:

```
Strict cascade
  <empty if clean, else one block per violation from `keystone policy verify`>

Policy: <name> (tier: <org|team>) @ <ref>
  ✓ <rule-path>: compliant — <evidence>
  ✗ <rule-path>: violation — <file:line>
  · <rule-path>: inapplicable — <reason>
  ? <rule-path>: uncheckable — <reason>

Policy: <name> (tier: <org|team>) @ <ref>
  ...
```

## When to invoke

- Before any release that touches a policy-covered area.
- Periodically (suggested: monthly or per audit cycle).
- After `keystone policy add` or `keystone policy update` — the policy changed; the codebase may now violate rules that weren't there before.

## What this action does not do

- Does not check lockfile integrity (file hashes, local edits, upstream ref freshness). Those checks are mechanical and belong in the binary if at all; today they live in [`audit.md`](audit.md) Pruning if anywhere.
- Does not modify code to fix violations. Findings are inputs to **spec** for the next round of work.
- Does not promote policy rules into project guides. If a policy rule is so consistently violated that you decide to weaken or remove it, that's a conversation with the policy owner — not a local edit.
