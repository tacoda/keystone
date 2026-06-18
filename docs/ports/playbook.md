# Port: Playbook

**Activation:** Invoked by name. Runs an ordered sequence of actions.
**Purpose:** Compose actions into a workflow. Playbooks own *which actions in what order*; actions own *what each action does*.

## Path convention

```
harness/playbooks/<name>.md                                   # project-owned
harness/policies/<policy>/playbooks/<name>.md                  # policy-owned (read-only)
```

Flat — no topic directory. Playbooks are global by name across the cascade.

## Required shape

```markdown
# Playbook: <name>

<one-sentence description>

## Sequence
1. **<action-name>** — <why this step>
2. **<action-name>** — <why this step>
3. ...

## Halt conditions
<when the playbook stops early — e.g., a sensor in a step returns fail>
```

- **H1 title** — required. Format: `# Playbook: <name>`.
- **Sequence** — required. Each step names an existing action by its `<name>`.
- **Halt conditions** — required. What stops the playbook between steps.
- **Frontmatter** — none required.

## Cascade behavior

Same as other ports: project wins by default; among policies, policies nested deeper in `keystone.json` refine the outer policies they're nested in; a `strict.playbooks: [<name>]` declaration locks the item absolutely — nothing else (project or any other policy) can override it. Exactly one file loads per `<name>`.

A playbook references actions by name; if a referenced action is missing from the cascade, `keystone verify` reports a broken reference at install time.

## Example

```markdown
# Playbook: task

Drive one unit of work end-to-end, from spec through review.

## Sequence
1. **spec** — capture what the task is and pin acceptance criteria.
2. **orient** — identify the files and code paths the change will touch.
3. **verify** — confirm the change works against the acceptance criteria.
4. **review** — surface issues before merge.

## Halt conditions
- `spec`'s acceptance criteria are not testable → escalate to the user.
- `verify` fails repeatedly with the same root cause → halt and report.
```

## Authoring

```
keystone new playbook <name>
```
