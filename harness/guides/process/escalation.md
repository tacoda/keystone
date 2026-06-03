# Escalation

When the agent is stuck, when to stop and ask the user — versus when to keep going. Counterpart to [`modes.md`](modes.md): modes set the *default* pace; escalation is the override that triggers regardless of mode.

## GOLDEN RULES

- **Aim to stop after three failed attempts at the same objective.** Three iterations on the same failing sensor, the same red test, or the same compile error without making it green is a signal that the model of the problem is wrong. Stop, report what was tried, and hand back to the user. Different errors after each attempt count as progress; the same error three times does not.
- **Aim to escalate early when the cost of being wrong is high.** Security, data integrity, prod state — escalate at the first hint of uncertainty, not after three attempts.
- **Aim to escalate with a recommendation, not a question.** "I think A because X — should I do A?" beats "what should I do?"

## RULES

- **Stop on contradictory rules.** Two guides say opposite things → surface the conflict, do not pick.
- **Stop on ambiguous acceptance criteria.** A criterion that admits two readings is not a criterion; return to spec.
- **Stop on a non-trivial-consequences decision with no clear winner.** Two equally plausible designs → the user decides.
- **Stop when a [[dangerous-actions|dangerous action]] is the only path forward.** Confirm before, not after.
- **Stop when the codebase contradicts state.** If `corpus/state/CODEBASE_STATE.md` says one thing and the code says another, do not assume which is right.

## What "stop and report" looks like

A structured handoff:

```
STUCK: <one-sentence summary>

What I tried:
1. <attempt> → <result>
2. <attempt> → <result>
3. <attempt> → <result>

What I think is going on:
<one paragraph: best hypothesis>

What I need from you:
<a specific question, not "what should I do"?>
```

The report is the deliverable. A "raise a hand" with no diagnosis is not escalation; it is giving up.

## Pacing modes

- **paired** — escalation triggers conversation immediately.
- **solo** — escalation triggers the structured-handoff report; agent waits.
- **autopilot** — escalation still pauses the loop. Autopilot does not override the stop conditions. Assumption-log entries are not a substitute for stopping when stuck.

## Anti-patterns

- "Trying one more thing" four times in a row.
- Reporting "I am stuck" with no diagnosis.
- Picking a direction on a high-stakes decision because the model is uncertain. Uncertainty is the signal to ask, not to guess.
- Treating a different-flavor failure on each attempt as "the same problem." Different errors = progress; report progress and continue.

---

See also: [`modes.md`](modes.md), [[dangerous-actions]].
