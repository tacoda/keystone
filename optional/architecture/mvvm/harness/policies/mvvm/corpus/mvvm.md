# Model-View-ViewModel (MVVM)

Articulated by John Gossman at Microsoft in 2005 for WPF and Silverlight; a refinement of [[mvc]] and Martin Fowler's *Presentation Model* (2004) for platforms with rich **data binding**. MVVM's central trick is the *ViewModel*: a stateful object whose properties are bound declaratively to the view, eliminating the imperative "controller sets values on widgets" code that fills traditional MVC.

The pattern now dominates client-side architectures with reactive UI frameworks: SwiftUI, Vue, Knockout, modern Android (Jetpack), and any system where the UI re-renders automatically when state changes.

> **Rules extracted:** [`guides/mvvm.md`](../guides/mvvm.md). This file holds the full reasoning, anti-patterns, and references.

## The three roles

- **Model** — the domain data and rules. Same as in [[mvc]]; the principles in `mvc.md` apply.
- **View** — the UI. Declares how each piece of ViewModel state is displayed. Declarative bindings, not imperative updates.
- **ViewModel** — a presentation-shaped projection of the model, plus the commands the view can trigger. The ViewModel exposes properties; the view binds to them; the framework propagates changes automatically.

The defining property: the **view has no reference to the model**. The view binds only to the ViewModel. The ViewModel does the translation. This separation is what makes the view trivially mockable and the ViewModel unit-testable as plain code.

## Data binding is the architecture

What distinguishes MVVM from MVC is not the three letters — it is the assumption that *binding is automatic*. The framework synchronizes view state with ViewModel state in both directions. The author writes ViewModel properties; the view notices. The user types; the ViewModel notices. The imperative `widget.setText(model.name)` of MVC disappears.

This is why MVVM thrives in WPF, SwiftUI, and Vue — frameworks that ship with rich binding primitives — and is awkward outside them. On a platform without binding, MVVM degenerates into MVC with more types.

## What it asks of you

- When you write a view, bind it to ViewModel properties only. The view should not import the model. See [[information-hiding]].
- When you write a ViewModel, it should be testable as a plain object: no framework imports, no UI types in the test. The ViewModel changing produces visible behavior only when wired to a real view; tests assert on the ViewModel's state and the commands it exposes.
- When you find domain logic in a ViewModel, push it down to the model. The ViewModel translates and presents; it does not decide business rules. See [[separation-of-concerns]].
- When a ViewModel grows large, ask whether two ViewModels are sharing one class. The split is usually obvious from which properties are touched by which user actions.

## Anti-patterns

- A "ViewModel" that is the model with a different name. The translation layer has been elided; refactor the model is now the view-bound type.
- Views with code-behind logic that mutates the ViewModel imperatively. The binding has been bypassed.
- ViewModels with framework imports (`UIKit`, `WPF`, `View`, `Window`). The ViewModel is no longer testable as plain code.
- Two-way binding to a property whose setter has business effects (validation, persistence). The binding looks innocent; the side effects fire on every keystroke. See [[fail-fast]].
- Property-change events that cause cascades of further property changes within the same ViewModel, leading to update storms. The dataflow has become non-obvious.

## References

- Gossman, J. (2005). "Introduction to Model/View/ViewModel pattern for building WPF apps." Microsoft Developer Network blog.
- Fowler, M. (2004). "Presentation Model." martinfowler.com/eaaDev/PresentationModel.html. (The pattern MVVM specializes for data-binding platforms.)
- Smith, J. (2009). "WPF Apps With The Model-View-ViewModel Design Pattern." *MSDN Magazine*, February 2009.
- Likness, J. (2014). *Designing Silverlight Business Applications*. Addison-Wesley. (Detailed MVVM treatment for enterprise apps.)
