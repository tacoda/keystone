package web

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/policies"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// walkFn is the indirect entry to primitive.Walk so tests can
// observe and count layer refreshes. Production code must not
// reassign it outside tests.
var walkFn = primitive.Walk

// layerSnapshot is one cache entry — the most recent walk of a
// single charter root.
type layerSnapshot struct {
	primitives []primitive.Primitive
	index      primitive.Index
	err        error
}

// primitiveCache holds the most recent walk of every accessed
// charter layer — project root plus every declared policy root.
// Page handlers read from here instead of re-walking on every
// request. The fsWatcher refreshes dirty layers on debounced file
// changes; other layers stay warm.
type primitiveCache struct {
	projectDir string

	mu     sync.RWMutex
	layers map[string]*layerSnapshot // key: charterRoot (filepath-relative)

	// coldMu serializes cold-miss walks per charterRoot key, so two
	// goroutines that hit getLayer simultaneously do not both walk
	// the same tree. Warm reads never touch coldMu.
	coldMu sync.Map
}

func newPrimitiveCache(projectDir string) *primitiveCache {
	return &primitiveCache{
		projectDir: projectDir,
		layers:     map[string]*layerSnapshot{},
	}
}

// refreshLayer walks a single charter root and replaces its cached
// snapshot. Safe to call from any goroutine.
func (c *primitiveCache) refreshLayer(charterRoot string) {
	prims, _, err := walkFn(c.projectDir, charterRoot)
	snap := &layerSnapshot{primitives: prims}
	if err != nil {
		// Vendored policy trees may not be installed — surface as an
		// empty layer, not a hard error, mirroring the previous
		// collectInventory behavior.
		snap.err = err
		snap.primitives = nil
	} else {
		snap.index = primitive.Build(prims, time.Now())
	}
	c.mu.Lock()
	c.layers[charterRoot] = snap
	c.mu.Unlock()
}

// getLayer returns a snapshot of the cached primitives + index for a
// single layer. Cold-miss triggers a synchronous walk so the first
// request never sees empty data. Concurrent cold-misses on the same
// layer collapse to a single walk via coldMu.
func (c *primitiveCache) getLayer(charterRoot string) ([]primitive.Primitive, primitive.Index, error) {
	c.mu.RLock()
	snap, ok := c.layers[charterRoot]
	c.mu.RUnlock()
	if !ok {
		// Per-key serialization: at most one walker per charterRoot.
		mIface, _ := c.coldMu.LoadOrStore(charterRoot, &sync.Mutex{})
		m := mIface.(*sync.Mutex)
		m.Lock()
		// Recheck under per-key lock — a sibling goroutine may have
		// already filled the cache while we waited.
		c.mu.RLock()
		snap, ok = c.layers[charterRoot]
		c.mu.RUnlock()
		if !ok {
			c.refreshLayer(charterRoot)
			c.mu.RLock()
			snap = c.layers[charterRoot]
			c.mu.RUnlock()
		}
		m.Unlock()
	}
	if snap == nil {
		return nil, primitive.Index{}, nil
	}
	return snap.primitives, snap.index, snap.err
}

// refresh is the project-default convenience wrapper used by callers
// that only care about the project layer (loadPrimitives, buildIndex).
func (c *primitiveCache) refresh() {
	c.refreshLayer(config.DefaultCharterRoot)
}

// get is the project-default convenience wrapper. Same contract as
// getLayer(DefaultCharterRoot).
func (c *primitiveCache) get() ([]primitive.Primitive, primitive.Index, error) {
	return c.getLayer(config.DefaultCharterRoot)
}

// knownLayerRoots returns the charter roots that should be cached —
// project + every declared policy. Recomputed per call so policy
// changes mid-run are picked up.
func (s *server) knownLayerRoots() []string {
	roots := []string{config.DefaultCharterRoot}
	cfg, _ := config.ReadProjectConfig(s.projectDir)
	if cfg == nil {
		return roots
	}
	for _, p := range flattenPolicies(cfg.Policies) {
		roots = append(roots, filepath.Join(config.DefaultCharterRoot, policies.PolicyRoot, p.Name))
	}
	return roots
}

// refreshAllLayers re-walks every known layer. Used as the
// fsWatcher's debounced callback so a single rapid save burst
// re-warms all layers exactly once; request goroutines stay on the
// cached snapshot.
func (s *server) refreshAllLayers() {
	for _, root := range s.knownLayerRoots() {
		s.primitiveCache.refreshLayer(root)
	}
}

// startCacheRefreshers kicks off the primitiveCache background refresh:
// it rebuilds on every fsWatcher publish (already debounced) plus a slow
// ticker fallback. One synchronous fill happens before this returns so the
// first request never sees an empty snapshot.
func (s *server) startCacheRefreshers(ctx context.Context) {
	s.refreshAllLayers()

	// Slow safety-net refresh for the primitive cache in case
	// fsnotify drops events (rare but possible under heavy churn or
	// on network filesystems).
	go func() {
		t := time.NewTicker(2 * time.Minute)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				s.refreshAllLayers()
			}
		}
	}()

	// Optional: log refresher startup once. Useful when diagnosing a
	// dashboard that looks stale — confirms the loop is running.
	fmt.Fprintf(os.Stderr, "keystone web: primitive cache warm\n")
}
