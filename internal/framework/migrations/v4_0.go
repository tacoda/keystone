package migrations

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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
	p.Add("fold hooks/ into sensors/ (the hook kind is retired; a check is a sensor)", func(absDir string) error {
		return foldHooksToSensors_4_0(filepath.Join(absDir, charterRoot_4_0))
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

// foldHooksToSensors_4_0 retires the `hook` kind: every file under
// <charter>/hooks/ becomes a `sensor` (a hook was only ever a check or a
// side-effect firing on an event). It rewrites `kind: hook` → `kind:
// sensor` and the `event:` key → `on:`, moves the file to sensors/, then
// removes the empty hooks/ dir. Idempotent — a no-op once hooks/ is gone.
func foldHooksToSensors_4_0(charterAbs string) error {
	hooksDir := filepath.Join(charterAbs, "hooks")
	entries, err := os.ReadDir(hooksDir)
	if os.IsNotExist(err) {
		return nil // already folded
	}
	if err != nil {
		return err
	}
	sensorsDir := filepath.Join(charterAbs, "sensors")
	if err := os.MkdirAll(sensorsDir, 0o755); err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		if err := foldOneHook_4_0(hooksDir, sensorsDir, e.Name()); err != nil {
			return err
		}
	}
	_ = os.Remove(hooksDir) // best-effort; only removes if empty
	return nil
}

// foldOneHook_4_0 rewrites one hook file to a sensor and moves it.
func foldOneHook_4_0(hooksDir, sensorsDir, name string) error {
	src := filepath.Join(hooksDir, name)
	dst := filepath.Join(sensorsDir, name)
	if _, err := os.Stat(dst); err == nil {
		return fmt.Errorf("fold hook %s: %s already exists — resolve by hand", name, dst)
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	out := reHookKind_4_0.ReplaceAllString(string(data), "${1}kind: sensor")
	out = reEventKey_4_0.ReplaceAllString(out, "${1}on:")
	if err := os.WriteFile(dst, []byte(out), 0o644); err != nil {
		return err
	}
	return os.Remove(src)
}

var (
	reHookKind_4_0 = regexp.MustCompile(`(?m)^(\s*)kind:\s*hook\s*$`)
	reEventKey_4_0 = regexp.MustCompile(`(?m)^(\s*)event:`)
)

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
