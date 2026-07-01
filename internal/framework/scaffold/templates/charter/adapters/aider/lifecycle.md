# Aider — Lifecycle binding

How each abstract lifecycle action is invoked in Aider.

## Invocation

Every action is invoked via natural language in the chat: "run task on TICKET-123," "run verify," "do a review pass." The agent reads `CONVENTIONS.md` (loaded via the `.aider.conf.yml` `read:` list) at session start, finds the action in the bulleted list, follows the link to `charter/actions/<action>.md`, and executes the playbook. Aider has no slash-command primitive for project actions (its `/<command>` set is for shell, git, and chat control), so natural-language invocation is the only path — and that's exactly what `charter/actions/` is designed for.

The canonical kickoff phrase is **"run task on `<ticket-id>`"** (or "run the task workflow") — `charter/actions/task.md` orchestrates `spec → orient → implementation → check-drift → verify → review`.

## Markdown is the native shape

Aider's value is that it works straightforwardly with markdown — reading a playbook and following its activities is its native pattern. The action playbooks in `charter/actions/` are exactly that. The corollary: keep them clear and executable. Aider has no slash-command layer to paper over a vague playbook.

## Sub-agent support

None. Aider runs a single conversation; the **review** action runs each review concern sequentially over the diff in the same session.

## Modes

Aider has no autonomy levels in the charter sense. There is:

- **Default chat** — Aider applies edits as it suggests them; you can undo via `/undo`.
- **`/architect`** mode — Aider explains intent first and asks before applying. Useful for the spec and planning phases.
- **`/ask`** mode — Aider answers questions without editing. Useful for review.

The charter's `paired`/`solo`/`autopilot` modes collapse to **paired** in practice — Aider applies edits immediately; the user reviews each diff in `/undo` reach.

## Capability matrix

| Capability | Supported? |
|---|---|
| Autonomous shell execution | ✓ (`/run <cmd>` and `/test`) |
| Sub-agent parallelism | ✗ |
| Autonomy levels (paired/solo/autopilot) | ✗ — effectively paired |
| Lazy-by-region | ✗ — Aider reads what the user passes via `--read` or includes via `/add`; no auto-load by path glob |
| Context-reset primitive | `/clear` (clear chat history); `/drop` (remove files from context) |
| Tracker integration | none native — user pastes card content |
| GitHub integration | partial (via `gh` in `/run`) |
