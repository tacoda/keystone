# Permission

<!-- voice-register: informational -->
<!-- voice-english: ste -->

A permission sets which tool calls the agent may make without asking, must ask about, or must never make. It's a constraint. You write the rules once, in Claude Code syntax, and keystone translates them for each host with a permission file. The cost is uneven reach: some hosts can't express a rule exactly, so keystone reports it instead of enforcing it.

## Example

```markdown
---
id: default
description: Protect secrets and gate pushes.
allow:
  - "Bash(go test:*)"
ask:
  - "Bash(git push:*)"
deny:
  - "Read(./.env*)"
---

# Default permissions

Tests run freely. A push leaves the machine, so it needs a person. Secrets never enter the context.
```

Save it as `.agents/permissions/default.md`. Then run `keystone index && keystone project`.

## Where it lives

```
.agents/permissions/<id>.md
```

Keystone infers `kind: permission` from the `permissions/` folder. You can split the rules over many files, because keystone joins the `allow`, `ask`, and `deny` lists of every file and drops duplicates.

## Frontmatter

| Key | Required | Meaning |
|---|---|---|
| `id` | yes | Stable id |
| `description` | yes | One sentence on what these rules protect |
| `allow` | no | Rules the agent may use without asking |
| `ask` | no | Rules that need the user's approval each time |
| `deny` | no | Rules the agent must never use |
| `tags` | no | Kebab-case labels for `keystone list --tag` |

Each rule is `Tool` or `Tool(specifier)`, as in Claude Code. For `Bash`, a specifier that ends in `:*` is a prefix match, and any other specifier is an exact command.

A 4.x file with `kind: posture` still parses. `keystone lint` warns, and `keystone lint --fix` rewrites it.

## Body

Write why you chose these rules, meaning the trust boundary they encode. Keystone doesn't project the body to any host.

## How each host gets it

- **Claude Code**: the rules merge unchanged into `.claude/settings.json` under `permissions`.
- **Cursor**: the rules merge into `.cursor/cli.json` as `allow` and `deny` tokens.
- **Codex**: keystone owns `.codex/rules/keystone.rules`. Each `Bash` rule becomes a `prefix_rule`, with `ask` as `prompt` and `deny` as `forbidden`.
- **Every other host**: no file. `keystone project` reports each rule.

For each host, the `AGENTS.md` managed block counts the rules it can't enforce. It also tells the agent to follow the `deny` and `ask` lists itself.

### What translates

A translation may make a `deny` or `ask` stricter, but it never makes an `allow` broader than the rule you wrote. So keystone reports these rules instead of writing them:

- **Cursor** has no `ask`. So keystone reports every `ask` rule. A `Bash` rule translates only when it names one command word, and an `allow` also needs the `cmd:*` form. `Read`, `Edit`, and `Write` paths translate only when they're project-relative. `WebFetch(domain:x)` translates.
- **Codex** translates only `Bash` rules, and keystone reports an exact-command `allow`, because a Codex prefix rule would match more than that one command.
- **Keystone reports a `*` inside a `Bash` specifier**, as in `Bash(curl *)`, on both Cursor and Codex.

## Shared files

Claude Code and Cursor share the file with your settings, so keystone records each entry it adds in the lockfile. When you drop a rule from `.agents/`, keystone removes its entry. But if you had the same entry before keystone added it, your entry stays. When you remove a host from `keystone.json` `adapters`, keystone strips its entries from that host's file.

## What a permission does not do

- A permission doesn't block on a host without a permission file. There, it's an instruction the agent can skip.
- Keystone doesn't check a rule against the tool's real behavior. A wrong specifier projects as written.
- The evaluation order is the host's. Claude Code checks `deny`, then `ask`, then `allow`, while Codex applies the most restrictive match.

## Failure modes

- **A rule doesn't apply on a host.** The host can't express it. Read the `keystone project` report, or the count in `AGENTS.md`.
- **A Cursor `allow` is missing.** It used an exact `Bash` command. Use `Bash(cmd:*)` if any arguments are safe to allow.
- **An entry you added by hand disappeared.** Keystone wrote that entry first. So keystone recorded the entry as its own. When the rule left `.agents/`, keystone removed the entry.

## Authoring

```
keystone new permission <id>
```

This writes `.agents/permissions/<id>.md` with one sample rule in each list. Replace the samples with your own. `keystone new posture` exits with an error that names `keystone new permission`.
