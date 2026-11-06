# Optional content

Files that `keystone init` installs **only when the user opts in** to a given category/label combination.

## Layout

```
optional/
  <category-id>/
    <label-id>/
      agents/   <files that mirror the charter root>
```

`<category-id>` is one of the categories defined in `options.go` (e.g. `architecture`, `language`, `testing`). `<label-id>` is one of that category's allowed values (e.g. `hexagonal`, `go`, `tdd`).

Files under `optional/<cat>/<label>/agents/` are copied **as if rooted at the charter root** (`.agents/` by default). For architecture and compliance overlays, the content lands inside a category-named subdirectory: `optional/architecture/hexagonal/agents/references/architecture/hexagonal.md` lands at `.agents/references/architecture/hexagonal.md`, with its paired rule file at `.agents/rules/architecture/hexagonal.md`.

Each opt-in label typically ships two files for every concept: the informational reference file and the paired rule file. For labels with stack-shaped content (e.g., `architecture/mvc`), the bundle may also seed concern-specific files under `references/idioms/<label>/` and `rules/idioms/<label>/` (`models.md`, `controllers.md`, `views.md`), since those are project-stack idioms.

The `agent` category is **excluded**: the chosen agents become `adapters` in `keystone.json`, and `keystone project` writes their host files.

## Collision policy

Optional content copies with `skipIfExists`: it never clobbers a file that already exists at the destination. A warning is printed when a file is skipped. Two labels with overlapping content will install the first and skip the second.

## When to add files here

A file belongs in `optional/` rather than `agents/` when it is:

- **Too opinionated** for the always-on universal content (e.g. a particular architecture style).
- **Stack-specific** in a way that ships ready-to-go starter content (e.g. a Go test idiom file).
- **Compliance-scoped** content that only matters under a particular regulatory regime.

If it belongs to *every* project, it goes in `agents/` directly.
