# synthesize

**Promote learning-inbox candidates into the right corpus or guide layer.** Read [`harness/learning/README.md`](learning/README.md).

## Activities

For each file in `harness/learning/inbox/`:

1. **Read the candidate.** Note its `proposed-layer` frontmatter.
2. **Decide the destination:**
   - Universal principle → `harness/policies/universal/corpus/principles/<name>.md` and `harness/policies/universal/guides/principles/<name>.md`
   - Stack idiom → `harness/corpus/idioms/<stack>/<name>.md` and `harness/guides/idioms/<stack>/<name>.md`
   - Process rule → `harness/guides/process/<name>.md`
   - Sensor → `harness/sensors/<name>.md`
3. **Author the corpus and guide pair.** Corpus explains the *why*; guide states the rule.
4. **Choose the rule tier.** Default: regular rule. Iron law / golden rule only when deviation is genuinely non-negotiable.
5. **Move the candidate** from `inbox/` to `promoted/` (kept), or `rejected/` (with a one-line reason in the moved file).

## Iron law

**No silent overwrites.** If a destination already exists, propose a diff and let the user merge.

## Gate

Synthesize is where the harness changes shape. Show the user every promotion and removal before applying. After synthesize writes to `guides/`, the active session has stale rules in context — recommend a context reset.
