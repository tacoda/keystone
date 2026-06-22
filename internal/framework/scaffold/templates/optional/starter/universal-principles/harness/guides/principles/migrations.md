# Migrations — rules

The non-negotiable and strongly-preferred rules from [`corpus/principles/migrations.md`](../../corpus/principles/migrations.md). Loaded ambient; enforced at planning and review.

## GOLDEN RULE

- **Aim to never ship a destructive schema change in the same release as code that depends on the old shape.** Expand first; contract only after every reader and writer is on the new shape.
- **Aim to never run a migration whose rollback path has not been written down.** "Drop the table" is not a rollback; restoring from backup is. If the rollback is restore-from-backup, that is the rollback — write it down and confirm the backup exists. See [[rollback]].
- **Aim for expand-contract.** Three phases: (1) **expand** — add the new shape, dual-write; (2) **migrate** — backfill, switch readers; (3) **contract** — remove the old shape. Each phase is its own PR.
- **Aim for online schema changes** when the table is large enough that a lock would matter. Tools: `pt-online-schema-change`, `gh-ost`, `pg_repack`, Vitess/PlanetScale online DDL. Bootstrap records which is wired.
- **Aim for backwards-compatible migrations end-to-end.** The new code reads the old shape; the old code tolerates the new shape. Roll forward and roll back must both work.
- **Aim to migrate in idempotent batches.** A backfill that fails halfway must be safe to re-run from the start.

## RULES

- **Migrations land in their own PR.** Schema and code mix only when the change is purely additive and the code does not yet read the new column.
- **No `DROP COLUMN`, `DROP TABLE`, `ALTER TYPE`** in the same PR as application code that no longer uses the column/table/type. Wait one release cycle minimum.
- **No data migrations inside a schema migration.** Schema migrations are DDL; data migrations are jobs. They have different failure modes and different rollback paths.
- **Long-running migrations run out-of-band.** Anything that holds a lock long enough to be noticeable by callers runs as a job, not as part of the deploy.
- **Every migration has a tracked end-state.** Active migrations live in `corpus/state/migrations/active/<NAME>.md` until contract is done; then move to `completed/`.

## Sensors

- The **drift sensor** flags migration files in a PR that also contains application code referencing the changed columns.
- **Bootstrap** detects the migration tool (Alembic, Knex, Rails, Goose, Diesel, Flyway, sqitch, atlas, dbmate, etc.) and writes its commands into `corpus/state/CODEBASE_STATE.md`.

---

Traces to: [`corpus/principles/migrations.md`](../../corpus/principles/migrations.md). See also: [[rollback]], [[scoping]].
