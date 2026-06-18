---
created_at: 2026-06-18
tracker_card:
status: pending-acceptance
---

# Adopt gin for the web dashboard + stop expensive handlers from hanging the server

## Intent

Replace the hand-rolled `net/http.ServeMux` router in `internal/framework/web/` with `github.com/gin-gonic/gin`, and stop expensive handlers (notably `/policies/investigate` → `handleInvestigator`) from blocking other dashboard requests while they run.

## Acceptance criteria

### Part 1 — gin adoption

1. `go.mod` declares `github.com/gin-gonic/gin` as a direct dependency and `go.sum` is consistent.
2. The dashboard router is a `*gin.Engine` (or a thin wrapper) and is what `http.Server.Handler` receives in `Serve()`.
3. All existing routes resolve to the same URL paths and produce the same response bodies as today: HTML pages under `/`, JSON under `/api/`, SSE at `/events`, and the existing form-action endpoints under `/web/actions/`.
4. Embedded `assets/*` and `templates/*.html` continue to be served from the `embed.FS`.
5. The existing test suite under `internal/framework/web/` passes unchanged. New tests are added where behavior changes (e.g., timeout/cancellation, concurrency cap).
6. `gin.SetMode(gin.ReleaseMode)` (or equivalent) is set so the dashboard does not emit gin's debug banner; logging stays compatible with how the server currently logs.
7. Per-handler timeout and the `ReadHeaderTimeout` / `IdleTimeout` discipline currently in `server.go` are preserved. The SSE handler remains exempt from a write timeout.

### Part 2 — non-blocking expensive handlers

8. While `/policies/investigate` is mid-flight, an unrelated dashboard request (e.g., `/api/primitives`) returns in well under 1s on a developer machine. (Concretely: a regression test asserts a fast endpoint completes in < 500ms while a slow investigate is held open.)
9. `handleInvestigator` honors client disconnection: cancelling the request stops the underlying work within a bounded time, no goroutines leak.
10. A per-handler upper-bound timeout exists for investigate (default a few minutes, configurable in code constant) and returns a 504-shaped error response (HTML for browser, JSON for `/api/`) if hit.
11. Concurrent investigate requests are bounded by an explicit semaphore/cap; over-cap requests receive a clear 429-style response rather than queueing unbounded.
12. The SSE hub at `/events` keeps pushing updates to other clients while an investigate is running.

## Non-goals

- Rewriting handler bodies beyond what's needed to fit the new router signature and the cancellation/timeout discipline.
- Introducing new endpoints, new UI, or new write paths.
- Replacing the SSE implementation with a different transport (WebSockets, etc.).
- Swapping out the file watcher, primitive layer, or any data source.
- Adding authentication, multi-origin support, or remote binding — the dashboard stays localhost.
- Performance work on other handlers that aren't currently hanging.

## Open questions — answered

1. **Investigate is IO-bound** (external source queries). Plain `context.Context` propagation through the call chain is sufficient; no finer-grained cooperative cancellation needed.
2. **Tests use an internal helper** (not `Serve()`). The router constructor stays callable in-process; tests get re-pointed at the gin engine with minimal churn.
3. **Investigate becomes a full async job.** `POST /policies/investigate` (or the existing endpoint, behavior-preserving on URL) enqueues a job and returns immediately with a job ID. The client polls `GET /policies/investigate/<id>` or subscribes via SSE `/policies/investigate/<id>/events` for status + final result. The job runs on a worker pool with bounded concurrency. Cancelling the job ID (DELETE) propagates ctx cancel to the in-flight work.
4. **The hang is generalized**, not investigate-specific: many simultaneous tab clicks fan out into a lot of expensive backend work. The fix is therefore at the *server* level, not just one handler. Concretely:
   - A bounded worker pool (semaphore) gates handlers that touch expensive paths (investigate + any source-querying endpoints).
   - The gin router itself stays unbounded for cheap reads (static, cached primitive listings) so tab-switch UX stays snappy.
   - Per-handler timeouts on the expensive class so a stuck IO call can't hold a worker slot forever.

## Refinements driven by the answers

These extend / clarify the acceptance criteria above:

- **(supersedes 8–11)** A `jobs` subsystem in `internal/framework/web/` exposes `Enqueue(ctx, fn) → id`, `Get(id) → status/result`, `Cancel(id)`, and `Subscribe(id) → <-chan event`. Backed by a worker pool sized by a code constant (default = `runtime.NumCPU()`, lower bound 2, upper bound 8). Jobs older than a retention window (e.g. 15 min) are GC'd.
- **(extends 9)** `handleInvestigator` becomes the async endpoint described in Q3. The existing GET path either redirects to the job-flow or returns a small HTMX fragment that subscribes to the job's SSE channel — UI behavior in the dashboard is preserved (user clicks Investigate, sees a result), but the HTTP handler returns in milliseconds.
- **(new)** A short list of "expensive" endpoints (investigate + any other source-querying / graph-walking ones surfaced during orient) is routed through the job subsystem. Cheap endpoints stay synchronous on gin's default goroutine-per-request model.
- **(new)** Under sustained click-storm load (e.g. 50 rapid clicks on the same expensive button), the server keeps responding to cheap requests with < 500ms latency, and the expensive queue stays bounded — no goroutine pile-up.
