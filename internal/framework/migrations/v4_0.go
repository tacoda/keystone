package migrations

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tacoda/keystone/internal/framework/primitive"
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

// planDown_4_0 reverses the root rename. Down is best-effort (like 3.0):
// it does NOT un-fold sensors back into hooks — a folded sensor is
// indistinguishable from a hand-authored one — nor restore context.json.
// Reversing the .charter/ → .harness/ move is enough to get back on a
// 3.0-compatible layout; re-running Up re-folds idempotently.
func planDown_4_0(absDir string) (*Plan, error) {
	p := &Plan{}
	p.Add("rename the charter root .charter/ → .harness/ (reverse 4.0; fold + context.json not reversed)", func(absDir string) error {
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

// foldOneHook_4_0 rewrites one hook file to a sensor and moves it. The
// rewrites (kind: hook → sensor, event: → on:, id: hooks/… → sensors/…)
// apply to the FRONTMATTER ONLY — never the markdown body, so prose or
// example YAML that mentions `event:` is left untouched.
func foldOneHook_4_0(hooksDir, sensorsDir, name string) error {
	src := filepath.Join(hooksDir, name)
	dst := filepath.Join(sensorsDir, name)
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	out := foldHookText(string(data))
	// Idempotent clobber: an identical dst means a prior run wrote it but
	// crashed before removing src — finish the move rather than erroring.
	if existing, rerr := os.ReadFile(dst); rerr == nil {
		if string(existing) != out {
			return fmt.Errorf("fold hook %s: %s already exists with different content — resolve by hand", name, dst)
		}
		return os.Remove(src)
	}
	if err := os.WriteFile(dst, []byte(out), 0o644); err != nil {
		return err
	}
	return os.Remove(src)
}

// foldHookText applies the hook→sensor frontmatter rewrites, leaving the
// body untouched. If there's no frontmatter block, returns the input
// unchanged (nothing safe to rewrite).
func foldHookText(text string) string {
	fm, body, ok := primitive.SplitFrontmatter(text)
	if !ok {
		return text
	}
	fm = reHookKind_4_0.ReplaceAllString(fm, "${1}kind: sensor")
	fm = reEventKey_4_0.ReplaceAllString(fm, "${1}on:")
	fm = reHookID_4_0.ReplaceAllString(fm, "${1}id: sensors/")
	return "---\n" + fm + "\n---" + body
}

var (
	reHookKind_4_0 = regexp.MustCompile(`(?m)^(\s*)kind:\s*hook\s*$`)
	reEventKey_4_0 = regexp.MustCompile(`(?m)^(\s*)event:`)
	reHookID_4_0   = regexp.MustCompile(`(?m)^(\s*)id:\s*hooks/`)
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
