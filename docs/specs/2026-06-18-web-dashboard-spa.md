---
kind: spec
id: 2026-06-18-web-dashboard-spa
status: draft
owner: tacoda
---

# Web dashboard — HTMX SPA, consolidated, observability-first

## Intent (restated)

Transform `internal/framework/web` from a multi-page-per-nav-link dashboard
into a **single-shell HTMX SPA** that:

- loads each section's data **on demand via the existing API/partial endpoints**,
- pushes **live updates over SSE** so any harness change (file write,
  primitive add/edit, sensor run, source health change) re-renders the
  affected pane immediately — no manual refresh,
- prettifies the styling and **consolidates** today's 14 top-nav links into
  a smaller set of sections-with-tabs that group by purpose,
- privileges **observability, auditability, on-the-fly harness updates,
  metrics, and insights** as first-class.

The whole point of this dashboard: see what the harness is doing, prove
what it did, change it on the fly.

## Acceptance criteria

A reviewer can verify each of these objectively.

### A. SPA shell

- A1. `layout.html` becomes a **persistent shell**: topbar + side/tab nav +
  one `<main id="app">` swap target + footer. Header/footer/nav never
  re-render on in-app navigation.
- A2. Every top-level nav link is `hx-get=<section-url> hx-target="#app"
  hx-swap="innerHTML show:window:top" hx-push-url="true"`.
- A3. Browser back/forward restores the previous section via
  `hx-push-url` history — verified by a router_test that asserts a section
  endpoint returns *just the section fragment* when called with
  `HX-Request: true`, and the full `layout` when called without.
- A4. Initial cold-load on any deep URL (`/metrics`, `/primitives?...`)
  still server-renders the full layout, so links shared and reloads work.

### B. On-demand data via API

- B1. Section fragments do **not** prefetch sub-widget data inline. Each
  widget within a section uses `hx-trigger="load"` (or `revealed`,
  `intersect once` for below-fold) to pull its own data from
  `/api/...` or `/_partials/...` endpoints.
- B2. List/filter endpoints stay HTML-fragment endpoints; new JSON
  endpoints are added only when a widget genuinely needs JSON (charts).
- B3. No widget makes more than one request per refresh cycle. Widget
  endpoints are individually cacheable.

### C. Live updates via SSE

- C1. The existing `/events` hub stays the single SSE channel.
- C2. Each live widget subscribes via `sse-swap="<event-name>"` to a
  named topic; widgets unrelated to that topic do not re-render.
- C3. Topics added (at minimum): `harness-changed`, `primitives-changed`,
  `sources-changed`, `health-changed`, `inbox-changed`, `verify-ran`,
  `eval-ran`. The watcher classifies its diff into the narrowest topic
  it can (fall back to `harness-changed` for cross-cutting changes).
- C4. On topic event, the SSE body is either (a) the replacement HTML
  fragment for the widget, or (b) a tiny "stale" pill that triggers a
  fresh `hx-get` for the widget. Pick (a) for cheap renders, (b) for
  expensive ones. Either choice MUST be reflected in the widget's
  HTMX attributes.

### D. On-the-fly harness updates

- D1. Forms that mutate the harness (new primitive, edit primitive,
  add source, run verify, accept inbox candidate, prune item) submit
  via `hx-post` to existing handlers and on success either swap their
  own panel or trigger an SSE event that swaps the affected panels —
  **no full reload anywhere**.
- D2. After a mutate succeeds, the watcher's debounced diff publishes
  the narrowest SSE topic; every subscriber's swap completes within
  one debounce window (acceptable: ≤500ms median in dev).
- D3. Mutations are auditable: a thin **audit log widget** (last N
  harness change events with timestamp + summary) lives in the
  Observability section. Backed by the watcher event stream; persisted
  **per session** as append-only JSONL under
  `.keystone/state/audit/session-<UTC-ISO8601>-<short-id>.jsonl`.
  - File created once at server start; never overwritten.
  - Server process owns one file for its lifetime.
  - Default widget tails the current session's file; a "history"
    selector lists prior sessions (read-only) sorted newest-first.
  - Pruning: a startup task deletes session files older than 30 days
    OR keeps at most 50 sessions, whichever is looser. Explicit, not
    silent — one log line per prune.
  - Dir is git-ignored.

### E. Consolidation — sections × tabs

Today: 14 nav links. After: **5 sections**, each with internal tabs.

| Section          | Tabs / widgets                                                   |
|------------------|------------------------------------------------------------------|
| Harness          | primitives · policies · investigator · graph                     |
| Sources          | list · per-source detail · new                                   |
| Flywheels        | inbox · prune · flywheels overview                               |
| Quality          | verify · evals                                                   |
| Observability    | metrics · insights · audit log · live event tail                 |

- E1. The home page becomes the Observability default tab (KPIs +
  kind breakdown + audit log + live tail). The "home" route stays as
  an alias to `/observability`.
