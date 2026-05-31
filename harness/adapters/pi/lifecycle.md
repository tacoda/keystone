# pi.dev — Lifecycle binding

How each abstract lifecycle action is invoked in pi.dev. pi has rich extensibility — prompt templates for slash commands, skills as on-demand capabilities, project-scoped JSON settings, and shell execution.

## Action → invocation

Pi's slash commands come from **prompt templates** (a `.md` file in `.pi/prompts/` becomes `/<filename>`) or **skills** (a `SKILL.md` in `.pi/skills/<name>/` becomes `/skill:<name>`). Keystone uses prompt templates for lifecycle actions because they're lighter-weight than skills.

| Action | Invocation | File |
|---|---|---|
| **spec** | `/spec [<tracker-card-id>]` | `.pi/prompts/spec.md` |
| **orient** | `/orient` | `.pi/prompts/orient.md` |
| **check-drift** | `/check-drift` | `.pi/prompts/check-drift.md` |
| **verify** | `/verify` | `.pi/prompts/verify.md` |
| **review** | `/review` | `.pi/prompts/review.md` |
| **learn** | `/learn` | `.pi/prompts/learn.md` |
| **bootstrap** | `/bootstrap` | `.pi/prompts/bootstrap.md` (one-time) |
| **audit** | `/audit` | `.pi/prompts/audit.md` |
| **synthesize** | `/synthesize` | `.pi/prompts/synthesize.md` |
| **mode** | `/mode <paired\|solo\|autopilot>` | `.pi/prompts/mode.md` |

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
