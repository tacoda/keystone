# Host support matrix

How Keystone 5.0 reaches each host, per primitive kind. Every cell was checked
against the host's own docs (or source, where marked) on **2026-11-05**. Hosts
change fast: re-verify a row before changing its adapter, and update the date.

Strategies:

- **native**: the host reads the canonical file in `AGENTS.md` or `.agents/` as-is. Keystone writes nothing.
- **symlink**: the host reads the same format at another path. Keystone links each entry.
- **shim**: the host needs a different format or file. Keystone generates it.
- **index**: the host has no equivalent mechanism. `AGENTS.md` lists the primitive with its path, so the agent opens it on demand.
- **report**: a constraint the host cannot enforce. `keystone project` prints it, and `AGENTS.md` lists it under "not enforced on this host".

`(unverified)` marks a fact the docs did not settle. Until an unverified cell
is confirmed, Keystone treats it as the safer strategy (shim instead of symlink,
or index/report instead of shim).

## Matrix

| Host | Root | Rules | Skills | Commands | Subagents | Hooks | Permissions | Tools (MCP) |
|---|---|---|---|---|---|---|---|---|
| claude-code | shim `CLAUDE.md` (`@AGENTS.md`) | shim `.claude/rules/` | symlink `.claude/skills/` | shim `.claude/commands/` | shim `.claude/agents/` | shim `.claude/settings.json` | shim `.claude/settings.json` | shim `.mcp.json` |
| codex | native | index | native | index | shim `.codex/agents/*.toml` | shim `.codex/hooks.json` | shim `.codex/rules/*.rules` + report | shim `.codex/config.toml` |
| cursor | native | shim `.cursor/rules/*.mdc` | native | shim `.cursor/skills/` (user-invoked) | shim `.cursor/agents/` | shim `.cursor/hooks.json` | shim `.cursor/cli.json` + report | shim `.cursor/mcp.json` |
| github-copilot | native | shim `.github/instructions/` | native | shim `.github/prompts/` | shim `.github/agents/` | shim `.github/hooks/` | report (no repo file) | shim `.mcp.json` |
| opencode | native | index | native | shim `.opencode/commands/` | shim `.opencode/agents/` | shim `.opencode/plugins/` | shim `opencode.json` | shim `opencode.json` |
| continue | native (IDE: root only) | shim `.continue/rules/` | shim (copy) `.claude/skills/` | shim `.continue/prompts/` | index | report | report (user-level only) | shim `.continue/mcpServers/` |
| cline | native | shim (copy) `.clinerules/` | shim (copy) `.cline/skills/` | index (format unverified) | index (format unverified) | report (format unverified) | report | report (global only) |
| goose | native | index | native | index | native `.agents/agents/` | shim `.agents/plugins/keystone/hooks/hooks.json` | report (global only) | report (global only) |
| pi | native | index | native | shim `.pi/prompts/` | index (none) | shim `.pi/extensions/` | report (none) | shim `.pi/mcp.json` |
| aider | shim `.aider.conf.yml` `read:` | index | index | index | index | partial: `lint-cmd` / `test-cmd` | report (none) | report (none) |

## Findings that shape the design

1. **Rule frontmatter uses `paths:`.** Claude Code reads only `paths` from
   `.claude/rules/*.md` and silently ignores other keys. Cline uses `paths` too.
   With `paths` as the canonical key, Claude Code gets a symlink and Cline a
   plain copy, with no format change for either.
2. **Several hosts read more than one skills directory.** Cursor, Copilot,
   opencode, and Goose read both `.agents/skills/` and `.claude/skills/`. If
   Keystone links `.claude/skills/` for Claude Code, those hosts may load each
   skill twice. How each host dedupes is unverified. The projection planner
   therefore computes the **smallest set of paths** that reaches every enabled
   host exactly once, and the Phase 9 smoke test checks for duplicates.
3. **Copilot reads `.claude/settings.json` hooks.** If both `claude-code` and
   `github-copilot` are enabled, a hook can fire twice. The same planner rule
   applies.
4. **Hook failure defaults differ.** A crashing Cursor hook fails open unless
   it sets `failClosed: true`. A crashing Copilot `preToolUse` command fails
   closed. Keystone sets `failClosed: true` on every blocking Cursor hook.
5. **Codex skips `prompt` and `agent` hook types.** Keystone projects them as
   `command` hooks that call `keystone hook fire`, which runs the inferential
   check itself. Codex runs project hooks, rules, and config only in trusted
   projects.
