# Policies

**Level 3 extensions.** Policies are distributable bundles of governance content that ship alongside the project harness. Each policy lives in its own namespace under `policies/` and is treated as a unit — installed, updated, and pinned together.

There are two kinds:

- **`universal/`** — the default policy. Ships embedded in the keystone binary. Contains the universal engineering principles (reasoning + rule extracts) that apply to every Keystone install regardless of stack, domain, or org.
- **`<name>/`** — org-authored policies installed via `keystone init --policy <ref>`, `keystone policy add <ref>`, or `keystone policy update`. Vendor lists, license rules, release gates, compliance regimes, internal coding standards — anything an org wants to push down across many projects.

## Layout inside a policy

Each policy mirrors the project layout — `corpus/` for reasoning, `guides/` for rules — but scoped to that policy's namespace:

```
harness/policies/<name>/
├── corpus/              # on-demand reasoning (reached via guide forward-link)
│   └── <topic>/<file>.md
└── guides/              # ambient rules — always loaded
    └── <topic>/<file>.md
```

Sensors are **not** part of a policy. Sensors describe project tooling (lint, type-check, test commands) — they're an integration concern that the project owns. A policy may declare *what* must be checked; the project decides *how*.

## Activation

Sub-paths inside a policy determine activation:

| Path | Activation |
|---|---|
| `<name>/guides/...` | **Ambient** — always loaded, same as project guides |
| `<name>/corpus/...` | **On-demand** — reached via the forward-link from a paired guide |
| `<name>/*.md` (flat) | Ambient — short policies (e.g., a vendor list) without an explicit guides/ subtree |

Policy guides participate in the drift sensor the same way project guides do. The agent doesn't distinguish "where the rule came from" when applying it; the distinction matters for authorship and update flow, not enforcement.

## Authorship

Policies are **not** project-authored. The `universal/` policy is owned by keystone; named policies are owned by the org or vendor that published the source repo. Consumers consume — they `init --policy`, `policy add`, and `policy update`, but they don't edit policy files in place. Local edits block `policy update` unless `--force` is passed, on the assumption that ad-hoc changes inside a policy namespace are usually mistakes.

If a project needs to soften or extend a policy, the right move is at the project layer — add a project guide that explicitly traces to the policy file with reasoning. The policy stays unmodified; the project's deviation is recorded where future readers will look.

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
├── keystone-policy.yaml      # name, version, optional description
├── README.md                 # for humans browsing the repo; ignored at install
└── policy/
    └── harness/policies/<name>/
        ├── corpus/<topic>/<file>.md
        └── guides/<topic>/<file>.md
```

The installer enforces that every file under `policy/` lives within the policy's own namespace (`policy/harness/policies/<name>/`). Files outside that prefix are an error — keeps policies from accidentally writing into project trees.
