# Model-View-ViewModel (MVVM) — rules

The rules from [`corpus/mvvm.md`](../corpus/mvvm.md).
Loaded ambient; enforced by the [drift sensor](../../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**The view binds to the ViewModel and only the ViewModel.** Views that import the model directly have lost the seam that MVVM exists to preserve. The view is dumb; the ViewModel is the only thing that knows the domain exists.

## GOLDEN RULES

- **Aim for ViewModels that are unit-testable without the UI.** A ViewModel test that needs to instantiate a window has lost the point.
- **Aim for commands, not events.** A command is a named user intent (`SaveCommand`, `RefreshCommand`); a click handler on a button is an event. The ViewModel exposes the former.
- **Aim for one ViewModel per view, by default.** Sharing ViewModels across views couples views to each other through the ViewModel's shape.
- **Aim for the ViewModel to mirror the *view's* shape, not the model's.** It is a projection — it can flatten, derive, format, or aggregate. Resist the urge to make the ViewModel "just expose the model fields."

---

Traces to: [`corpus/mvvm.md`](../corpus/mvvm.md).
