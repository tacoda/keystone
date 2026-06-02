# pi.dev — Lifecycle binding

How each abstract lifecycle action is invoked in pi.dev. pi has rich extensibility — prompt templates for slash commands, skills as on-demand capabilities, project-scoped JSON settings, and shell execution.

## Action → invocation

Pi's slash commands come from **prompt templates** (a `.md` file in `.pi/prompts/` becomes `/<filename>`) or **skills** (a `SKILL.md` in `.pi/skills/<name>/` becomes `/skill:<name>`). Keystone uses prompt templates for lifecycle actions because they're lighter-weight than skills. The filename drives the command name, so the harness ships `keystone-<action>.md` prompt files — pi exposes them as `/keystone-<action>` (hyphen, not colon, because colons aren't filesystem-safe everywhere).

| Action | Invocation | File |
|---|---|---|
| **spec** | `/keystone-spec [<tracker-card-id>]` | `.pi/prompts/keystone-spec.md` |
| **orient** | `/keystone-orient` | `.pi/prompts/keystone-orient.md` |
| **check-drift** | `/keystone-check-drift` | `.pi/prompts/keystone-check-drift.md` |
| **verify** | `/keystone-verify` | `.pi/prompts/keystone-verify.md` |
| **review** | `/keystone-review` | `.pi/prompts/keystone-review.md` |
| **learn** | `/keystone-learn` | `.pi/prompts/keystone-learn.md` |
| **bootstrap** | `/keystone-bootstrap` | `.pi/prompts/keystone-bootstrap.md` (one-time; also inventories computational guides into `guides/computational/` and classifies sensors by kind) |
| **audit** | `/keystone-audit` | `.pi/prompts/keystone-audit.md` |
| **synthesize** | `/keystone-synthesize` | `.pi/prompts/keystone-synthesize.md` |
| **mode** | `/keystone-mode <paired\|solo\|autopilot>` | `.pi/prompts/keystone-mode.md` |

All prompt templates live in `targets/pi/.pi/prompts/` in this repo. The installer drops them into the consumer's `.pi/prompts/`.

## Capability matrix

| Capability | Supported? |
|---|---|
| Autonomous shell execution | ✓ (pi runs shell commands directly) |
| Sub-agent parallelism | ✗ built-in (workaround: spawn pi instances via tmux, or extensions) |
| Autonomy levels (paired/solo/autopilot) | partial — pi doesn't model autonomy, but the corpus' `process/modes.md` is still readable; the agent reads it and adjusts behavior |
| Lazy-by-region | ✗ (no glob-based loading; **orient** reads `state/CODEBASE_STATE.md` and loads matching idioms by hand) |
| Context-reset primitive | `/compact` and session branching |
| Project-scoped settings | ✓ (`.pi/settings.json`) |

## Sub-agent degradation

The **review** action would ideally spawn `review-functional` and `review-security` in parallel. Pi doesn't support this natively. Two workable approaches:

1. **Sequential** (default) — pi runs the review prompt template twice, once per agent role, and combines findings.
2. **tmux fan-out** — power users can launch parallel pi instances in tmux panes, each running one reviewer prompt, then merge results.

The shipped `.pi/prompts/review.md` template defaults to sequential.

## Session and branching

Pi's tree-structured session history with branching is a unique fit for the spec → planning → implementation → verification flow: each phase can be a branch point, and the user can return to a prior branch if a downstream phase reveals the plan was wrong.
