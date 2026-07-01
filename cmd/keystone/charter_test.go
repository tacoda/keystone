package main

import (
	"testing"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

func TestFilterByKind(t *testing.T) {
	prims := []primitive.Primitive{
		{Frontmatter: primitive.Frontmatter{Kind: "guide", ID: "x"}},
		{Frontmatter: primitive.Frontmatter{Kind: "sensor", ID: "y"}},
	}
	if got := filterByKind(prims, ""); len(got) != 2 {
		t.Errorf("empty kind should pass all, got %d", len(got))
	}
	got := filterByKind(prims, "sensor")
	if len(got) != 1 || got[0].Kind != "sensor" {
		t.Errorf("--kind sensor should filter to sensors, got %+v", got)
	}
}

func TestParseShowOpts(t *testing.T) {
	o := parseShowOpts([]string{"--effective", "--kind", "sensor", "--dir", "/tmp/x"})
	if !o.effective {
		t.Error("--effective not parsed")
	}
	if o.kind != "sensor" {
		t.Errorf("--kind = %q, want sensor", o.kind)
	}
	if o.dir != "/tmp/x" {
		t.Errorf("--dir = %q, want /tmp/x", o.dir)
	}
}
