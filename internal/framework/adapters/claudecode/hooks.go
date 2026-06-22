// Package claudecode is the host adapter for Claude Code. It projects
// keystone primitives and the project's keystone.json hooks block
// into the `.claude/` host-native layout that Claude Code reads
// directly.
//
// Architectural note: keystone is agent-agnostic in source; adapters
// like this one own the agent-specific projection. The host adapter is
// the only place that knows what `.claude/settings.json` even looks
// like. Other hosts (Cursor, Aider, Continue) get their own packages
// alongside this one — same shape, different output.
package claudecode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

// SettingsRelPath is the project-local Claude Code settings file the
// hooks projection merges into. Re-exported so callers know where the
// write lands.
const SettingsRelPath = ".claude/settings.json"

// HookEntryStatusPrefix marks every projector-managed hook entry's
// statusMessage. The projector recognizes its own entries by this
// prefix on subsequent runs, stripping them before re-emit so the
// merge is idempotent. User-authored hooks (any statusMessage that
// doesn't start with this prefix) are left alone.
const HookEntryStatusPrefix = "keystone:"

// HookProjectionResult records what happened when ProjectHooks ran.
type HookProjectionResult struct {
	Path    string // SettingsRelPath, repo-relative
	Wrote   bool   // true if the file's content changed
	Added   int    // hook entries the projector added in this run
	Removed int    // pre-existing keystone:* entries removed before re-emit
}

// hookEntryFromSensor is the in-adapter shape: a flattened view of one
// sensor's HostTrigger plus the owning sensor's id (used for the
// statusMessage marker).
type hookEntryFromSensor struct {
	SensorID string
	primitive.HostTrigger
}

// ProjectHooks merges every sensor's host_triggers into
// .claude/settings.json. Idempotent + additive: keystone-managed
// entries (statusMessage prefix "keystone:") are removed and
// re-emitted from the current sensor frontmatter; everything else
// (other top-level keys, user-authored hook entries) is preserved
// byte-for-byte where possible and structurally otherwise.
//
// Source of truth is the per-sensor frontmatter under
// `.keystone/harness/sensors/`. The adapter receives the already-walked
// primitive slice from `keystone project` — no second filesystem walk.
//
// The settings file is created on first run if absent. Atomic write
// via same-dir temp + rename — the user's settings never observe a
// partial state.
func ProjectHooks(projectDir string, primitives []primitive.Primitive) (HookProjectionResult, error) {
	rel := SettingsRelPath
	abs := filepath.Join(projectDir, rel)

	existing, err := readSettings(abs)
	if err != nil {
		return HookProjectionResult{Path: rel}, fmt.Errorf("read %s: %w", rel, err)
	}

	removed := stripManagedHooks(existing)
	entries := collectSensorTriggers(primitives)
	added := injectHooks(existing, entries)

	out, err := marshalSettings(existing)
	if err != nil {
		return HookProjectionResult{Path: rel}, fmt.Errorf("marshal %s: %w", rel, err)
	}

	prev, _ := os.ReadFile(abs)
	if bytes.Equal(prev, out) {
		return HookProjectionResult{Path: rel, Wrote: false, Added: added, Removed: removed}, nil
	}

	if err := atomicWrite(abs, out); err != nil {
		return HookProjectionResult{Path: rel}, err
	}
	return HookProjectionResult{Path: rel, Wrote: true, Added: added, Removed: removed}, nil
}

// collectSensorTriggers flattens every sensor's host_triggers into a
// linear slice of (sensor_id, trigger) entries the injector can group
// by (phase, matcher). Only kind=sensor primitives are considered —
// other primitives may legally carry frontmatter fields, but only
// sensors map to host hooks.
func collectSensorTriggers(primitives []primitive.Primitive) []hookEntryFromSensor {
	var out []hookEntryFromSensor
	for _, p := range primitives {
		if primitive.Kind(p.Kind) != primitive.KindSensor {
			continue
		}
		for _, t := range p.HostTriggers {
			out = append(out, hookEntryFromSensor{SensorID: p.ID, HostTrigger: t})
		}
	}
	return out
}

// readSettings parses .claude/settings.json into a generic map. Missing
// file returns an empty map (not an error) — first-run is the common
// case.
func readSettings(absPath string) (map[string]any, error) {
	data, err := os.ReadFile(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{}, nil
		}
		return nil, err
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return map[string]any{}, nil
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	if m == nil {
		m = map[string]any{}
	}
	return m, nil
}

