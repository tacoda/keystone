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
		ByGlob:     map[string][]string{},
	}
	for _, p := range primitives {
		idx.ByKind[p.Kind] = append(idx.ByKind[p.Kind], p.ID)
		for _, g := range p.Globs {
			ref := p.Kind + "/" + p.ID
			idx.ByGlob[g] = append(idx.ByGlob[g], ref)
		}
	}
	for k := range idx.ByKind {
		sort.Strings(idx.ByKind[k])
	}
	for g := range idx.ByGlob {
		sort.Strings(idx.ByGlob[g])
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
// discovery surface — kind/id/description per primitive, nothing else.
// The agent reads this on session start to pick which primitives to
// open; the full INDEX.json (with paths, globs, triggers, traces) is
// read only when the agent needs to activate a primitive.
type LiteEntry struct {
	Kind        string `json:"kind"`
	ID          string `json:"id"`
	Description string `json:"description"`
}

// LiteIndex is the serialized shape of INDEX.lite.json. Sorted by kind
// then id so diffs stay readable across runs.
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
			Kind: p.Kind, ID: p.ID, Description: p.Description,
		})
	}
	sort.SliceStable(out.Primitives, func(i, j int) bool {
		if out.Primitives[i].Kind != out.Primitives[j].Kind {
			return out.Primitives[i].Kind < out.Primitives[j].Kind
		}
		return out.Primitives[i].ID < out.Primitives[j].ID
	})
	return out
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
