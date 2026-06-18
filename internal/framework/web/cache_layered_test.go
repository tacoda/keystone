package web

import (
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/policies"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// seedGuide writes a minimal primitive doc under <root>/<harnessRoot>/guides/process/<name>.md.
func seedGuide(t *testing.T, root, harnessRoot, idLeaf string) {
	t.Helper()
	dir := filepath.Join(root, harnessRoot, "guides", "process")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "---\nkind: guide\nid: process/" + idLeaf + "\ndescription: x\n---\nbody\n"
	if err := os.WriteFile(filepath.Join(dir, idLeaf+".md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLayeredCache_ProjectOnly(t *testing.T) {
	root := t.TempDir()
	seedGuide(t, root, config.DefaultHarnessRoot, "spec")

	c := newPrimitiveCache(root)
	c.refreshLayer(config.DefaultHarnessRoot)

	prims, _, err := c.getLayer(config.DefaultHarnessRoot)
	if err != nil {
		t.Fatalf("getLayer err: %v", err)
	}
	if len(prims) != 1 || prims[0].ID != "process/spec" {
		t.Fatalf("layer contents wrong: %+v", prims)
	}
}

func TestLayeredCache_MultipleLayers(t *testing.T) {
	root := t.TempDir()
	seedGuide(t, root, config.DefaultHarnessRoot, "spec")

	policyA := filepath.Join(config.DefaultHarnessRoot, policies.PolicyRoot, "alpha")
	policyB := filepath.Join(config.DefaultHarnessRoot, policies.PolicyRoot, "beta")
	seedGuide(t, root, policyA, "rule-a")
	seedGuide(t, root, policyB, "rule-b")

	c := newPrimitiveCache(root)
	for _, h := range []string{config.DefaultHarnessRoot, policyA, policyB} {
		c.refreshLayer(h)
	}

	cases := map[string]string{
		config.DefaultHarnessRoot: "process/spec",
		policyA:                   "process/rule-a",
		policyB:                   "process/rule-b",
	}
	for h, wantID := range cases {
		prims, _, err := c.getLayer(h)
		if err != nil {
			t.Errorf("layer %s err: %v", h, err)
			continue
		}
		if len(prims) != 1 || prims[0].ID != wantID {
			t.Errorf("layer %s: got %+v want id %s", h, prims, wantID)
		}
	}
}

func TestLayeredCache_ColdMissSyncs(t *testing.T) {
	root := t.TempDir()
	seedGuide(t, root, config.DefaultHarnessRoot, "spec")
	c := newPrimitiveCache(root)

	// No refresh first — getLayer must walk on demand.
	prims, _, err := c.getLayer(config.DefaultHarnessRoot)
	if err != nil {
		t.Fatalf("cold-miss err: %v", err)
	}
	if len(prims) != 1 {
		t.Fatalf("cold-miss missing data: got %d", len(prims))
	}
}

func TestLayeredCache_ConcurrentReads(t *testing.T) {
	root := t.TempDir()
	seedGuide(t, root, config.DefaultHarnessRoot, "spec")
	c := newPrimitiveCache(root)
	c.refreshLayer(config.DefaultHarnessRoot)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, _ = c.getLayer(config.DefaultHarnessRoot)
		}()
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("concurrent reads blocked")
	}
}

// Walk-counter seam: production code calls walkFn so tests can
// observe how often a layer is walked.
func TestCollectInventory_UsesCache(t *testing.T) {
	root := t.TempDir()
	seedGuide(t, root, config.DefaultHarnessRoot, "spec")

	srv, err := newServer(root)
	if err != nil {
		t.Fatalf("newServer: %v", err)
	}
	srv.primitiveCache.refreshLayer(config.DefaultHarnessRoot)

	var calls int64
	prev := walkFn
	walkFn = func(projectDir, harnessRoot string) ([]primitive.Primitive, []primitive.Warning, error) {
		atomic.AddInt64(&calls, 1)
		return prev(projectDir, harnessRoot)
	}
	t.Cleanup(func() { walkFn = prev })

	if _, err := srv.collectInventory(t.Context(), ""); err != nil {
		t.Fatalf("collectInventory: %v", err)
	}
	if got := atomic.LoadInt64(&calls); got != 0 {
		t.Errorf("expected 0 Walk calls (warm cache), got %d", got)
	}
}
