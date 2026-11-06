package main

import (
	"testing"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

func TestList_CategoryFilterAndOrder(t *testing.T) {
	for _, c := range []struct {
		kind, filter string
		want         bool
	}{
		{"hook", "constraints", true}, {"hook", "capability", false},
		{"skill", "capabilities", true}, {"rule", "guide", true}, {"rule", "skill", false},
	} {
		if got := matchesKindOrCategory(c.kind, c.filter); got != c.want {
			t.Errorf("matches(%s, %s) = %v", c.kind, c.filter, got)
		}
	}
	ps := []primitive.Primitive{{Frontmatter: primitive.Frontmatter{Kind: "skill", ID: "a"}},
		{Frontmatter: primitive.Frontmatter{Kind: "hook", ID: "b"}}, {Frontmatter: primitive.Frontmatter{Kind: "rule", ID: "c"}}}
	primitive.SortByCategory(ps)
	if got := ps[0].Kind + "," + ps[1].Kind + "," + ps[2].Kind; got != "rule,hook,skill" {
		t.Errorf("order = %s, want constraints (in kind order) before capabilities", got)
	}
}
