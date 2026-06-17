package web

import (
	"net/http"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

func (s *server) handleGraph(w http.ResponseWriter, r *http.Request) {
	primitives, err := s.loadPrimitives()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	mermaid := primitive.RenderGraph(primitives, primitive.GraphMermaid)
	s.render(w, "graph.html", map[string]any{
		"ProjectDir": s.projectDir,
		"Mermaid":    mermaid,
	})
}
