package claudecode

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

// PostureProjectionResult records what happened when ProjectPosture ran.
type PostureProjectionResult struct {
	Path  string // SettingsRelPath, repo-relative
	Wrote bool   // true if the file's content changed
	Added int    // permission entries newly merged in this run
}

// permissionKeys are the three Claude Code permission buckets a posture
// projects into, in stable order.
var permissionKeys = []string{"allow", "ask", "deny"}

// ProjectPosture merges every posture's allow/ask/deny lists into
// .claude/settings.json's `permissions` block. Additive + idempotent: the
// union with any existing (user-authored) entries is deduped and sorted, so a
// re-run is a no-op and user permissions are never dropped. Permission strings
// carry no ownership marker, so this projection does not remove entries — a
// removed posture leaves its grants until a future prune.
func ProjectPosture(projectDir string, primitives []primitive.Primitive) (PostureProjectionResult, error) {
	rel := SettingsRelPath
	abs := filepath.Join(projectDir, rel)

	existing, err := readSettings(abs)
	if err != nil {
		return PostureProjectionResult{Path: rel}, fmt.Errorf("read %s: %w", rel, err)
	}

	added := applyPermissions(existing, collectPosture(primitives))

	out, err := marshalSettings(existing)
	if err != nil {
		return PostureProjectionResult{Path: rel}, fmt.Errorf("marshal %s: %w", rel, err)
	}
	prev, _ := os.ReadFile(abs)
	if bytes.Equal(prev, out) {
		return PostureProjectionResult{Path: rel, Wrote: false, Added: added}, nil
	}
	if err := atomicWrite(abs, out); err != nil {
		return PostureProjectionResult{Path: rel}, err
	}
	return PostureProjectionResult{Path: rel, Wrote: true, Added: added}, nil
}

// collectPosture gathers the allow/ask/deny lists across all posture
// primitives, keyed by permission bucket.
func collectPosture(primitives []primitive.Primitive) map[string][]string {
	want := map[string][]string{}
	for _, p := range primitives {
		if primitive.Kind(p.Kind) != primitive.KindPosture {
			continue
		}
		want["allow"] = append(want["allow"], p.Allow...)
		want["ask"] = append(want["ask"], p.Ask...)
		want["deny"] = append(want["deny"], p.Deny...)
	}
	return want
}

// applyPermissions merges the wanted permission buckets into settings,
// returning the count of newly added entries.
func applyPermissions(settings map[string]any, want map[string][]string) int {
	perm, _ := settings["permissions"].(map[string]any)
	if perm == nil {
		perm = map[string]any{}
	}
	added := 0
	for _, key := range permissionKeys {
		if len(want[key]) == 0 {
			continue
		}
		merged, n := unionSorted(stringSlice(perm[key]), want[key])
		perm[key] = anySlice(merged)
		added += n
	}
	if len(perm) > 0 {
		settings["permissions"] = perm
	}
	return added
}

// unionSorted returns the deduped, sorted union of existing and incoming, plus
// the count of entries not already present in existing.
func unionSorted(existing, incoming []string) ([]string, int) {
	set := map[string]bool{}
	for _, s := range existing {
		set[s] = true
	}
	added := 0
	for _, s := range incoming {
		if !set[s] {
			added++
			set[s] = true
		}
	}
	out := make([]string, 0, len(set))
	for s := range set {
		out = append(out, s)
	}
	sort.Strings(out)
	return out, added
}

// stringSlice coerces a settings value (decoded as []any) into []string,
// dropping non-string entries.
func stringSlice(v any) []string {
	raw, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, e := range raw {
		if s, ok := e.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// anySlice lifts []string back to []any for the settings map.
func anySlice(ss []string) []any {
	out := make([]any, len(ss))
	for i, s := range ss {
		out[i] = s
	}
	return out
}
