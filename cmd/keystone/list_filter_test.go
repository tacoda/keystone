package main

import (
	"testing"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

func TestList_KindFilterAndOrder(t *testing.T) {
	ps := []primitive.Primitive{{Frontmatter: primitive.Frontmatter{Kind: "skill", ID: "a"}},
		{Frontmatter: primitive.Frontmatter{Kind: "hook", ID: "b"}}, {Frontmatter: primitive.Frontmatter{Kind: "rule", ID: "c"}}}
	primitive.SortByKind(ps)
	if got := ps[0].Kind + "," + ps[1].Kind + "," + ps[2].Kind; got != "rule,hook,skill" {
		t.Errorf("order = %s, want rule,hook,skill in KnownKinds order", got)
	}
	got := filterPrimitives(ps, "hook", nil)
	if len(got) != 1 || got[0].Kind != "hook" {
		t.Errorf("filter hook = %v, want the one hook", got)
	}
	if got := filterPrimitives(ps, "guide", nil); len(got) != 1 || got[0].Kind != "rule" {
		t.Errorf("filter guide (4.x) = %v, want the rule", got)
	}
}
