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
// hooks projection merges into.
const SettingsRelPath = ".claude/settings.json"

// HookEntryStatusPrefix marks every projector-managed hook entry's
// statusMessage. The projector recognizes its own entries by this
// prefix and strips them before re-emit so the merge is idempotent.
// User-authored hooks are left alone.
const HookEntryStatusPrefix = "keystone:"

// HookProjectionResult records what happened when ProjectHooks ran.
type HookProjectionResult struct {
	Path    string // SettingsRelPath, repo-relative
	Wrote   bool   // true if the file's content changed
	Added   int    // bridge entries added this run (one per host phase)
	Removed int    // pre-existing keystone:* entries removed before re-emit
}

// ProjectHooks installs the single host→keystone bridge into
// .claude/settings.json: one generic entry per host phase that any
// `kind: hook` binds to, each running `keystone hook fire <phase>`.
// keystone then dispatches the matching hooks itself.
//
// This is the deliberate design (see the 3.0 hook layer): hooks are
// framework-owned and too host-divergent to map individually, so the
// adapter writes no per-hook entries — just the bridge. Framework
// events (pre-verify, on-gate, …) are keystone-fired and never bridged.
//
// Idempotent + additive: keystone-managed entries (statusMessage prefix
// "keystone:") are stripped and re-emitted; everything else (other keys,
// user-authored hooks) is preserved. Atomic write via temp + rename.
func ProjectHooks(projectDir string, primitives []primitive.Primitive) (HookProjectionResult, error) {
	rel := SettingsRelPath
	abs := filepath.Join(projectDir, rel)

	existing, err := readSettings(abs)
	if err != nil {
		return HookProjectionResult{Path: rel}, fmt.Errorf("read %s: %w", rel, err)
	}

	removed := stripManagedHooks(existing)
	added := injectBridges(existing, hostPhases(primitives))

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

// hostPhases returns the distinct host-phase events any deterministically-
// firing primitive binds to (sorted, deduped) — a `hook`, a computational
// `guide` (LSP), or a computational `sensor` (gate check), unified via
// primitive.HookFire. A framework event (pre-verify, on-gate, …) is
// keystone-fired and never reaches the host, so it is not bridged.
func hostPhases(primitives []primitive.Primitive) []string {
	seen := map[string]bool{}
	var out []string
	for _, p := range primitives {
		event, _, ok := primitive.HookFire(p)
		if !ok || primitive.IsFrameworkEvent(event) || seen[event] {
			continue
		}
		seen[event] = true
		out = append(out, event)
	}
	sort.Strings(out)
	return out
}

// injectBridges writes one bridge entry per host phase, each running
// `keystone hook fire <phase>`. Returns the count of phases bridged.
func injectBridges(settings map[string]any, phases []string) int {
	if len(phases) == 0 {
		return 0
	}
	hooksMap, _ := settings["hooks"].(map[string]any)
	if hooksMap == nil {
		hooksMap = map[string]any{}
	}
	for _, phase := range phases {
		hooksMap[phase] = []any{
			map[string]any{
				"matcher": "",
				"hooks": []any{
					map[string]any{
						"type":          "command",
						"shell":         "bash",
						"command":       "keystone hook fire " + phase,
						"statusMessage": HookEntryStatusPrefix + "bridge:" + phase,
					},
				},
			},
		}
	}
	settings["hooks"] = hooksMap
	return len(phases)
}

// readSettings parses .claude/settings.json into a generic map. Missing
// file returns an empty map (not an error) — first-run is the common case.
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
// emitted (statusMessage prefix "keystone:"). Returns the count removed.
// Matcher groups left with zero hooks are removed; phases left with zero
// groups are dropped; an empty hooks map is removed entirely.
func stripManagedHooks(settings map[string]any) int {
	hooksMap, ok := settings["hooks"].(map[string]any)
	if !ok {
		return 0
	}
	removed := 0
	for phase, groupsAny := range hooksMap {
		groups, ok := groupsAny.([]any)
		if !ok {
			continue
		}
		kept, n := stripPhaseGroups(groups)
		removed += n
		if len(kept) == 0 {
			delete(hooksMap, phase)
		} else {
			hooksMap[phase] = kept
		}
	}
	if len(hooksMap) == 0 {
		delete(settings, "hooks")
	}
	return removed
}

// stripPhaseGroups drops managed hook commands from a phase's matcher
// groups, returning the surviving groups and the count removed.
func stripPhaseGroups(groups []any) ([]any, int) {
	removed := 0
	kept := make([]any, 0, len(groups))
	for _, gAny := range groups {
		g, ok := gAny.(map[string]any)
		if !ok {
			kept = append(kept, gAny)
			continue
		}
		cmds, ok := g["hooks"].([]any)
		if !ok {
			kept = append(kept, g)
			continue
		}
		keptCmds, n := stripManagedCmds(cmds)
		removed += n
		if len(keptCmds) == 0 {
			continue
		}
		g["hooks"] = keptCmds
		kept = append(kept, g)
	}
	return kept, removed
}

// stripManagedCmds drops hook commands whose statusMessage marks them as
// projector-managed, returning the survivors and the count removed.
func stripManagedCmds(cmds []any) ([]any, int) {
	removed := 0
	kept := make([]any, 0, len(cmds))
	for _, cAny := range cmds {
		c, ok := cAny.(map[string]any)
		if !ok {
			kept = append(kept, cAny)
			continue
		}
		if sm, ok := c["statusMessage"].(string); ok && strings.HasPrefix(sm, HookEntryStatusPrefix) {
			removed++
			continue
		}
		kept = append(kept, c)
	}
	return kept, removed
}

// marshalSettings serializes the map with 2-space indent + trailing
// newline, HTML escaping off so shell metacharacters round-trip cleanly.
func marshalSettings(settings map[string]any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(settings); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// atomicWrite is the same temp+rename shape primitive.copyOne uses.
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
