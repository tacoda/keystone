package web

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/mcp"
	"github.com/tacoda/keystone/internal/framework/policies"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// healthRefreshInterval is how often the background ticker re-probes
// every configured source's health. 30s strikes a balance: stale
// enough that handlers never block, fresh enough that the dashboard
// catches a flapping source within a single user attention span.
const healthRefreshInterval = 30 * time.Second

// healthProbeTimeout caps a single source's probe. Independent of any
// per-adapter timeout — if an adapter ignores ctx, this still bounds
// the refresh wall-clock.
const healthProbeTimeout = 8 * time.Second

// healthProbeConcurrency bounds the worker pool that fans out probes.
// Most installs have ≤4 sources; the cap mainly protects against a
// pathological context.json with dozens of entries.
const healthProbeConcurrency = 8

// walkFn is the indirect entry to primitive.Walk so tests can
// observe and count layer refreshes. Production code must not
// reassign it outside tests.
var walkFn = primitive.Walk

// layerSnapshot is one cache entry — the most recent walk of a
// single harness root.
type layerSnapshot struct {
	primitives []primitive.Primitive
	index      primitive.Index
	err        error
}

// primitiveCache holds the most recent walk of every accessed
// harness layer — project root plus every declared policy root.
// Page handlers read from here instead of re-walking on every
// request. The fsWatcher refreshes dirty layers on debounced file
// changes; other layers stay warm.
type primitiveCache struct {
	projectDir string

	mu     sync.RWMutex
	layers map[string]*layerSnapshot // key: harnessRoot (filepath-relative)

	// coldMu serializes cold-miss walks per harnessRoot key, so two
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

// refreshLayer walks a single harness root and replaces its cached
// snapshot. Safe to call from any goroutine.
func (c *primitiveCache) refreshLayer(harnessRoot string) {
	prims, _, err := walkFn(c.projectDir, harnessRoot)
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
	c.layers[harnessRoot] = snap
	c.mu.Unlock()
}

// getLayer returns a snapshot of the cached primitives + index for a
// single layer. Cold-miss triggers a synchronous walk so the first
// request never sees empty data. Concurrent cold-misses on the same
// layer collapse to a single walk via coldMu.
func (c *primitiveCache) getLayer(harnessRoot string) ([]primitive.Primitive, primitive.Index, error) {
	c.mu.RLock()
	snap, ok := c.layers[harnessRoot]
	c.mu.RUnlock()
	if !ok {
		// Per-key serialization: at most one walker per harnessRoot.
		mIface, _ := c.coldMu.LoadOrStore(harnessRoot, &sync.Mutex{})
		m := mIface.(*sync.Mutex)
		m.Lock()
		// Recheck under per-key lock — a sibling goroutine may have
		// already filled the cache while we waited.
		c.mu.RLock()
		snap, ok = c.layers[harnessRoot]
		c.mu.RUnlock()
		if !ok {
			c.refreshLayer(harnessRoot)
			c.mu.RLock()
			snap = c.layers[harnessRoot]
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
	c.refreshLayer(config.DefaultHarnessRoot)
}

// get is the project-default convenience wrapper. Same contract as
// getLayer(DefaultHarnessRoot).
func (c *primitiveCache) get() ([]primitive.Primitive, primitive.Index, error) {
	return c.getLayer(config.DefaultHarnessRoot)
}

// healthCache holds the most recent source-list snapshot with each
// source's probed health. Page handlers read from here instead of
// triggering a synchronous fan-out of HTTP probes on every render.
// A background ticker keeps the snapshot fresh.
type healthCache struct {
	projectDir string

	mu        sync.RWMutex
	entries   []sourceEntry
	lastError error
}

func newHealthCache(projectDir string) *healthCache {
	return &healthCache{projectDir: projectDir}
}

// refresh re-reads context.json and probes every configured source's
// health in parallel under a bounded worker pool. One slow source
// can no longer hold up the whole refresh — the per-probe timeout
// fires independently per goroutine.
func (c *healthCache) refresh(ctx context.Context) {
	cfg, err := mcp.LoadContextConfig(c.projectDir)
	if err != nil {
		c.mu.Lock()
		c.lastError = err
		c.mu.Unlock()
		return
	}
	if cfg == nil {
		c.mu.Lock()
		c.entries = []sourceEntry{}
		c.lastError = nil
		c.mu.Unlock()
		return
	}

	results := make([]sourceEntry, len(cfg.Sources))
	sem := make(chan struct{}, healthProbeConcurrency)
	var wg sync.WaitGroup
	for i, src := range cfg.Sources {
		i, src := i, src
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			a := mcp.BuildAdapter(src)
			probeCtx, cancel := context.WithTimeout(ctx, healthProbeTimeout)
			defer cancel()
			h := a.Health(probeCtx)
			results[i] = sourceEntry{
				Name:   a.Name(),
				Type:   a.Type(),
				Health: h,
			}
		}()
	}
	wg.Wait()

	c.mu.Lock()
	c.entries = results
	c.lastError = nil
	c.mu.Unlock()
}

// get returns a snapshot of the cached source entries. Caller treats
// the slice as read-only.
func (c *healthCache) get() ([]sourceEntry, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.entries, c.lastError
}

// knownLayerRoots returns the harness roots that should be cached —
// project + every declared policy. Recomputed per call so policy
// changes mid-run are picked up.
func (s *server) knownLayerRoots() []string {
	roots := []string{config.DefaultHarnessRoot}
	cfg, _ := config.ReadProjectConfig(s.projectDir)
	if cfg == nil {
		return roots
	}
	for _, p := range flattenPolicies(cfg.Policies) {
		roots = append(roots, filepath.Join(config.DefaultHarnessRoot, policies.PolicyRoot, p.Name))
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

// startCacheRefreshers kicks off the background refresh loops:
//
//   - primitiveCache rebuilds on every fsWatcher publish (already
//     debounced) plus a slow ticker fallback.
//   - healthCache rebuilds on a 30s ticker. Source-mutation handlers
//     also call s.healthCache.refresh directly for immediate feedback.
//
// Both caches do one synchronous fill before this returns so the
// first request never sees an empty snapshot.
func (s *server) startCacheRefreshers(ctx context.Context) {
	s.refreshAllLayers()
	s.healthCache.refresh(ctx)

	go func() {
		t := time.NewTicker(healthRefreshInterval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				s.healthCache.refresh(ctx)
			}
		}
	}()

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
	fmt.Fprintf(os.Stderr, "keystone web: caches warm (primitive + health)\n")
}
