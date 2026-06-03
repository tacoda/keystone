package main

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Migration is one loaded migration file. ID is the filename without
// extension; Version is the parent directory name (e.g. "0.6.0").
type Migration struct {
	Version    string
	ID         string
	SourcePath string // path within the embedded FS, for error messages

	Description string      `yaml:"description"`
	Operations  []Operation `yaml:"operations"`
}

// Operation is the raw shape of a single op in a migration file. Op-specific
// fields are union-style; only the ones matching Type are populated.
type Operation struct {
	Type string `yaml:"type"`
	Path string `yaml:"path"`

	// add_file
	Content string `yaml:"content,omitempty"`

	// frontmatter_set
	Key   string `yaml:"key,omitempty"`
	Value string `yaml:"value,omitempty"`

	// ensure_section
	AfterHeading string `yaml:"after_heading,omitempty"`
	Heading      string `yaml:"heading,omitempty"`
	Body         string `yaml:"body,omitempty"`

	// replace_block (reuses Heading)
	Match       string `yaml:"match,omitempty"`
	Replacement string `yaml:"replacement,omitempty"`
}

// loadMigrations walks the embedded migrations/ tree and returns every
// migration whose version directory is strictly greater than fromVersion,
// sorted by (version asc, filename asc). fromVersion may be empty or "dev";
// "dev" returns nothing (callers should short-circuit before calling).
func loadMigrations(assets embed.FS, fromVersion string) ([]Migration, error) {
	versionDirs, err := fs.ReadDir(assets, "migrations")
	if err != nil {
		return nil, err
	}

	var out []Migration
	for _, vd := range versionDirs {
		if !vd.IsDir() {
			continue
		}
		v := vd.Name()
		if !looksLikeVersion(v) {
			continue
		}
		if fromVersion != "" && compareSemver(v, fromVersion) <= 0 {
			continue
		}

		dir := path.Join("migrations", v)
		entries, err := fs.ReadDir(assets, dir)
		if err != nil {
			return nil, err
		}
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			n := e.Name()
			if !strings.HasSuffix(n, ".yaml") && !strings.HasSuffix(n, ".yml") {
				continue
			}
			names = append(names, n)
		}
		sort.Strings(names)

		for _, n := range names {
			full := path.Join(dir, n)
			data, err := fs.ReadFile(assets, full)
			if err != nil {
				return nil, fmt.Errorf("read %s: %w", full, err)
			}
			var m Migration
			if err := yaml.Unmarshal(data, &m); err != nil {
				return nil, fmt.Errorf("parse %s: %w", full, err)
			}
			m.Version = v
			m.ID = strings.TrimSuffix(strings.TrimSuffix(n, ".yaml"), ".yml")
			m.SourcePath = full
			out = append(out, m)
		}
	}

	sort.SliceStable(out, func(i, j int) bool {
		if c := compareSemver(out[i].Version, out[j].Version); c != 0 {
			return c < 0
		}
		return out[i].ID < out[j].ID
	})
	return out, nil
}

// looksLikeVersion returns true for dotted-numeric directory names
// (e.g. "0.6.0", "1.2.3"). Anything else is ignored — keeps a templates/
// or examples/ directory from being treated as a release.
func looksLikeVersion(s string) bool {
	if s == "" {
		return false
	}
	for _, part := range strings.Split(s, ".") {
		if part == "" {
			return false
		}
		if _, err := strconv.Atoi(part); err != nil {
			return false
		}
	}
	return true
}

// compareSemver compares two dotted-numeric versions. Returns -1, 0, or 1.
// Non-numeric or malformed inputs sort as 0; callers should pre-validate.
func compareSemver(a, b string) int {
	ap := strings.Split(a, ".")
	bp := strings.Split(b, ".")
	n := len(ap)
	if len(bp) > n {
		n = len(bp)
	}
	for i := 0; i < n; i++ {
		var ai, bi int
		if i < len(ap) {
			ai, _ = strconv.Atoi(ap[i])
		}
		if i < len(bp) {
			bi, _ = strconv.Atoi(bp[i])
		}
		if ai < bi {
			return -1
		}
		if ai > bi {
			return 1
		}
	}
	return 0
}
