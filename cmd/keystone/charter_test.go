package main

import (
	"testing"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

func prim(kind, id, prov string) primitive.Primitive {
	return primitive.Primitive{
		Frontmatter: primitive.Frontmatter{Kind: kind, ID: id, Description: "d"},
		Provenance:  prov,
	}
}

func TestRosterEntries_EffectiveProjectWins(t *testing.T) {
	prims := []primitive.Primitive{
		prim("guide", "idioms/go/x", "policy/acme"),
		prim("guide", "idioms/go/x", "project"), // project overrides the policy
		prim("guide", "idioms/go/y", "policy/acme"),
	}
	entries := rosterEntries(prims, showOpts{effective: true})
	if len(entries) != 2 {
		t.Fatalf("expected 2 effective entries (deduped), got %d", len(entries))
	}
	// Find the overridden one.
	var x *rosterEntry
	for i := range entries {
		if entries[i].p.ID == "idioms/go/x" {
			x = &entries[i]
		}
	}
	if x == nil {
		t.Fatal("missing idioms/go/x")
	}
	if x.p.Provenance != "project" {
		t.Errorf("project should win, got provenance %q", x.p.Provenance)
	}
	if len(x.shadows) != 1 || x.shadows[0] != "policy/acme" {
		t.Errorf("expected policy/acme shadowed, got %v", x.shadows)
	}
}

func TestRosterEntries_NonEffectiveListsAll(t *testing.T) {
	prims := []primitive.Primitive{
		prim("guide", "x", "policy/acme"),
		prim("guide", "x", "project"),
	}
	entries := rosterEntries(prims, showOpts{}) // no --effective
	if len(entries) != 2 {
		t.Errorf("non-effective should list every layer, got %d", len(entries))
	}
}

func TestRosterEntries_KindFilter(t *testing.T) {
	prims := []primitive.Primitive{
		prim("guide", "x", "project"),
		prim("sensor", "y", "project"),
	}
	entries := rosterEntries(prims, showOpts{kind: "sensor"})
	if len(entries) != 1 || entries[0].p.Kind != "sensor" {
		t.Errorf("--kind sensor should filter to sensors only, got %+v", entries)
	}
}
