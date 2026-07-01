package migrations

import (
	"fmt"
	"os"
	"path/filepath"
)

// Version 4.0 — the harness→charter rename. Keystone is the charter
// manager, not a harness, so the single root directory moves from
// .harness/ (3.0) to .charter/ (4.0) — the name now describes the
// charter, not the engine that runs it. The vestigial context.json is
// dropped and INDEX* is rebuilt at the new root.
//
// Paths are frozen literals, NOT config.DefaultCharterRoot: a migration
// reproduces the layout of its own era and must not shift if the
// current default is renamed again later.
func init() {
	Register(Migration{
		Version: "4.0",
		Up:      planUp_4_0,
		Down:    planDown_4_0,
	})
}

const (
	legacyRoot_4_0  = ".harness" // 3.0 single root
	charterRoot_4_0 = ".charter" // 4.0 single root
)

func planUp_4_0(absDir string) (*Plan, error) {
	p := &Plan{}
	if !dirExists(filepath.Join(absDir, legacyRoot_4_0)) && !dirExists(filepath.Join(absDir, charterRoot_4_0)) {
		// Fresh install — scaffolded directly from 4.0 templates.
		return p, nil
	}
	p.Add("rename the charter root .harness/ → .charter/", func(absDir string) error {
		return renameRoot_4_0(absDir, legacyRoot_4_0, charterRoot_4_0)
	})
	p.Add("drop the vestigial context.json", func(absDir string) error {
		if err := os.Remove(filepath.Join(absDir, charterRoot_4_0, "context.json")); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	})
	p.Add("rebuild INDEX.json + INDEX.lite.json at the .charter/ root", func(absDir string) error {
		// indexDir == the charter root: at 4.0 the umbrella IS the root.
		return rebuildIndex(absDir, charterRoot_4_0, charterRoot_4_0)
	})
	return p, nil
}

func planDown_4_0(absDir string) (*Plan, error) {
	p := &Plan{}
	p.Add("rename the charter root .charter/ → .harness/ (reverse 4.0)", func(absDir string) error {
		return renameRoot_4_0(absDir, charterRoot_4_0, legacyRoot_4_0)
	})
	return p, nil
}

// renameRoot_4_0 moves src→dst under absDir. Idempotent: a no-op once
// src is gone. Refuses to clobber an existing dst.
func renameRoot_4_0(absDir, src, dst string) error {
	srcAbs := filepath.Join(absDir, src)
	dstAbs := filepath.Join(absDir, dst)
	if !dirExists(srcAbs) {
		return nil // already moved
	}
	if dirExists(dstAbs) {
		return fmt.Errorf("both %s and %s exist — resolve by hand", src, dst)
	}
	return os.Rename(srcAbs, dstAbs)
}