6. **Repo-level permissions are often missing.** Copilot, Continue, Goose,
   Cline, Pi, and Aider have no committed allow/deny file. On hosts that do
   have blocking hooks (Copilot, Pi), Keystone can enforce `deny` rules through
   a `preToolUse` hook bridge. On the rest the rule is reported.
7. **Permission evaluation order differs.** Claude Code: deny, then ask, then
   allow. Codex: most restrictive wins. opencode: **last match wins**, so the
   shim must write `deny` patterns last. Cursor CLI has no `ask`.
8. **Commands are being folded into skills.** Claude Code merged them, Cursor
   ships `/migrate-to-skills`, and Codex deprecated custom prompts. Keystone
   keeps `command` as a kind, and projects it as a user-invoked skill
   (`disable-model-invocation: true`) where a host has no command format.
9. **Claude Code reads `AGENTS.md` only when no `CLAUDE.md` exists.** Keystone
   keeps the `CLAUDE.md` shim with `@AGENTS.md`, because most repos already
   have a `CLAUDE.md`.
10. **The Agent Skills validator rejects unknown frontmatter keys.** Canonical
    `.agents/skills/*/SKILL.md` must carry spec fields only. Host-specific keys
    (`disable-model-invocation`, `paths`, `argument-hint`) appear only in
    shims, never in the canonical file.

## Sources

Specs:
- Agent Skills: https://agentskills.io/specification, validator at
  https://github.com/agentskills/agentskills (`skills-ref/src/skills_ref/validator.py`),
  client guide at https://agentskills.io/client-implementation/adding-skills-support.md
- AGENTS.md: https://agents.md/ (stewarded by the Agentic AI Foundation, Linux Foundation)

Hosts:
- claude-code: https://code.claude.com/docs/en/ (`memory`, `skills`, `sub-agents`, `hooks`, `permissions`, `mcp`)
- codex: https://developers.openai.com/codex/ (`guides/agents-md`, `skills`, `custom-prompts`, `subagents`, `hooks`, `config-reference`, `rules`, `mcp`)
- cursor: https://cursor.com/docs/ (`context/rules`, `context/skills`, `context/subagents`, `agent/hooks`, `cli/reference/permissions`, `context/mcp`)
- github-copilot: https://docs.github.com/en/copilot/ (`reference/hooks-reference`, `reference/custom-agents-configuration`, `reference/copilot-cli-reference/cli-config-dir-reference`, `how-tos/configure-custom-instructions/add-repository-instructions`) and https://code.visualstudio.com/docs/copilot/customization/
- opencode: https://opencode.ai/docs/ (`rules`, `skills`, `commands`, `agents`, `plugins`, `permissions`, `mcp-servers`). v1 docs; a v2 banner was present and v2 is not yet checked.
- continue: https://docs.continue.dev/ (`customize/deep-dives/rules`, `prompts`, `mcp`, `cli/tool-permissions`). Root-file and skills behavior from source: `core/config/markdown/loadMarkdownRules.ts`, `loadMarkdownSkills.ts` in github.com/continuedev/continue.
- cline: https://docs.cline.bot/ (`customization/cline-rules`, `customization/skills`, `getting-started/config`, `features/subagents`, `mcp/mcp-overview`)
- goose: https://block.github.io/goose/docs/ (`using-goosehints`, `using-skills`, `custom-agents`, `hooks`, `goose-permissions`). Repo now at github.com/aaif-goose/goose.
- pi: https://github.com/earendil-works/pi (`packages/coding-agent/docs/`: `configuration`, `skills`, `prompt-templates`, `extensions`, `security`, `mcp`)
- aider: https://aider.chat/docs/ (`usage/conventions`, `config/aider_conf`, `usage/lint-test`)

Phase 3 correction to the claude-code row: rules, commands, and subagents
are shims, not symlinks. Claude Code subagents require `name:` (canonical
subagents carry `id:`), commands map `args:` to `argument-hint:`, and rule
shims keep only the iron-law / golden-rule / rules / anti-pattern sections
to save always-loaded context. Skills are byte-identical, so they link.

## Open verification items

- Symlink following: confirmed for Claude Code skill directories and Codex.
  Unverified for Claude Code rule, command, and subagent files, and for
  Cursor, Copilot, opencode, Cline, Goose, Pi. Continue CLI likely skips
  symlinked skill dirs (`Dirent.isDirectory()`), so Continue gets copies.
- Whether Copilot maps PascalCase event names from `.claude/settings.json`.
- Duplicate-name handling when a host sees the same skill in two directories.
- Cline workflow, subagent, and hook file formats.
- opencode v2 changes.
