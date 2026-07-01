package web

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/policies"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// PolicyInventory is the per-request snapshot the investigator panel
// renders. Each layer (project + each declared policy) carries the
// primitives it ships; strict items are flagged per policy.
type PolicyInventory struct {
	ProjectDir string       `json:"project_dir"`
	Layers     []LayerEntry `json:"layers"`
	Search     string       `json:"search,omitempty"`
}

// LayerEntry is one cascade layer — either the project or a named
// policy. Order in the inventory mirrors the cascade order: project
// first (wins by default), then policies in declaration order
// (deeper-nested wins between policies).
type LayerEntry struct {
	Name       string                `json:"name"`
	Kind       string                `json:"kind"` // "project" | "policy"
	Source     string                `json:"source,omitempty"`
	Version    string                `json:"version,omitempty"`
	Strict     map[string][]string   `json:"strict,omitempty"`
	Primitives []primitive.Primitive `json:"primitives"`
}

// collectInventory returns the primitives every cascade layer ships
// — project + each declared policy — from the cache. Cold layers
// walk on first access; subsequent requests are pointer reads.
// Honors the optional `search` substring filter — applied
// case-insensitively against id + description + path.
//
// ctx is reserved for future cancellation hooks; current
// implementation is non-blocking on warm cache and bounded by a
// single-layer walk on cold miss.
func (s *server) collectInventory(_ context.Context, search string) (*PolicyInventory, error) {
	inv := &PolicyInventory{ProjectDir: s.projectDir, Search: search}
	needle := strings.ToLower(strings.TrimSpace(search))

	// Project layer.
	projectPrimitives, _, err := s.primitiveCache.getLayer(config.DefaultCharterRoot)
	if err != nil {
		return nil, err
	}
	inv.Layers = append(inv.Layers, LayerEntry{
		Name:       "project",
		Kind:       "project",
		Primitives: filterPrimitivesBySearch(projectPrimitives, needle),
	})

	// Each declared policy.
	cfg, _ := config.ReadProjectConfig(s.projectDir)
	if cfg == nil {
		return inv, nil
	}
	for _, p := range flattenPolicies(cfg.Policies) {
		policyRoot := filepath.Join(config.DefaultCharterRoot, policies.PolicyRoot, p.Name)
		// getLayer returns nil + err on a missing vendored tree; we
		// surface that as zero primitives, mirroring the previous
		// per-call Walk behavior.
		policyPrimitives, _, _ := s.primitiveCache.getLayer(policyRoot)
		inv.Layers = append(inv.Layers, LayerEntry{
			Name:       p.Name,
			Kind:       "policy",
			Source:     p.Source,
			Version:    p.Version,
			Strict:     p.Strict,
			Primitives: filterPrimitivesBySearch(policyPrimitives, needle),
		})
	}
	return inv, nil
}

// flattenPolicies walks the nested policy tree pre-order and returns
// every node, flattened. Matches the install / loader convention.
func flattenPolicies(nodes []config.PolicyNode) []config.PolicyNode {
	var out []config.PolicyNode
	var walk func(ns []config.PolicyNode)
	walk = func(ns []config.PolicyNode) {
		for _, n := range ns {
			out = append(out, n)
			if len(n.Children) > 0 {
				walk(n.Children)
			}
		}
	}
	walk(nodes)
	return out
}

func filterPrimitivesBySearch(in []primitive.Primitive, needle string) []primitive.Primitive {
	if needle == "" {
		// Sort by kind+id for stable display.
		out := append([]primitive.Primitive(nil), in...)
		sort.Slice(out, func(i, j int) bool {
			if out[i].Kind != out[j].Kind {
				return out[i].Kind < out[j].Kind
			}
			return out[i].ID < out[j].ID
		})
		return out
	}
	out := in[:0:0]
	for _, p := range in {
		if strings.Contains(strings.ToLower(p.ID), needle) ||
			strings.Contains(strings.ToLower(p.Description), needle) ||
			strings.Contains(strings.ToLower(p.Path), needle) {
			out = append(out, p)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Kind != out[j].Kind {
			return out[i].Kind < out[j].Kind
		}
		return out[i].ID < out[j].ID
	})
	return out
}

// handleInvestigator renders the full policy-investigator page.
func (s *server) handleInvestigator(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("q")
	inv, err := s.collectInventory(r.Context(), search)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.renderPage(w, r, "investigator.html", map[string]any{
		"ProjectDir": s.projectDir,
		"Inventory":  inv,
	})
}

// handleInvestigatorFragment returns just the layers section for
// HTMX search-input updates.
func (s *server) handleInvestigatorFragment(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("q")
	inv, err := s.collectInventory(r.Context(), search)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "_investigator_layers.html", map[string]any{
		"Inventory": inv,
	})
}

// statSize returns the file size on disk for a primitive. Used by
// the investigator template to flag empty / oversized primitives.
func statSize(projectDir, relPath string) int64 {
	info, err := os.Stat(filepath.Join(projectDir, relPath))
	if err != nil {
		return -1
	}
	return info.Size()
}
