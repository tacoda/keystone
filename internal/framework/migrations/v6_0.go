package migrations

import (
	"fmt"
	"os"
	"path/filepath"
)

// Version 6.0 — the charter is the codified context of the repository.
// One structural change: the retired `tools/` directory archives (the
// tool kind is removed — keystone manages context, not callables). The
// CLI regenerates INDEX.json and host projections after the plan runs,
// and frontmatter rewrites (`kind: subagent` → `kind: agent`, 4.x
// spellings) stay with `keystone lint --fix` — parse-time normalization
// keeps every tree working until then (iron laws 2 and 6).
//
// Paths are frozen literals, NOT config constants (see v5_0).
func init() {
	Register(Migration{
		Version: "6.0",
		Up:      planUp_6_0,
		Down:    planDown_6_0,
	})
}

const (
	toolsDir_6_0 = ".agents/tools"               // 5.x tool-kind home, retired
	archive_6_0  = ".keystone/archive/6.0/tools" // where it goes
)

func planUp_6_0(absDir string) (*Plan, error) {
	p := &Plan{}
	if dirExists_6(absDir, toolsDir_6_0) {
		if _, err := os.Stat(filepath.Join(absDir, archive_6_0)); err == nil {
			return nil, fmt.Errorf("%s/ already exists — resolve the conflict by hand, then re-run", archive_6_0)
		}
		src, dst := toolsDir_6_0, archive_6_0
		p.Add(fmt.Sprintf("archive %s/ → %s/ (the tool kind is retired; keystone manages context, not callables)", src, dst), func(absDir string) error {
			return moveDir_6_0(filepath.Join(absDir, src), filepath.Join(absDir, dst))
		})
		p.Add("leave a pointer where tools/ lived", func(absDir string) error {
			return writeFile_6_0(filepath.Join(absDir, src, "README.md"), toolsPointer_6_0)
		})
	}
	return p, nil
}

func planDown_6_0(absDir string) (*Plan, error) {
	p := &Plan{}
	if dirExists_6(absDir, archive_6_0) {
		src, dst := archive_6_0, toolsDir_6_0
		p.Add(fmt.Sprintf("restore %s/ from %s/", dst, src), func(absDir string) error {
			if err := os.RemoveAll(filepath.Join(absDir, dst)); err != nil {
				return err
			}
			return moveDir_6_0(filepath.Join(absDir, src), filepath.Join(absDir, dst))
		})
	}
	return p, nil
}

const toolsPointer_6_0 = `# tools/ is retired in 6.0

The ` + "`tool`" + ` kind is removed: keystone manages the context (charter), not
author-defined callables. The former contents are archived at
` + "`.keystone/archive/6.0/tools/`" + `. Delete this folder when you have
re-homed anything you need.
`

// moveDir_6_0 renames a directory, creating the parent first.
func moveDir_6_0(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.Rename(src, dst)
}

// writeFile_6_0 writes text, creating the parent directory.
func writeFile_6_0(path, text string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(text), 0o644)
}

// dirExists_6 reports whether the repo-relative directory exists.
func dirExists_6(absDir, rel string) bool {
	info, err := os.Stat(filepath.Join(absDir, rel))
	return err == nil && info.IsDir()
}
