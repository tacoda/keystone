# Product Shape

> **Template.** Replace with your project's actual content. The **bootstrap** action will offer to fill this in from your README, manifest, and git history.

## What this project is

One-paragraph statement of the product. What does it do? Who is it for? What problem does it solve?

## What it ships

- The deliverable (binary, library, service, plugin, document set, etc.).
- The distribution channel (package manager, container registry, marketplace, etc.).
- The supported platforms / runtimes / agents.

## What it does NOT ship

- Things that are explicitly out of scope.
- Adjacent problems the project deliberately does not solve.

## What survives a release

- Backwards-compatibility commitments.
- The public interface (API, CLI, file format, plugin contract).
- What consumers can rely on between versions.

## What does NOT survive a release

- Internal implementation details.
- Files / paths / structures that consumers shouldn't depend on.
- Anything in `internal/`, `private/`, or undocumented.

## Invariants

Business rules that must always hold. Examples:

- "A user cannot be both an admin and a guest at the same time."
- "Once an order is shipped, its line items are immutable."
- "The plugin must not write outside the consumer's project directory."

These are the constraints that agent-generated code must respect, regardless of whether they're enforced by tests.
