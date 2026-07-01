package web

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPrimitiveCache_RefreshAndGet(t *testing.T) {
	root := t.TempDir()
	// Seed one guide so the walker has something to find.
	src := filepath.Join(root, ".charter/guides/process/spec.md")
	if err := os.MkdirAll(filepath.Dir(src), 0o755); err != nil {
		t.Fatal(err)
	}
	body := "---\nkind: guide\nid: process/spec\ndescription: x\n---\nbody\n"
	if err := os.WriteFile(src, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	c := newPrimitiveCache(root)
	// Cold get walks on demand — first call must return populated
	// data, not an empty slice. (Previously the cache required an
	// explicit refresh; the layered cache fills on access.)
	prims, idx, err := c.get()
	if err != nil {
		t.Fatalf("get post-refresh err: %v", err)
	}
	if len(prims) != 1 {
		t.Fatalf("expected 1 primitive post-refresh, got %d", len(prims))
	}
	if prims[0].Kind != "guide" || prims[0].ID != "process/spec" {
		t.Errorf("wrong primitive: %+v", prims[0])
	}
	if len(idx.Primitives) != 1 {
		t.Errorf("expected index to also have 1 primitive, got %d", len(idx.Primitives))
	}
}