// stripManagedHooks removes every hook entry the projector previously
// emitted. Returns the count of removed entries. Walks the
// (hooks → phase → []matcherGroup → matcherGroup.hooks → []hookCmd)
// shape Claude Code expects, dropping any hookCmd whose statusMessage
// starts with HookEntryStatusPrefix. Matcher groups left with zero
// hooks are also removed; phases left with zero groups are dropped.
func stripManagedHooks(settings map[string]any) int {
	hooksAny, ok := settings["hooks"]
	if !ok {
		return 0
	}
	hooksMap, ok := hooksAny.(map[string]any)
	if !ok {
		return 0
	}
	removed := 0
	for phase, groupsAny := range hooksMap {
		groups, ok := groupsAny.([]any)
		if !ok {
			continue
		}
		keptGroups := make([]any, 0, len(groups))
		for _, gAny := range groups {
			g, ok := gAny.(map[string]any)
			if !ok {
				keptGroups = append(keptGroups, gAny)
				continue
			}
			cmdsAny, hasCmds := g["hooks"]
			if !hasCmds {
				keptGroups = append(keptGroups, g)
				continue
			}
			cmds, ok := cmdsAny.([]any)
			if !ok {
				keptGroups = append(keptGroups, g)
				continue
			}
			keptCmds := make([]any, 0, len(cmds))
			for _, cAny := range cmds {
				c, ok := cAny.(map[string]any)
				if !ok {
					keptCmds = append(keptCmds, cAny)
					continue
				}
				if sm, ok := c["statusMessage"].(string); ok && strings.HasPrefix(sm, HookEntryStatusPrefix) {
					removed++
					continue
				}
				keptCmds = append(keptCmds, c)
			}
			if len(keptCmds) == 0 {
				continue
			}
			g["hooks"] = keptCmds
			keptGroups = append(keptGroups, g)
		}
		if len(keptGroups) == 0 {
			delete(hooksMap, phase)
		} else {
			hooksMap[phase] = keptGroups
		}
	}
	if len(hooksMap) == 0 {
		delete(settings, "hooks")
	}
	return removed
}

// injectHooks adds every sensor trigger into the settings map.
// Returns the count added.
//
// Grouping: each phase carries a list of matcher groups, each group
// carries one or more hookCmd entries. The projector groups triggers
// by (phase, matcher) — multiple triggers sharing both reuse the same
// group, matching Claude Code's preferred shape and keeping the file
// compact.
func injectHooks(settings map[string]any, entries []hookEntryFromSensor) int {
	if len(entries) == 0 {
		return 0
	}
	hooksMap, _ := settings["hooks"].(map[string]any)
	if hooksMap == nil {
		hooksMap = map[string]any{}
	}

	sorted := make([]hookEntryFromSensor, len(entries))
	copy(sorted, entries)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Phase != sorted[j].Phase {
			return sorted[i].Phase < sorted[j].Phase
		}
		if sorted[i].Matcher != sorted[j].Matcher {
			return sorted[i].Matcher < sorted[j].Matcher
		}
		return sorted[i].SensorID < sorted[j].SensorID
	})

	added := 0
	for _, e := range sorted {
		groups, _ := hooksMap[e.Phase].([]any)
		group := findOrCreateMatcherGroup(&groups, e.Matcher)
		cmds, _ := group["hooks"].([]any)
		cmds = append(cmds, hookEntry(e))
		group["hooks"] = cmds
		hooksMap[e.Phase] = groups
		added++
	}
	settings["hooks"] = hooksMap
	return added
}

// findOrCreateMatcherGroup returns the matcher group for a given
// matcher within a phase's group list, or appends a new one (mutating
// the slice via the supplied pointer) and returns that. Centralising
// this keeps injectHooks readable.
func findOrCreateMatcherGroup(groups *[]any, matcher string) map[string]any {
	for _, gAny := range *groups {
		g, ok := gAny.(map[string]any)
		if !ok {
			continue
		}
		if existing, _ := g["matcher"].(string); existing == matcher {
			return g
		}
	}
	g := map[string]any{
		"matcher": matcher,
		"hooks":   []any{},
	}
	*groups = append(*groups, g)
	return g
}

// hookEntry builds the per-hook map Claude Code expects.
func hookEntry(e hookEntryFromSensor) map[string]any {
	timeout := e.Timeout
	if timeout <= 0 {
		timeout = 5
	}
	entry := map[string]any{
		"type":          "command",
		"shell":         "bash",
		"command":       e.Command,
		"timeout":       timeout,
		"statusMessage": HookEntryStatusPrefix + e.SensorID,
	}
	return entry
}

// marshalSettings serializes the map with 2-space indent + trailing
// newline. Matches the shape `keystone init` writes for keystone.json
// (and the shape Claude Code's own tooling tends to write), so the diff
// from a user's manual edit is minimal on first projection.
func marshalSettings(settings map[string]any) ([]byte, error) {
	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(out, '\n'), nil
}

// atomicWrite is the same temp+rename shape primitive.copyOne uses;
// duplicated here to avoid pulling primitive as a dep just for one
// helper. The OS guarantees the rename is atomic on the same fs.
func atomicWrite(destAbs string, contents []byte) error {
	if err := os.MkdirAll(filepath.Dir(destAbs), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(destAbs), err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(destAbs), ".keystone-adapter.*")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(contents); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("close temp: %w", err)
	}
	if err := os.Rename(tmpName, destAbs); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("rename %s -> %s: %w", tmpName, destAbs, err)
	}
	return nil
}
