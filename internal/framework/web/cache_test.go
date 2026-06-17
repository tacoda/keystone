package web

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPrimitiveCache_RefreshAndGet(t *testing.T) {
	root := t.TempDir()
	// Seed one guide so the walker has something to find.
	src := filepath.Join(root, ".keystone/harness/guides/process/spec.md")
	if err := os.MkdirAll(filepath.Dir(src), 0o755); err != nil {
		t.Fatal(err)
	}
	body := "---\nkind: guide\nid: process/spec\ndescription: x\n---\nbody\n"
	if err := os.WriteFile(src, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	c := newPrimitiveCache(root)
	// Empty before refresh.
	prims, _, err := c.get()
	if err != nil {
		t.Fatalf("get pre-refresh err: %v", err)
	}
	if len(prims) != 0 {
		t.Fatalf("expected empty pre-refresh, got %d", len(prims))
	}

	c.refresh()
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

func TestHealthCache_EmptyConfig(t *testing.T) {
	root := t.TempDir() // no context.json
	c := newHealthCache(root)
	c.refresh(context.Background())
	entries, err := c.get()
	if err != nil {
		t.Fatalf("get err: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries with no context.json, got %d", len(entries))
	}
}

func TestHealthCache_ProbesInParallel(t *testing.T) {
	// Three sources of unknown type → unknownAdapter.Health returns
	// instantly. The point of the test isn't latency vs serial — it's
	// that refresh completes without blocking under bounded concurrency
	// and the cache holds N entries.
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".keystone"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := `{
  "version": 1,
  "sources": [
    {"name": "a", "type": "unknown"},
    {"name": "b", "type": "unknown"},
    {"name": "c", "type": "unknown"}
  ]
}`
	if err := os.WriteFile(filepath.Join(root, ".keystone/context.json"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}

	c := newHealthCache(root)
	done := make(chan struct{})
	go func() {
		c.refresh(context.Background())
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("refresh did not return within 5s — parallel fan-out is blocked")
	}

	entries, err := c.get()
	if err != nil {
		t.Fatalf("get err: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	seen := map[string]bool{}
	for _, e := range entries {
		seen[e.Name] = true
	}
	for _, want := range []string{"a", "b", "c"} {
		if !seen[want] {
			t.Errorf("missing entry %q", want)
		}
	}
}
