# Single-Page Application (SPA)

A web application that loads **once**, then updates the page in place via JavaScript — fetching data and re-rendering portions of the DOM rather than reloading whole pages from the server. Articulated in Steve Yegge's "Stevey's Tech Talk" (2007) and operationalized by frameworks beginning with Backbone (2010), Angular (2010), React (2013), Vue (2014), and Svelte (2016). The dominant frontend architecture for the past decade, though now in active rebalancing as server-rendered approaches return.

The SPA's promise: a *desktop-class* interactive experience inside the browser, with state preserved across navigations and a single bundle that handles the whole UI. The cost: the browser becomes a heavyweight runtime, the network becomes a critical path, and the back button — a feature you got for free in 1996 — is now your responsibility to implement.

> **Rules extracted:** [`guides/principles/spa.md`](../../guides/principles/spa.md). This file holds the full reasoning, anti-patterns, and references.

## Defining properties

- **Client-side rendering** (CSR) is the default. The first response is a near-empty HTML shell; the JavaScript bundle builds the page on arrival.
- **Routing in the browser.** URLs change without server round-trips; the framework's router intercepts navigation and renders the matching component.
- **API-backed.** The backend exposes JSON/GraphQL endpoints; the SPA composes views from API data.
- **State on the client.** Form state, navigation state, and often domain state live in client memory (Redux, Vuex, Zustand, MobX). Server is a sync target, not the source of truth between user actions.
- **Build pipeline.** The SPA is *compiled*: bundling, tree-shaking, code-splitting, source maps. Build complexity is part of the architecture, not optional infrastructure.

## What is no longer free

Things the multi-page web gave you, that SPAs must reimplement:

- **The back button.** History API integration; care with redirects, replace-vs-push semantics, scroll restoration.
- **Initial render time.** No HTML until the bundle parses, downloads its data, and renders. Latency and CPU on the client are now first-class concerns. See [[premature-optimization]] (measure, then optimize) and [[observability]] (Core Web Vitals; Real User Monitoring).
- **SEO.** Search crawlers historically struggled with client-rendered content; some still do. Server-side rendering (SSR), static generation (SSG), or pre-rendering recovers it — at the cost of more architecture.
- **Accessibility.** Native page navigation announces itself to screen readers; SPA navigation does not, unless you implement it. ARIA live regions, focus management, route announcements.
- **State preservation across crashes.** The user typed half a form; the tab crashed; the state is gone. Browsers don't restore client state automatically.
- **Caching.** The browser's HTTP cache no longer caches "pages"; you cache API responses, route chunks, and whole bundles, each with its own invalidation story. See [[hyrums-law]] (every cache key becomes a contract).

The SPA pattern *can* deliver excellent UX; it does so reliably only when these costs are budgeted from day one.

## What it asks of you

- When you choose SPA, choose it for a **use case** that benefits — rich interactivity, application-shaped flows (editors, dashboards, IDEs). Marketing pages, blogs, and documentation are bad SPA fits. See [[least-astonishment]] — the back button working correctly is non-negotiable.
- When you fetch data, design for the **loading state, the error state, the empty state, and the success state** of every screen. Server-rendered pages mostly ignored these; SPAs cannot. See [[error-handling]].
- When you manage client state, the simplest store that works is the best store. Reach for Redux / global state only when local state genuinely cannot do the job. See [[simple-design]] and [[simplicity]].
- When you ship code, ship **less of it.** Code-split by route; lazy-load expensive components; analyze bundle size as a CI metric. Every kilobyte is paid by every user.
- When you secure the SPA, remember that **the client is hostile.** Anything in the bundle is visible; anything in client state is reachable from the console. Authorization is enforced by the server, not by hiding the button. See [[security]] and [[security-threats]] (broken access control).
- When you render dynamic content from user input, encode it correctly for HTML, attributes, and JS contexts. The SPA framework helps but does not eliminate XSS. See [[security-threats]] (A03 — Injection).

## Anti-patterns

- A "single-page application" that loads a 4MB JavaScript bundle to render a static marketing page. The user is paying for an SPA but getting nothing in return.
- Authorization implemented by "hiding the admin button" in the UI. The API endpoint accepts the request anyway. See [[security-threats]] (broken access control).
- State management as the application's most complex subsystem. Every action goes through five layers (action → middleware → reducer → selector → component) before anything happens; the state library has become the architecture.
- "It works on my machine" — performance acceptance on a fast network and a fast device, while real users on midrange phones over 4G are timing out.
- A back-button click that loses unsaved work, with no warning, no preservation, and no autosave.
- Mixing many state-management libraries in one app because each one was the right choice the week it was added.
- Client-side routing that doesn't update the page `<title>` and ARIA-announces the navigation. Accessibility regressed silently.

## References

- Yegge, S. (2007). *Tech Talk: Steve Yegge on Single-Page Web Applications.* (Early articulation of the term.)
- Resig, J. (2010). *Modern JavaScript Web Development*. (The Backbone era articulation.)
- Osmani, A. (2019). *Learning Patterns for the Web Developer.* O'Reilly + patterns.dev. (Modern patterns and trade-offs; RUM, code-splitting, hydration.)
- Russell, A. (2017). "Can You Afford It?: Real-world Web Performance Budgets." infrequently.org. (The most cited modern critique of bundle weight.)
- Google. *Web Vitals* (2020+). web.dev/vitals. (The de facto user-experience metrics any SPA must measure against.)
- Walke, J. (2013). *React: A JavaScript library for building user interfaces.* (The library whose component model defined the modern SPA shape.)
