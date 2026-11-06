package primitive

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Build assembles an Index from a sorted slice of primitives. The
// caller controls the generated timestamp so tests can pin it; pass
// time.Now().UTC() in production.
func Build(primitives []Primitive, generatedAt time.Time) Index {
	idx := Index{
		Version:    IndexVersion,
		Generated:  generatedAt.UTC().Format(time.RFC3339),
		Primitives: primitives,
		ByKind:     map[string][]string{},
		ByCategory: map[string][]string{},
		ByPath:     map[string][]string{},
	}
	for _, p := range primitives {
		idx.ByKind[p.Kind] = append(idx.ByKind[p.Kind], p.ID)
		if c := CategoryOf(Kind(p.Kind)); c != "" {
			idx.ByCategory[string(c)] = append(idx.ByCategory[string(c)], p.Kind+"/"+p.ID)
		}
		for _, g := range p.Paths {
			ref := p.Kind + "/" + p.ID
			idx.ByPath[g] = append(idx.ByPath[g], ref)
		}
	}
	for k := range idx.ByKind {
		sort.Strings(idx.ByKind[k])
	}
	for c := range idx.ByCategory {
		sort.Strings(idx.ByCategory[c])
	}
	for g := range idx.ByPath {
		sort.Strings(idx.ByPath[g])
	}
	return idx
}

// Write serializes the index as pretty-printed JSON to outPath. The
// file is written atomically via os.Rename — partial writes never
// leave a half-baked INDEX.json on disk for a running agent to read.
func Write(outPath string, idx Index) error {
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal index: %w", err)
	}
	data = append(data, '\n')
	return atomicWriteJSON(outPath, data)
}

// LiteEntry is the minimal descriptor written to INDEX.lite.json. Cheap
// discovery surface — category/kind/id/description per primitive.
// The agent reads this on session start to pick which primitives to
// open; the full INDEX.json (with paths, globs, triggers, traces) is
// read only when the agent needs to activate a primitive.
type LiteEntry struct {
	Category    string `json:"category"`
	Kind        string `json:"kind"`
	ID          string `json:"id"`
	Description string `json:"description"`
}

// LiteIndex is the serialized shape of INDEX.lite.json. Sorted by
// category (constraints first), then kind, then id, so diffs stay readable.
type LiteIndex struct {
	Version    string      `json:"version"`
	Generated  string      `json:"generated"`
	Primitives []LiteEntry `json:"primitives"`
}

// BuildLite extracts the lite descriptor set from a full index. The
// only field that varies between runs is the generated timestamp; the
// content order is stable.
func BuildLite(idx Index) LiteIndex {
	out := LiteIndex{
		Version:    idx.Version,
		Generated:  idx.Generated,
		Primitives: make([]LiteEntry, 0, len(idx.Primitives)),
	}
	for _, p := range idx.Primitives {
		out.Primitives = append(out.Primitives, LiteEntry{
			Category: string(CategoryOf(Kind(p.Kind))),
			Kind:     p.Kind, ID: p.ID, Description: p.Description,
		})
	}
	sort.SliceStable(out.Primitives, func(i, j int) bool {
		if out.Primitives[i].Category != out.Primitives[j].Category {
			return categoryRank(out.Primitives[i].Category) < categoryRank(out.Primitives[j].Category)
		}
		if out.Primitives[i].Kind != out.Primitives[j].Kind {
			return out.Primitives[i].Kind < out.Primitives[j].Kind
		}
		return out.Primitives[i].ID < out.Primitives[j].ID
	})
	return out
}

// SortByCategory orders primitives constraints first, then capabilities,
// each kind in KnownKinds order, then by id. A retired kind sorts last.
func SortByCategory(ps []Primitive) {
	sort.SliceStable(ps, func(i, j int) bool {
		if a, b := KindRank(ps[i].Kind), KindRank(ps[j].Kind); a != b {
			return a < b
		}
		return ps[i].ID < ps[j].ID
	})
}

// KindRank is kind's position in KnownKinds (constraints first); a retired
// kind ranks last.
func KindRank(kind string) int {
	for i, k := range KnownKinds {
		if string(k) == kind {
			return i
		}
	}
	return len(KnownKinds)
}

// CategoryTitle is a category's heading: "Constraints", "Capabilities", or
// "Other" for a retired kind.
func CategoryTitle(kind string) string {
	switch CategoryOf(Kind(kind)) {
	case CategoryConstraint:
		return "Constraints"
	case CategoryCapability:
		return "Capabilities"
	}
	return "Other"
}

// categoryRank orders constraints before capabilities, and anything
// uncategorized (a retired kind) last.
func categoryRank(c string) int {
	switch Category(c) {
	case CategoryConstraint:
		return 0
	case CategoryCapability:
		return 1
	}
	return 2
}

// WriteLite serializes the lite index to outPath, atomic-rename style.
func WriteLite(outPath string, lite LiteIndex) error {
	data, err := json.MarshalIndent(lite, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal lite index: %w", err)
	}
	data = append(data, '\n')
	return atomicWriteJSON(outPath, data)
}

// atomicWriteJSON centralizes the temp+rename shape both Write and
// WriteLite use. Single helper, two callers, no behavior split.
func atomicWriteJSON(outPath string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(outPath), err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(outPath), ".INDEX.json.*")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("close temp: %w", err)
	}
	if err := os.Rename(tmpName, outPath); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("rename %s -> %s: %w", tmpName, outPath, err)
	}
	return nil
}
