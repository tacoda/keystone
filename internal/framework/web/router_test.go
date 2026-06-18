package web

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newTestRouter spins up a server against a temp project dir and
// returns the http.Handler the production server uses. Centralized
// so the gin migration is a one-line change here.
func newTestRouter(t *testing.T) (*server, http.Handler) {
	t.Helper()
	root := t.TempDir()
	srv, err := newServer(root)
	if err != nil {
		t.Fatalf("newServer: %v", err)
	}
	return srv, withTimeoutExceptSSE(srv.engine)
}

func TestRouter_HTMLPage(t *testing.T) {
	_, h := newTestRouter(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200", rr.Code)
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/html") {
		t.Errorf("content-type: got %q want text/html prefix", ct)
	}
}

// TestRouter_NewSectionURLs confirms every consolidated section URL
// renders 200. Old single-purpose URLs (e.g. /metrics, /primitives)
// are deliberately retired in 2.1.1 — those are checked separately
// for 404 by TestRouter_OldURLsRetired.
func TestRouter_NewSectionURLs(t *testing.T) {
	_, h := newTestRouter(t)
	urls := []string{
		"/",
		"/observability",
		"/observability/metrics",
		"/observability/insights",
		"/harness",
		"/harness/primitives",
		"/harness/primitives/new",
		"/harness/policies",
		"/harness/investigator",
		"/harness/graph",
		"/sources",
		"/sources/new",
		"/flywheels",
		"/flywheels/inbox",
		"/flywheels/prune",
		"/quality",
		"/quality/verify",
		"/quality/evals",
	}
	for _, u := range urls {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, u, nil)
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("GET %s: status %d, body=%s", u, rr.Code, firstChars(rr.Body.String(), 200))
		}
	}
}

// TestRouter_OldURLsRetired confirms the pre-2.1.1 single-purpose page
// URLs no longer resolve. Tools or links that still point at them
// must be migrated to the new section/tab URLs.
func TestRouter_OldURLsRetired(t *testing.T) {
	_, h := newTestRouter(t)
	urls := []string{
		"/metrics", "/insights",
		"/primitives", "/primitives/new", "/policies", "/policies/investigate", "/graph",
		"/verify", "/evals",
		"/prune", "/inbox",
	}
	for _, u := range urls {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, u, nil)
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Errorf("GET %s: status %d want 404 (URL was retired in 2.1.1)", u, rr.Code)
		}
	}
}

// TestRouter_KPIWidgets confirms each headline KPI endpoint responds
// 200 and returns an HTML fragment (no full document).
func TestRouter_KPIWidgets(t *testing.T) {
	_, h := newTestRouter(t)
	for _, name := range []string{"primitives", "sources", "inbox", "lint", "index"} {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/web/widgets/kpi/"+name, nil)
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("GET /web/widgets/kpi/%s: status %d, body=%q", name, rr.Code, firstChars(rr.Body.String(), 300))
			continue
		}
		body := rr.Body.String()
		if strings.Contains(body, "<!doctype") || strings.Contains(body, "<html") {
			t.Errorf("KPI %s: unexpected full document; got %q", name, firstChars(body, 80))
		}
		if !strings.Contains(body, "kpi-") {
			t.Errorf("KPI %s: response missing kpi-* class hook; got %q", name, firstChars(body, 200))
		}
	}
}

// TestRouter_KPIWidgetUnknown — 404 on an unknown KPI name keeps the
// surface tight; adding a KPI is a deliberate code change, not a
// path-shape guess.
func TestRouter_KPIWidgetUnknown(t *testing.T) {
	_, h := newTestRouter(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/web/widgets/kpi/nope", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("unknown KPI: status %d want 404", rr.Code)
	}
}

