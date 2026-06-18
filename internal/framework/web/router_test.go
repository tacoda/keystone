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
