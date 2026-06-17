package web

import (
	"fmt"
	"net/http"
	"sync"
)

// sseHub is a small fan-out: handlers subscribe over HTTP via
// /events, the fsWatcher publishes events, the hub forwards each
// event to every subscriber. Goroutine-per-subscriber; no buffering
// beyond the channel.
type sseHub struct {
	mu          sync.Mutex
	subscribers map[chan sseEvent]struct{}
}

type sseEvent struct {
	// Name is the SSE `event:` line (e.g. "primitive-changed").
	Name string
	// Data is the SSE `data:` payload. For htmx-ext-sse, this is
	// typically an HTML fragment that swap-targets a dashboard
	// element.
	Data string
}

func newSSEHub() *sseHub {
	return &sseHub{subscribers: map[chan sseEvent]struct{}{}}
}

// ServeHTTP implements the /events endpoint. Sets the SSE headers,
// flushes immediately, then loops forwarding events from the
// subscriber channel until the client disconnects.
func (h *sseHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	ch := make(chan sseEvent, 8)
	h.subscribe(ch)
	defer h.unsubscribe(ch)

	// Initial comment so the client transport knows the stream is
	// open. Some browsers wait for the first byte before firing the
	// `open` event.
	fmt.Fprintf(w, ":connected\n\n")
	flusher.Flush()

	notify := r.Context().Done()
	for {
		select {
		case <-notify:
			return
		case ev := <-ch:
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.Name, sseEscape(ev.Data))
			flusher.Flush()
		}
	}
}

func (h *sseHub) subscribe(ch chan sseEvent) {
	h.mu.Lock()
	h.subscribers[ch] = struct{}{}
	h.mu.Unlock()
}

func (h *sseHub) unsubscribe(ch chan sseEvent) {
	h.mu.Lock()
	delete(h.subscribers, ch)
	h.mu.Unlock()
}

// Publish fans an event out to every current subscriber.
// Non-blocking — if a subscriber's channel is full, the event is
// dropped for that subscriber. Better to skip a beat than back the
// hub up.
func (h *sseHub) Publish(ev sseEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.subscribers {
		select {
		case ch <- ev:
		default:
		}
	}
}

// sseEscape collapses multi-line data into a single `data:` line by
// joining with `\ndata: `. Required by the SSE spec — every newline
// inside the payload becomes a continuation prefix.
func sseEscape(s string) string {
	out := ""
	for i, r := range s {
		if r == '\n' {
			out += "\ndata: "
			continue
		}
		_ = i
		out += string(r)
	}
	return out
}
