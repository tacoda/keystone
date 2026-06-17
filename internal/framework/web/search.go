package web

import (
	"net/http"
	"strconv"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

func (s *server) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	primitives, err := s.loadPrimitives()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	limit := 50
	if n, _ := strconv.Atoi(r.URL.Query().Get("limit")); n > 0 {
		limit = n
	}
	hits := primitive.Search(s.projectDir, primitives, q, limit)
	s.render(w, "search.html", map[string]any{
		"ProjectDir": s.projectDir,
		"Query":      q,
		"Hits":       hits,
	})
}

func (s *server) handleSearchFragment(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	primitives, err := s.loadPrimitives()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	hits := primitive.Search(s.projectDir, primitives, q, 50)
	s.render(w, "_search_results.html", map[string]any{
		"Query": q,
		"Hits":  hits,
	})
}

func (s *server) apiSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	primitives, err := s.loadPrimitives()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	limit := 100
	if n, _ := strconv.Atoi(r.URL.Query().Get("limit")); n > 0 {
		limit = n
	}
	hits := primitive.Search(s.projectDir, primitives, q, limit)
	writeJSON(w, http.StatusOK, map[string]any{
		"query": q,
		"count": len(hits),
		"hits":  hits,
	})
}
