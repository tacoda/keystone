package web

import (
	"net/http"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

// handleGraph renders the graph page shell. The mermaid diagram
// itself is loaded lazily — see handleGraphWidget — so opening this
// page is cheap even for large charters; the heavy render only
// fires when the user actually scrolls the canvas into view.
func (s *server) handleGraph(w http.ResponseWriter, r *http.Request) {
	s.renderPage(w, r, "graph.html", map[string]any{
		"ProjectDir": s.projectDir,
	})
}

// handleGraphWidget returns the mermaid fragment used by the lazy
// loader in graph.html. Expensive — full primitive walk + render —
// so it's gated behind `hx-trigger="intersect once"` on the page.
func (s *server) handleGraphWidget(w http.ResponseWriter, r *http.Request) {
	primitives, err := s.loadPrimitives()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	mermaid := primitive.RenderGraph(primitives, primitive.GraphMermaid)
	s.render(w, "_graph_widget.html", map[string]any{
		"Mermaid": mermaid,
	})
}
