# Optional content

Files that `keystone init` installs **only when the user opts in** to a given category/label combination.

## Layout

```
optional/
  <category-id>/
    <label-id>/
      <files that mirror the install root>
```

`<category-id>` is one of the categories defined in `options.go` (e.g. `architecture`, `language`, `testing`). `<label-id>` is one of that category's allowed values (e.g. `hexagonal`, `go`, `tdd`).

Files under `optional/<cat>/<label>/` are copied **as if rooted at the install destination**, so `optional/architecture/hexagonal/harness/corpus/principles/hexagonal-architecture.md` lands at `<destDir>/harness/corpus/principles/hexagonal-architecture.md`. A paired guide at `optional/architecture/hexagonal/harness/guides/principles/hexagonal-architecture.md` lands at `<destDir>/harness/guides/principles/hexagonal-architecture.md`.

Each opt-in label typically ships two files for every concept — the informational corpus file and the rule-bearing guide. For some labels (e.g., `architecture/mvc`), the bundle also seeds concern-specific files under `corpus/idioms/<label>/` and `guides/idioms/<label>/` (e.g., `models.md`, `controllers.md`, `views.md`).

The `agent` category is **excluded** — agent bundles already live in `targets/<agent>/`.

## Collision policy

Optional content copies with `skipIfExists`: it never clobbers a file that already exists at the destination. A warning is printed when a file is skipped. Two labels with overlapping content will install the first and skip the second.

## When to add files here

A file belongs in `optional/` rather than `harness/` when it is:

- **Too opinionated** for the always-on principles (e.g. a particular architecture style).
- **Stack-specific** in a way that ships ready-to-go starter content (e.g. a Go test idiom file).
- **Compliance-scoped** content that only matters under a particular regulatory regime.

If it belongs to *every* project, it goes in `harness/` directly.
