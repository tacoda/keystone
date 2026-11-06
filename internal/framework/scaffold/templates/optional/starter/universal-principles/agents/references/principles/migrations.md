# Migrations

A schema migration changes the *shape* of state that survives the deploy. The deploy itself is finite; the state is permanent. Migrations that treat schema as code — atomically, with rollback by `git revert` — are migrations that lose data. The discipline of migrations is the discipline of separating the *change in the schema* from the *change in the code that uses it*.

> **Rules extracted:** [`guides/principles/migrations.md`](../../guides/principles/migrations.md).

## What it asks of you

- Separate schema change from behavior change. They have different rollback semantics, different blast radii, and different time-to-recover. Mixing them couples them.
- Make every step independently revertable. Each migration commit is also a migration-revert commit; if the revert is "restore from backup," the migration is not designed.
- Plan for the rolling deploy. There is no instant when "the deploy" happens — there is a window where old and new code run together. The schema must serve both.
- Treat large tables as production. A `DROP COLUMN` on 100 rows is free; on a hundred million rows it is an incident.

## Why it holds

Expand-contract (Sadalage & Fowler, *Refactoring Databases*, 2006) is the canonical pattern. Add the new column without removing the old; backfill; switch readers; switch writers; only then remove the old column. Each step is a small, reviewable, individually-revertable change. The pattern predates the cloud — it works because it never assumes the deploy is instantaneous.

PlanetScale, GitHub (gh-ost), Percona (pt-online-schema-change), and Vitess all exist to extend the pattern to tables where the *change itself* cannot be atomic at the database level. Online schema changes use shadow tables, change-data-capture, and atomic swaps to avoid the long-lock failure mode.

The "blue-green deploy + migration" pattern fails when a migration is not backwards-compatible: the green pool starts up, runs the migration, and the blue pool — still serving traffic — sees a schema it does not understand. Backwards-compatible migrations make blue-green safe.

## Anti-patterns

- A single PR that adds a column, populates it, switches all callers, and drops the old column.
- A `DROP TABLE` rollback that assumes the data is recoverable. Once dropped, it is gone.
- A backfill in a migration file. Backfills time out, retry awkwardly, and run inside the deploy's transaction boundary. Make them jobs.
- "We have a maintenance window" as the safety net. The window is not the safety; backwards compatibility is.
- Forgetting to test the *intermediate* state — old code, new schema. That state runs in production for the duration of the rolling deploy.

## References

- Sadalage, P. & Fowler, M. *Refactoring Databases* (2006).
- Humble, J. & Farley, D. *Continuous Delivery* (2010) — zero-downtime database changes.
- GitHub Engineering. [*gh-ost: Online Schema Migration for MySQL*](https://github.com/github/gh-ost).
- PlanetScale docs. [*Branching and deploy requests*](https://planetscale.com/docs/concepts/deploy-requests).
- Fowler, M. [*ParallelChange*](https://martinfowler.com/bliki/ParallelChange.html) — the expand-contract pattern at the API level.

---

Forward link: [`guides/principles/migrations.md`](../../guides/principles/migrations.md).
