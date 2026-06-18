package policies

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// DriftKind describes how a file's on-disk state differs from the lockfile.
type DriftKind string

const (
	DriftMissing  DriftKind = "missing"  // file in lockfile, not on disk
	DriftModified DriftKind = "modified" // file present but content changed
	DriftExtra    DriftKind = "extra"    // file on disk, not in lockfile
)

// Drift is one file-level deviation from the recorded state.
type Drift struct {
	Path string
	Kind DriftKind
}

// Verify walks the installed policy directory and compares per-file hashes
// to the expected set from the lockfile. Returns the list of drifted files
// (empty when clean). A nil expected map combined with an existing policy
// directory counts the whole directory as Extra; an empty installed
// directory with a non-empty expected map counts as all Missing.
//
// Drift paths are returned in stable (sorted) order so callers can
// deterministically render them.
func Verify(name, projectDir, harnessRoot string, expected map[string]string) ([]Drift, error) {
	target := policyDir(projectDir, harnessRoot, name)

	current, err := hashFilesUnder(target, harnessRoot, name)
	if err != nil {
		return nil, fmt.Errorf("hash installed files: %w", err)
	}

	var drifts []Drift
	for path, want := range expected {
		got, found := current[path]
		switch {
		case !found:
			drifts = append(drifts, Drift{Path: path, Kind: DriftMissing})
		case got != want:
			drifts = append(drifts, Drift{Path: path, Kind: DriftModified})
		}
	}
	for path := range current {
		if _, found := expected[path]; !found {
			drifts = append(drifts, Drift{Path: path, Kind: DriftExtra})
		}
	}

	sort.Slice(drifts, func(i, j int) bool { return drifts[i].Path < drifts[j].Path })
	return drifts, nil
}

// hashFilesUnder walks the install directory of one policy and returns a
// map of project-relative slash-paths → sha256 hashes. Missing target
// returns an empty map (not an error) — callers interpret that as "the
// whole policy is gone."
func hashFilesUnder(target, harnessRoot, name string) (map[string]string, error) {
	out := map[string]string{}
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return out, nil
	}
	err := filepath.WalkDir(target, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		hash, err := hashFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(target, path)
		if err != nil {
			return err
		}
		key := filepath.ToSlash(filepath.Join(harnessRoot, PolicyRoot, name, rel))
		out[key] = hash
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}