- E2. Each section's URL pattern: `/<section>` and
  `/<section>/<tab>`. Tab-switch is HTMX swap of the tab panel only,
  with `hx-push-url`.
- E3. Search lives in the topbar as a **command-K popover**. The
  existing `/search` endpoint returns a result-list fragment that
  renders into the popover (`hx-get` debounced on keystrokes). Old
  `/search` URL kept as a render fallback for bookmarks.
- E4. Old page URLs are **retired**. No redirects, no aliases.
  Layout nav + every internal link points at the new section/tab
  URLs. Old links 404. (REST `/api/...` and `/web/actions/...`
  stay; only the page URLs move.)

### F. Styling pass

- F1. `keystone.css` keeps one file (no framework added). Updates:
  CSS variables for color tokens; consistent card / pill / muted /
  ok / bad components; better typography scale; spacing scale on a
  4px grid; sticky topbar; section-tab styling; subtle live-update
  flash animation on SSE-driven swaps.
- F2. **Dark-by-default**, slightly lighter than today's near-black
  — a softer dark surface (`#11151c` family) with **blue + gray
  accents**. Audience = software engineers using agentic coding,
  so the visual language matches modern dev tools (subdued
  background, high-contrast text, blue as primary accent, gray
  for chrome/dividers, no neon). No light mode in this pass.
- F3. **Snappy** is a hard requirement (devs are harsh judges).
  Targets:
  - First-paint < 100ms on localhost (server-rendered HTML, no
    client-side hydration cost).
  - Section/tab swap < 50ms perceived — achieved by HTMX
    `boost`-style fragment swaps + cache-warm endpoints + zero
    layout shift on swap (reserve panel min-height).
  - Static assets served with `Cache-Control: public, max-age=
    31536000, immutable` (assets are embedded and version-pinned
    via the binary).
  - Preconnect / preload the embedded JS in the layout `<head>`.
  - Use the View Transitions API where available for free fade
    on swap (no fallback overhead; progressive enhancement).
  - No spinners on swaps under 200ms — use `htmx-request` opacity
    nudge only.
- F4. No JS dependencies added beyond what's already vendored
  (`htmx.min.js`, `htmx-ext-sse.js`).

### G. Metrics & insights uplift

- G1. The Observability section's metrics tab gets a small set of
  **headline KPIs** at the top (primitive count, sources OK / total,
  inbox depth, last verify status, last eval pass rate). Each KPI is
  its own SSE-aware widget.
- G2. Insights tab keeps its current report but renders as collapsible
  cards rather than one long page.
- G3. Adding a new KPI is a single Go handler + a single CSS-only
  card — documented in code via a one-line comment, not a new doc.

### H. Tests

- H1. `router_test.go` adds cases that confirm:
  - section endpoints return a *fragment* under `HX-Request: true`,
  - section endpoints return the *full layout* otherwise,
  - tab endpoints return the *tab panel only* under `HX-Request: true`,
  - deep-link `/primitives?kind=guide` resolves to harness section,
    primitives tab, with the filter applied.
- H2. `sse_test` (new or extending existing) confirms the watcher
  publishes the narrowest topic for representative diffs (touch a
  guide → `primitives-changed`; touch `context.json` →
  `sources-changed`; touch random doc → `harness-changed`).
- H3. No regressions in `cache_test.go` / `cache_layered_test.go`.

## Non-goals

- Not introducing React / Vue / any SPA framework. HTMX + Go templates only.
- Not changing the data model in `internal/framework/primitive`.
- Not adding persistent storage for the audit log (in-memory ring buffer is fine).
- Not redesigning the brand identity / logo.
- Not adding a dark-mode toggle UI (auto via media query only).
- Not building a chart library; KPIs render as numbers + tiny
  sparkline-as-inline-SVG only if the data is already on hand.
- Not exposing the dashboard beyond localhost. Auth / RBAC stays out.
- Not migrating off Gin.

## Resolved decisions (2026-06-18)

1. **Audit log** — **per-session** append-only JSONL under
   `.keystone/state/audit/session-<UTC-ISO>-<id>.jsonl`. New file per
   `keystone web serve` process; never overwritten. Startup prune
   trims to ≤50 sessions or ≤30 days, whichever is looser. Dashboard
   defaults to current session, history selector for prior.
2. **Search** — command-K popover; `/search` URL kept as fallback.
3. **KPIs** — five proposed are accepted (primitive count, sources OK/total,
   inbox depth, last verify status, last eval pass rate).
4. **Graph tab** — lazy render (`intersect once` when tab activates).
5. **Section names** — accepted (Harness, Sources, Flywheels, Quality,
   Observability).

## Release tail

After verify + review pass: changelog entry under `[2.1.1]`, tag `v2.1.1`,
push tag — CI / goreleaser handles release. No manual `gh release`.

## Out-of-band

- Touched paths: `internal/framework/web/**`, `cmd/keystone/web.go`
  (only if a new sub-flag is needed; default no).
- Iron laws still apply: sensors must run before "done"; no `--no-verify`;
  no AI attribution.
