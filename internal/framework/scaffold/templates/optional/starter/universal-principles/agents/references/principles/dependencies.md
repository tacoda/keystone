# Dependencies

Every dependency added to a project is a permanent decision with permanent cost. Code you do not own runs in your build, in your CI, in your binary, and — at runtime — in your users' processes. Treating dependency selection as the same kind of decision as writing a function, rather than as the kind of decision as picking an API to expose, is the root of supply-chain risk.

> **Rules extracted:** [`guides/principles/dependencies.md`](../../guides/principles/dependencies.md).

## What it asks of you

- Treat every direct dependency as a stable, named contract with the rest of the codebase. Adding one is API design.
- Treat the lockfile as the *actual* dependency declaration. The manifest is a *constraint*; the lockfile is the *resolution*. Reviews read the lockfile.
- Treat upgrades as their own work, not as cleanup. Upgrades are behavior changes — the wire is the same, the runtime is not.
- Treat transitive dependencies as a *surface area* metric, not a *correctness* metric. They run too.

## Why it holds

The npm `left-pad` incident (2016) made the cost of unmanaged dependencies legible: an 11-line module's removal broke a meaningful slice of the JavaScript ecosystem in real time. The `event-stream` (2018) and `ua-parser-js` (2021) compromises showed that transitive dependencies are an attack surface even when the direct list looks clean. The xz backdoor (2024) extended this to system-level supply chains.

The economic argument is older. Adding a dependency moves work *out* of the project's control. Some moves are correct (you should not rewrite OpenSSL); many are not (you do not need 14 packages to format a date). Hunt & Thomas's *ETC — Easier To Change* — generally favors fewer, deeper dependencies over many shallow ones.

Lockfiles exist because semver is a *promise*, and promises are *broken*. A package manager resolving `^1.2.3` today may give you `1.2.5`; tomorrow `1.2.6`; the day after a patched-but-incompatible `1.2.7`. The lockfile pins the resolution so the build is reproducible across machines, CI runs, and weeks.

## Anti-patterns

- A dependency added to solve a one-line problem ("I need to debounce — let me install `lodash.debounce`").
- Mass-installing a meta-package to get one function ("`npm install lodash` for `lodash.get`").
- Hand-editing the lockfile to "fix" a version conflict.
- Running `--force` / `--legacy-peer-deps` to make an install pass without understanding what the resolver was protecting you from.
- Pinning all transitives ("overriding" the resolver). Sometimes correct; often a sign of misunderstanding semver ranges.
- Auto-merging dependabot PRs without a test suite that exercises the relevant code paths.
- Treating dev dependencies as harmless. A compromised dev dep ships malicious code in your CI.

## References

- Hunt, A. & Thomas, D. *The Pragmatic Programmer* — ETC, orthogonality.
- Williams, C. (2016). [left-pad incident retrospective](https://www.theregister.com/2016/03/23/npm_left_pad_chaos/).
- Cox, R. (2019). [*Surviving Software Dependencies*](https://research.swtch.com/deps).
- Ohm, M. et al. (2020). *Backstabber's Knife Collection: A Review of Open Source Software Supply Chain Attacks*.
- npm, GitHub, OSV.dev advisories — the lived corpus.

---

Forward link: [`guides/principles/dependencies.md`](../../guides/principles/dependencies.md).
