# Dependencies — rules

The non-negotiable and strongly-preferred rules from [`corpus/principles/dependencies.md`](../../corpus/principles/dependencies.md). Loaded ambient; enforced at planning, implementation, and review.

## GOLDEN RULES

- **Aim to never add a dependency the team has not approved.** Adding a package is a planning-phase decision — surface it in the plan, get explicit approval, then add. No package gets installed as a side effect of solving an unrelated problem.
- **Aim to never bypass the lockfile.** `package-lock.json`, `pnpm-lock.yaml`, `yarn.lock`, `poetry.lock`, `Gemfile.lock`, `Cargo.lock`, `go.sum` — these are committed, regenerated only by the package manager, and never hand-edited.
- **Aim to prefer the standard library** over a new dependency when the function fits in 50 lines and is well-understood.
- **Aim to justify every new direct dependency** in the commit or PR body: what it does, what alternatives exist, why this one.
- **Aim to upgrade in dedicated commits.** A dependency bump and a feature change in the same commit hide each other. See [[scoping]].
- **Aim to read the changelog on major-version bumps.** Major bumps are breaking by definition. Patch and minor bumps may proceed with the test suite as the gate.

## RULES

- **Lockfile changes are review-visible.** A diff that touches a lockfile but not its manifest is suspicious — investigate.
- **No transitive dependency added directly.** If the agent finds it needs a package only because another package needs it, pin the *direct* dependency, not the transitive.
- **No `--force`, `--legacy-peer-deps`, or version-resolution bypass flags** to make an install succeed. The error is the signal.
- **Removed code → removed dependency.** When the last call site disappears, the dependency goes too. The drift sensor watches for orphans.
- **Security advisories block the release phase.** If `npm audit`, `pip-audit`, `bundle audit`, `cargo audit`, or `govulncheck` reports a high-severity vulnerability on a touched dependency, fix it before merge.

## Sensors

- **Bootstrap** records the package manager and inserts the install / upgrade / audit commands into `corpus/state/CODEBASE_STATE.md`.
- The **drift sensor** flags new direct dependencies introduced by the diff so the review phase can confirm they were approved at planning.

---

Traces to: [`corpus/principles/dependencies.md`](../../corpus/principles/dependencies.md). See also: [[secrets-management]], [[scoping]].