// TestRouter_GraphWidget confirms the lazy graph fragment endpoint
// returns a mermaid block and nothing else (no <html>).
func TestRouter_GraphWidget(t *testing.T) {
	_, h := newTestRouter(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/web/widgets/graph", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200", rr.Code)
	}
	body := rr.Body.String()
	if strings.Contains(body, "<!doctype") || strings.Contains(body, "<html") {
		t.Errorf("graph widget returned full document; got %q", firstChars(body, 80))
	}
	if !strings.Contains(body, "mermaid") {
		t.Errorf("graph widget missing mermaid block; got %q", firstChars(body, 200))
	}
}

// TestRouter_HarnessPrimitiveDetailURL exercises the kind/id prefix
// dispatch under the new harness section.
func TestRouter_HarnessPrimitiveDetailURL(t *testing.T) {
	_, h := newTestRouter(t)
	rr := httptest.NewRecorder()
	// No matching primitive — empty tempdir — but the route must
	// reach the handler (which returns 404 on miss), not the
	// catch-all NotFound. We accept any non-500 here; the contract
	// is that the URL pattern is recognized.
	req := httptest.NewRequest(http.MethodGet, "/harness/primitives/guide/some-id", nil)
	h.ServeHTTP(rr, req)
	if rr.Code >= 500 {
		t.Errorf("GET /harness/primitives/guide/some-id: status %d want < 500", rr.Code)
	}
}

func TestRouter_JSONAPI(t *testing.T) {
	_, h := newTestRouter(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/primitives", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200, body=%s", rr.Code, rr.Body.String())
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		t.Errorf("content-type: got %q want application/json prefix", ct)
	}
}

func TestRouter_404(t *testing.T) {
	_, h := newTestRouter(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/definitely-not-a-real-route", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status: got %d want 404", rr.Code)
	}
}

func TestRouter_StaticAssets(t *testing.T) {
	_, h := newTestRouter(t)
	// /assets/ is registered as a static FileServer. With nothing else
	// matching, a non-existent file under /assets/ must 404.
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/assets/nope.css", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("missing asset: got %d want 404", rr.Code)
	}
}

// TestRouter_HXRequestReturnsFragment confirms a page handler called
// with `HX-Request: true` returns just the page's `main` block, not
// the full layout. Necessary for the SPA shell — htmx swaps must not
// nest a complete document inside #app.
func TestRouter_HXRequestReturnsFragment(t *testing.T) {
	_, h := newTestRouter(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("HX-Request", "true")
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200", rr.Code)
	}
	body := rr.Body.String()
	if strings.Contains(body, "<!doctype") || strings.Contains(body, "<html") {
		t.Errorf("fragment response unexpectedly contains full document markers; body starts with %q", firstChars(body, 80))
	}
}

// TestRouter_NoHXReturnsFullLayout confirms that a normal browser GET
// (no HX-Request header) still server-renders the full layout — so
// deep links, reloads, and shared URLs continue to produce a complete
// document.
func TestRouter_NoHXReturnsFullLayout(t *testing.T) {
	_, h := newTestRouter(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "<!doctype") {
		t.Errorf("full response missing doctype; body starts with %q", firstChars(body, 80))
	}
	if !strings.Contains(body, `<main id="app"`) {
		t.Error("full response missing the SPA shell anchor <main id=\"app\">")
	}
}

func firstChars(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func TestRouter_SSEEndpoint(t *testing.T) {
	// /events is the long-lived SSE stream. It must NOT be wrapped in
	// the per-handler timeout. We start a real server, open the
	// connection, and confirm the server still responds with 200 + the
	// SSE content type. We close the connection ourselves shortly
	// after.
	_, h := newTestRouter(t)
	ts := httptest.NewServer(h)
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+"/events", nil)
	if err != nil {
		t.Fatalf("new req: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want 200", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "text/event-stream") {
		t.Errorf("content-type: got %q want text/event-stream", ct)
	}
	// Close on our terms — the stream is long-lived; we proved the
	// handler returned headers, which is what TimeoutHandler would have
	// suppressed.
	cancel()
	_, _ = io.Copy(io.Discard, resp.Body)
}
