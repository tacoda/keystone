# Logging — rules

The non-negotiable and strongly-preferred rules from [`corpus/principles/logging.md`](../../corpus/principles/logging.md). Loaded ambient. Sibling of [[observability]] — observability is the why; logging is one of the hows.

## IRON LAW

**NEVER LOG A SECRET, A CREDENTIAL, OR USER PII.**

Tokens, passwords, session cookies, API keys, full credit-card numbers, government IDs, full email addresses (when avoidable), full request bodies (when they may contain credentials) — none of these go to a log. Once a value lands in a log file, it lands in log aggregation, in backups, in S3, in whatever the team's retention policy is. Treat logs as public.

## GOLDEN RULES

- **Aim for structured logs.** JSON or key-value pairs, not free-form strings. Searchable beats prosaic.
- **Aim for one log line per request boundary**, not one per function call. Logging on the inside of a hot loop is how disks fill.
- **Aim to log decisions, not data.** "Routed request to handler X" is a decision. "Request body: {...}" is data — and is the riskiest kind of log line.
- **Aim to include enough context to act on the line without opening another tool.** Request ID, user ID (hashed if needed), the operation, the outcome.

## RULES

- **Redact at the boundary.** A logger that "knows" to redact secrets is fragile; data that has been redacted before reaching the logger is safe. See [[secrets-management]].
- **Log at level.** `error` for things a human must act on; `warn` for things worth knowing; `info` for the request-shaped story; `debug` for development. No `error` for things that are merely surprising.
- **No `print()` / `console.log()` in committed code.** Use the project logger. Print statements escape redaction.
- **Sample high-volume logs.** A log line that fires a million times an hour is noise. Sample, or remove.
- **Errors carry the cause chain.** A re-raised exception that drops the original is a debugger's enemy. Use `raise ... from`, `Error.cause`, `errors.Wrap`, depending on the language.

## Sensors

- The **drift sensor** flags `print(`, `console.log(`, `fmt.Println(`, `dbg!(` and the project-specific equivalents recorded by bootstrap.
- **Bootstrap** records the logger (e.g., `structlog`, `pino`, `slog`, `tracing`, Monolog) and any PII-redaction middleware in `corpus/state/CODEBASE_STATE.md`.

---

Traces to: [`corpus/principles/logging.md`](../../corpus/principles/logging.md). See also: [[observability]], [[secrets-management]], [[sensitive-files]].
