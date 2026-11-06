# Logging

A log is a record of what the system did, written for a future reader who is trying to understand a past failure. Logs that cannot be searched, cannot be trusted, or cannot be safely retained are worse than no logs — they cost storage, mislead investigations, and create liability. The discipline of logging is the discipline of producing records that are *useful*, *honest*, and *safe to keep*.

> **Rules extracted:** [`guides/principles/logging.md`](../../guides/principles/logging.md).

## What it asks of you

- Write logs for the next reader, not for the current author. The reader is on-call at 03:00 — give them the answer, not the breadcrumbs.
- Treat logs as a permanent surface. Anything that lands there is durable, replicated, and visible to anyone with log access. That includes attackers who get log access.
- Distinguish *the story of the request* (info) from *the situations that need human attention* (warn, error). A log stream where everything is `info` and everything is `error` carries no signal.
- Redact upstream. The logger is too late to be the last line of defense.

## Why it holds

The CWE catalogue lists "Insertion of Sensitive Information into Log File" (CWE-532) as a recurring class of breach — and the lived record of incidents (T-Mobile 2018, Twitter 2018 password-log, Facebook 2019 password-log) shows it is not a theoretical concern. Once a secret lands in a log, it propagates: hot storage, cold storage, log aggregators, vendor systems, backups. The only safe assumption is that *any value in a log is permanently disclosed*.

The structured-log argument (Majors, Sridharan) is empirical: free-form logs are unsearchable at scale. A platform on which "find every 502 in the last hour, grouped by route" takes thirty minutes of grep is a platform without observability. Structured logs make those queries cheap; cheap queries make incident response fast.

The level discipline (most language-specific style guides) is about producing signal. An `ERROR` log that fires on every input-validation rejection is not an error log — it is a wolf-cry channel, and the team will learn to ignore it.

## Anti-patterns

- `log.info("request body: %s", body)` on a route that takes credentials.
- `log.error()` on every 4xx response. 4xx is the client; errors are about the server.
- "Temporary" `print()` left in shipped code.
- A redaction step that runs *inside* the logger, which can be disabled at runtime by raising the level.
- A log line that says "something went wrong" with no request ID, no user, no operation. Untraceable.
- Logging exceptions and continuing without re-raising or handling. The log is not a substitute for handling the error.

## References

- Majors, C., Fong-Jones, L. & Miranda, G. *Observability Engineering* (2022).
- Sridharan, C. *Distributed Systems Observability* (2018).
- OWASP. [*Logging Cheat Sheet*](https://cheatsheetseries.owasp.org/cheatsheets/Logging_Cheat_Sheet.html).
- CWE-532: *Insertion of Sensitive Information into Log File*.
- Gregg, B. *Systems Performance* — the operational case for structured, low-cardinality logs.

---

Forward link: [`guides/principles/logging.md`](../../guides/principles/logging.md). See also: [`observability.md`](observability.md), [`secrets-management.md`](secrets-management.md).
