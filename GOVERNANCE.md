# Governance

How the Keystone project is run, and how a **charter** is amended. Keystone is
the agent charter framework; this document governs both the software and
the charter model it maintains. It complements [`CONTRIBUTING.md`](CONTRIBUTING.md)
(how to contribute) and [`GLOSSARY.md`](GLOSSARY.md) (the charter/harness lexicon).

## Roles

- **Maintainers** — merge changes, cut releases, and ratify charter amendments.
  Maintainers are listed in `OWNERS.md` (or the repo's maintainer list). Today
  the project is small; the growth path below describes how maintainership
  expands.
- **Contributors** — anyone who opens an issue or PR. Contributions are welcome
  under the terms in `CONTRIBUTING.md`. No prior standing is required.
- **Users** — anyone running Keystone or operating a charter. User feedback
  (issues, adoption reports) is a first-class input to charter amendments.

### Becoming a maintainer

Sustained, high-quality contribution (code, review, docs, or charter/policy
authorship) earns a maintainer invitation by lazy consensus of the existing
maintainers. The bar is judgment and follow-through, not commit volume.

## Decision-making

- **Lazy consensus.** Most changes proceed once proposed and un-objected-to
  within a reasonable window. Silence is assent.
- **Maintainer approval.** Code changes merge on at least one maintainer's
  approval and a green Code Health gate (see `CONTRIBUTING.md`).
- **Escalation.** Unresolved disagreement is decided by the maintainers; while
  the project is solo-maintained, the maintainer decides and records the
  rationale (an ADR under `docs/adr/`).

Breaking changes to the **charter contract** (the on-disk layout or the
frontmatter schema) ship in a major release with a migration and an ADR.

## Charter amendments

A charter is authored standards; changing it is an *amendment*, and amendments
follow a defined pipeline rather than ad-hoc edits:

1. **Propose.** A surprise, incident, or review finding is captured as a
   learning candidate under `.charter/learning/inbox/` (`keystone learn`).
2. **Synthesize.** Accepted candidates are promoted into the right layer —
   a guide, corpus entry, sensor, or tool (`keystone synthesize`).
3. **Ratify.** A maintainer accepts the promotion in review. The change lands
   in `.charter/` and is indexed.
4. **Record.** Material amendments are recorded as a governed document with
   `type: amendment` (or an ADR for architectural ones), so the charter's
   evolution is human-readable, not just reconstructable from git.

This is the same flywheel the tool ships (`learn → synthesize`); governance
names it as the charter's amendment process.

## Provenance & ratification

Every primitive carries a **provenance** the walker derives from its path:
`project` for the repo's own charter, `policy/<name>` for a vendored policy
layer. `keystone charter show --effective` renders the resolved cascade and
annotates what each winner overrides.

- **Project layer** wins by default — the repo's authors have final say over
  their own charter.
- **Policies** are vendored, **pinned by version and content hash**, and
  **drift-reset by `keystone verify`**. A policy's `strict:` declarations lock
  items absolutely; nothing else may override them. This is the ratification
  boundary between an org/team policy and a project.
- **Amendment records** (`type: amendment` documents / ADRs) name who proposed
  and ratified a change and why — provenance for humans, alongside the
  path-derived provenance for tools.

## Changes to this document

Governance changes are themselves amendments: propose via PR, ratified by
maintainer lazy consensus. Substantive changes are recorded as an ADR.

## Code of conduct

Be respectful and assume good faith. Harassment, discrimination, and bad-faith
conduct are not tolerated. Report concerns to the maintainers privately. A
formal `CODE_OF_CONDUCT.md` may be adopted as the contributor base grows.

## Growth path

The project is honest about its stage: it is small and currently
solo-to-lightly maintained. The intended trajectory:

1. Production use in more than one organization, documented as references.
2. Maintainers from more than one organization.
3. A published roadmap and a public issue tracker (in place).

These are the standard markers of a healthy, multi-stakeholder open-source
project, and the criteria this governance is designed to grow into.
