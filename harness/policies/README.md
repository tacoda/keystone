# Policies

**Level 3 extensions.** Policies are distributable bundles of governance content that ship alongside the project harness. Each policy lives in its own namespace under `policies/` and is treated as a unit — installed, updated, and pinned together.

Policies sit at one of two tiers — declared in the policy manifest's `tier:` field:

- **`org`** — distributed by the whole organization. Authoritative for cross-team standards (vendor lists, license rules, release gates, compliance regimes).
- **`team`** — distributed by a specific team within the org. Layered on top of org policies, can refine or extend, can mark its own items strict (subject to any org strict above it).

The **project** is the third tier: the harness root itself (`harness/{guides,playbooks,actions}/`). The project is always the leaf — items here are not distributed as policies.

There are two kinds of installed policies on disk:

- **`universal/`** — the default policy. Ships embedded in the keystone binary at tier `org`. Contains the universal engineering principles (reasoning + rule extracts) that apply to every Keystone install regardless of stack, domain, or org.
- **`<name>/`** — installed via `keystone init --policy <ref>`, `keystone policy add <ref>`, or `keystone policy update`. Tier is declared by the policy author in the manifest.

## Layout inside a policy

Each policy mirrors the project layout — `corpus/` for reasoning, `guides/` for rules, plus optional `playbooks/` and `actions/` — scoped to that policy's namespace:

```
harness/policies/<name>/
├── corpus/              # on-demand reasoning (reached via guide forward-link)
│   └── <topic>/<file>.md
├── guides/              # ambient rules — always loaded
│   └── <topic>/<file>.md
├── playbooks/           # optional — ordered sets of actions an org distributes
│   └── <name>.md
└── actions/             # optional — shared actions (e.g., rubocop_for_ruby)
    └── <name>.md
```

A policy playbook references actions by name; the action itself can live in the same policy, in another policy, or in the project tree. Policies do not embed action definitions inside their playbooks.

Sensors are **not** part of a policy. Sensors describe project tooling (lint, type-check, test commands) — they're an integration concern that the project owns. A policy may declare *what* must be checked; the project decides *how*.

## Activation

Sub-paths inside a policy determine activation:

| Path | Activation |
|---|---|
| `<name>/guides/...` | **Ambient** — always loaded, same as project guides |
| `<name>/playbooks/...` | **On invocation** — named in a "run `<playbook>`" request |
| `<name>/actions/...` | **On invocation** — named in a "run `<action>`" request |
| `<name>/corpus/...` | **On-demand** — reached via the forward-link from a paired guide |
| `<name>/*.md` (flat) | Ambient — short policies (e.g., a vendor list) without an explicit guides/ subtree |

Policy guides participate in the drift sensor the same way project guides do. The agent doesn't distinguish "where the rule came from" when applying it; the distinction matters for authorship and update flow, not enforcement.

## Override model — Org → Team → Project cascade

Each lower tier can override the same-basename item from any tier above it — **unless** a higher tier marks that item strict. Cascade order is: **project beats team beats org**.

For any `<kind>/<name>` (kind ∈ `guides`, `playbooks`, `actions`), the file that wins at runtime is the highest-priority one that exists:

| Tier | File location |
|---|---|
| Project (leaf) | `harness/<kind>/<name>.md` |
| Team policy | `harness/policies/<team-policy>/<kind>/<name>.md` (manifest `tier: team`) |
| Org policy | `harness/policies/<org-policy>/<kind>/<name>.md` (manifest `tier: org`) |

**Corpus does not cascade.** Corpus is loaded on-demand by forward-link from a guide — every guide links to its own intended corpus, so corpus from different tiers coexists without collision. `corpus` is never strict-able.

### `strict` — block override from below

Any policy tier can lock specific items against override from below. Declare it in the manifest:

```yaml
# keystone-policy.yaml
name: acme
version: 1.0.0
tier: org              # or `team`
strict:
  guides:
    - documentation
  playbooks:
    - trunk_based_development
  actions:
    - rubocop_for_ruby
```

Rules:

- Each key is a kind (`guides`, `playbooks`, `actions`); under it a list of item basenames.
- `corpus` is **not** strict-able.
- Default is `strict: {}` (empty). Nothing is strict unless declared.
- **Org-strict** blocks both team and project overrides of that item.
- **Team-strict** blocks project overrides only (org can still override the team's item from above — but at install time `policy verify` catches a team policy attempting to violate an org strict).
- `keystone policy verify` (run after `policy add` / `policy update` / on demand) errors if any lower tier shadows a strict item.

Non-strict items remain freely overridable down the chain.

### `required` — declare items the project must define

A policy can flag items that should exist somewhere in the cascade but which the policy itself does not ship. The intent is to surface gaps — things the project needs to define because no higher tier has prescribed them:

```yaml
# keystone-policy.yaml
name: acme
tier: org
required:
  guides:
    - business_continuity_plan
  actions:
    - oncall_rotation_handoff
```

Rules:

- Same structure as `strict` (kind → list of item basenames).
- `corpus` is not required-able (same rationale as strict).
- An item satisfies `required` when any tier — project, team, or another org policy — has a file at the matching path.
- `keystone policy verify` reports unmet required items as advisory **gaps**, not hard errors. They are listed under "needs to be defined" so the project can fill them in.

Solo projects with no org/team policies installed have no required items by definition — the harness works for a single project without any policies present.

## Authorship

Policies are **not** project-authored. The `universal/` policy is owned by keystone; named policies are owned by the org or vendor that published the source repo. Consumers consume — they `init --policy`, `policy add`, and `policy update`, but they don't edit policy files in place. Local edits block `policy update` unless `--force` is passed, on the assumption that ad-hoc changes inside a policy namespace are usually mistakes.

If a project needs to soften or extend a non-strict policy item, the override model above is the right tool: drop a file at the matching project path. For strict items, the project records its deviation by raising it back to the org rather than overriding locally.

## Pinning + updates

Each installed policy is pinned in [`../.keystone.lock`](../.keystone.lock):

```yaml
policies:
  acme:
    source_ref: "git+https://github.com/acme/keystone-policy.git#v1.2.0"
    resolved_sha: "abc1234..."
    policy_version: "1.2.0"
    keystone_version: "0.9.0"
    files:
      "harness/policies/acme/guides/vendors.md": "sha256:..."
```

The lockfile records per-file hashes so `keystone policy update <name>` can detect locally edited files before clobbering them.

## Authoring a policy

A policy is a git repo with two things at its root: a `keystone-policy.yaml` manifest and a `policy/` content directory:

```
my-policy-repo/
├── keystone-policy.yaml      # name, version, optional strict map, optional description
├── README.md                 # for humans browsing the repo; ignored at install
└── policy/
    └── harness/policies/<name>/
        ├── corpus/<topic>/<file>.md
        ├── guides/<topic>/<file>.md
        ├── playbooks/<name>.md      # optional
        └── actions/<name>.md        # optional
```

The installer enforces that every file under `policy/` lives within the policy's own namespace (`policy/harness/policies/<name>/`). Files outside that prefix are an error — keeps policies from accidentally writing into project trees.
