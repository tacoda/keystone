# Grounding

The discipline of verifying that what the agent references — functions, APIs, packages, files, flags — actually exists before invoking it.

## GOLDEN RULES

- **Aim to grep before invoking.** Before calling a function or method, confirm it exists at the expected path with the expected signature. The agent's training data is not the codebase.
- **Aim to read the manifest before importing.** Before importing a package, confirm it is in the project's dependency list. If it is not, that is a [[dependencies|new-dependency decision]], not a paper-over.
- **Aim to check the binary's `--help` before passing a flag.** Documentation drifts; the binary is the truth.

## RULES

- **No invented imports.** A package the agent does not see in the manifest does not exist for this project. Adding it is a separate, approved decision.
- **No invented methods.** If the agent suggests `Foo.bar()`, the agent has read a file showing `bar` is a real method on `Foo`. Otherwise: grep first.
- **No invented config keys.** Configuration files have schemas — read them before setting a key.
- **No invented CLI flags.** `--no-verify` exists; `--no-gpg-sign` exists; `--really-just-this-once` does not. Check first.
- **No "this should work" guesses on error handling.** If the agent does not know what exception a function raises, read it. The exception class is data, not opinion.

## Why this is agent-specific

A coding agent's training data includes millions of codebases. Patterns from one bleed into another. The agent will *plausibly* suggest a method that exists in five popular libraries — and does not exist in the one this project uses. Grounding is the discipline of refusing to act on plausibility.

The remedy is mechanical: before any reference, grep the codebase, read the import, read the binary's help, read the schema. The check costs a tool call; the bug from skipping it costs a debugger session and a reverted PR.

## Sensors

- The **type-check sensor** catches a large class of grounding failures (wrong function signature, wrong type, undefined import). Type-check failures during implementation are a grounding warning, not just a syntax warning.
- The **drift sensor** flags imports that do not match the manifest, when the bootstrap action has wired this up.

## Anti-patterns

- "I'm pretty sure this method exists" — replace "pretty sure" with `grep`.
- Importing `lodash` because it is "probably available" — check the manifest.
- Passing `--force` to an unfamiliar CLI without checking the help text.
- Writing a try/catch for `SomeException` without confirming the function raises it.
- Catching a generic `Exception` because the specific class is unknown — find the specific class first.

---

See also: [[dependencies]], [[escalation]], [[context-budget]].
