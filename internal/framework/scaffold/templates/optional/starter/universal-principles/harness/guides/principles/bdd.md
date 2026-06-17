# Behavior-Driven Development — rules

The non-negotiable and strongly-preferred rules from [`corpus/principles/bdd.md`](../../corpus/principles/bdd.md).
Loaded ambient; enforced by the [drift sensor](../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**The test name is a promise to the next reader.** A test named for a behavior commits to specifying that behavior, and only that behavior. A test that asserts five unrelated things under one behavior name has broken the promise — split it. A test named *test_method_2* has made no promise at all, which is worse.

## GOLDEN PATH

- **Aim for tests that read aloud as English.** If the test cannot be read to a non-engineer, it is not yet a specification — it is still a test.
- **Aim for examples produced *with* the people who care about the behavior, not for them.** A spec written alone is a spec rewritten under pressure.
- **Aim for the system's vocabulary to match the business vocabulary.** Each translation between them is a place for the truth to drift.
- **Aim for examples that are the regression suite.** When the same artifacts serve discovery, verification, and prevention-of-regression, the cost of maintenance drops to one suite from three.

---

Traces to: [`corpus/principles/bdd.md`](../../corpus/principles/bdd.md).
