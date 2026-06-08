# Premature Optimization — rules

The non-negotiable and strongly-preferred rules from [`corpus/principles/premature-optimization.md`](../../corpus/principles/premature-optimization.md).
Loaded ambient; enforced by the [drift sensor](../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**Measure before you optimize, and measure after.** Without a before-and-after measurement against a realistic workload, an "optimization" is a guess. Guesses sometimes guess right; the discipline is to know which.

## GOLDEN RULES

- **Aim for the simplest implementation first.** It is the baseline against which any optimization must justify itself. See [[simplicity]], [[simple-design]].
- **Aim to optimize the architecture, not the lines.** A correct algorithm in clear code beats a clever algorithm in unreadable code at almost every scale.
- **Aim to record the measurement.** "We tried X, it was N% faster on workload Y" beats "X is faster" — and beats it again next year when someone questions the choice.

---

Traces to: [`corpus/principles/premature-optimization.md`](../../corpus/principles/premature-optimization.md).
