# Continuous Delivery

The software is **always in a releasable state.** Every commit on the main branch passes through an automated pipeline that proves it can ship to production; the decision to deploy is a business decision, not an engineering one. Articulated by Jez Humble and David Farley in *Continuous Delivery* (2010) and reinforced by the empirical findings of Forsgren, Humble & Kim in *Accelerate* (2018), which identified Continuous Delivery practices as the strongest predictors of high-performing engineering organizations.

Continuous Delivery is a *practice* — but it is also an **architectural** choice in the sense that it constrains every other architectural choice. A codebase architected without CD in mind cannot adopt CD without significant restructuring. It belongs in the architecture catalog because choosing it shapes every layer.

> **Rules extracted:** [`guides/continuous-delivery.md`](../guides/continuous-delivery.md). This file holds the full reasoning, anti-patterns, and references.

## The core claim

> The cheapest way to deliver software, end to end, is to deliver it continuously. — Humble & Farley

The bet: integration and release pain compound. Deferring them produces large, risky batches. Removing the deferral — by integrating and proving release-readiness on every commit — produces small, safe changes. The empirical evidence (lead time, deploy frequency, change failure rate, mean time to recover) is in *Accelerate*; CD organizations consistently outperform on all four.

Distinct from related terms:

- **Continuous Integration** — every commit merges to the trunk and runs the full build + test suite. Necessary but not sufficient.
- **Continuous Delivery** — every commit produces a release artifact that **could** be deployed to production. The deploy itself remains a manual click.
- **Continuous Deployment** — every passing commit **is** deployed to production automatically. Continuous Delivery is the precondition; Continuous Deployment is the extension.

CD is the bar; whether the final deploy step is automatic or human-gated is a product/risk decision.

## What it asks of the architecture

- **Trunk-based development.** Long-lived branches are CD's enemy. Branches integrate divergent code; CD requires continuous integration. Feature toggles, not feature branches, control visibility. See [[refactoring]] (the two-hat rule applies here too: small commits, decoupled from release).
- **Decouple deploy from release.** Deploying code to production and exposing the feature to users are two different actions. Feature flags, dark launches, canary releases — the deploy happens often; the release happens when business says so.
- **Every commit triggers the pipeline.** Build, lint, type-check, unit tests, integration tests, security scans, artifact build, artifact signing. The pipeline is the gate; nothing reaches production without passing it.
- **Database migrations are part of the deploy.** Forward-compatible schema changes; expand-then-contract migration patterns. Rollback rarely means "undo the migration"; it means "deploy the previous code, which still works against the new schema."
- **Observability before automation.** Continuous Deployment without observability is unsupervised production change. See [[observability]] — the IRON LAW (if you cannot debug it from the data, you cannot debug it) is the precondition for automating deployment.
- **Tests fast enough to run on every commit.** A 90-minute test suite stops being run; CD disintegrates. See [[testing-patterns]] on the pyramid.
- **Production-like environments early in the pipeline.** "It worked on staging" requires staging to actually resemble production. Containers, infrastructure-as-code, and configuration parity reduce the gap.

These are not optional; they are what makes CD possible. A team that adopts CD without them runs the pipeline as theater while shipping the same way as before.

## What it asks of you

- When you commit, ask: *can this be deployed right now?* If not, why not — and is the answer something to fix in the commit, or in the pipeline? See [[fail-fast]].
- When you introduce a long-lived branch, push back. Most "I'll merge it when it's ready" branches accumulate cost faster than they create value; feature flags solve the same visibility problem without the merge debt.
- When you deploy a schema change, design it to be **forward-compatible** with the running code, *then* deploy the code that uses it. Two phases — never one big-bang. See [[idempotency]] (rollouts must be safely repeatable).
- When the pipeline is slow, treat it as a defect, not a fact of life. Test parallelization, build caching, dependency-graph-aware execution. The pipeline's runtime is the lower bound on the team's feedback loop. See [[modern-software-engineering]] (short feedback loops).
- When the pipeline is flaky, find the source and fix it. A flaky pipeline gets routed around (retries, skips, force-merges), which is how CD silently dies. See [[fail-fast]] anti-patterns on `@Retry(3)`.
- When you write a feature behind a flag, write the **flag retirement** ticket at the same time. Flags that outlive their feature are technical debt that compounds.

## Anti-patterns

- A "release branch" that diverges from main for two weeks before deployment. Integration cost compounds; the deploy is now a high-risk merge.
- Manual QA as a deployment gate, performed by a separate team after dev "is done." The handoff is the failure mode; CD demands quality is built in, not inspected in.
- Feature branches that live for weeks, integrated only when "complete." See [[refactoring]] (two-hat rule); see Conway's law (the team has split into "in-progress" and "merged" subteams).
- A deploy that requires a runbook with 23 steps, two pages of warnings, and an incantation about cache invalidation. Each manual step is a place to make a mistake under pressure.
- A pipeline whose tests fail intermittently and are routinely re-run until green. The pipeline is no longer a gate; it is a coin flip with the team's trust as the stake.
- A "feature flag" that has been on for nine months but cannot be removed because no one remembers what the off-path does. The flag has metastasized into permanent complexity. See [[simplicity]].
- A team that has adopted CD tooling without trunk-based development. The dashboard says "Continuous Delivery"; the workflow is still gated, batched, and slow.

## References

- Humble, J., & Farley, D. (2010). *Continuous Delivery: Reliable Software Releases through Build, Test, and Deployment Automation*. Addison-Wesley. (The canonical text.)
- Farley, D. (2021). *Modern Software Engineering: Doing What Works to Build Better Software Faster*. Addison-Wesley. (CD recast as the foundation of engineering-as-discipline; see [[modern-software-engineering]].)
- Forsgren, N., Humble, J., & Kim, G. (2018). *Accelerate: The Science of Lean Software and DevOps*. IT Revolution Press. (The empirical evidence — the four DORA metrics, the practices that correlate with high performance.)
- Fowler, M. (2006). "Continuous Integration." martinfowler.com/articles/continuousIntegration.html.
- Hodgson, P. (2017). "Feature Toggles (aka Feature Flags)." martinfowler.com/articles/feature-toggles.html.
- Newman, S. (2021). *Building Microservices*, 2nd ed. O'Reilly. (Operational practices that make CD sustainable across many services.)
- Allspaw, J., & Robbins, J. (2010). *Web Operations: Keeping the Data on Time*. O'Reilly. (The operational culture that makes "deploy on Friday" not insane.)
