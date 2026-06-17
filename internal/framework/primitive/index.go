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
