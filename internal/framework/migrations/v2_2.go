package migrations

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tacoda/keystone/internal/framework/adapters/agnostic"
	"github.com/tacoda/keystone/internal/framework/adapters/aider"
	"github.com/tacoda/keystone/internal/framework/adapters/claudecode"
	"github.com/tacoda/keystone/internal/framework/adapters/continueide"
	"github.com/tacoda/keystone/internal/framework/adapters/cursor"
	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// Version 2.2 — host-native projection surface lands. No data
// migration required; this version's contract is purely additive:
//
//   - Sensor frontmatter now supports `host_triggers:` — Claude Code
//     hooks projected from sensors instead of from a keystone.json
//     hooks block (which never shipped).
//   - Skill / persona frontmatter gains `model:` and `tools:` fields.
//   - Guide shims emit `kind: rule` + `id: rules/<slug>` frontmatter
//     instead of host-specific shape.
//   - `keystone index` writes a sibling `INDEX.lite.json` with
//     descriptors only (kind + id + description).
//   - `keystone project` emits AGENTS.md unconditionally; cursor /
//     aider / continueide adapters opt-in via keystone.json `adapters:`.
//   - "GOLDEN PATH" sections renamed to "GOLDEN RULE" everywhere.
//
// All of the above are idempotent file-regeneration steps. The
// migration's Up plan walks the existing harness and re-runs the
// index + project pipeline once — same code path `keystone project`
// uses interactively. Existing user content (sensor markdown bodies,
// adapter overlays, project-layer files) is preserved unchanged. Only
// generated artifacts (INDEX.json, INDEX.lite.json, .claude/*,
// AGENTS.md, .cursor/rules/, .aider.*, .continue/rules/) are
// rewritten.
//
// Down: deletes the new generated artifacts the user can safely
// regenerate from 2.1 sources. Preserves source content.

// Legacy 2.x layout, pinned so the pre-3.0 migrations keep operating on
// `.keystone/` no matter where DefaultHarnessRoot points now. The 3.0
// migration (v3_0) moves this tree to `.harness/`.
const (
	legacyHarnessRoot = ".keystone/harness"
	legacyKeystoneDir = ".keystone"
)

func init() {
	Register(Migration{
		Version: "2.2",
		Up:      planUp_2_2,
		Down:    planDown_2_2,
	})
}

func planUp_2_2(absDir string) (*Plan, error) {
	p := &Plan{}

	// 2.2 predates the 3.0 `.harness` move — pin to the legacy layout so
	// this migration keeps operating on `.keystone/` regardless of the
	// current DefaultHarnessRoot.
	harnessRoot := legacyHarnessRoot
	harnessAbs := filepath.Join(absDir, harnessRoot)
	if !dirExists(harnessAbs) {
		// Fresh install — `keystone init` will scaffold from the 2.2
		// templates directly. Nothing to migrate.
		return p, nil
	}

	p.Add("re-walk .keystone/harness/ and rebuild INDEX.json + INDEX.lite.json", func(absDir string) error {
		return rebuildIndex(absDir, harnessRoot, legacyKeystoneDir)
	})

	p.Add("re-project primitives → .claude/{agents,commands,skills,rules}", func(absDir string) error {
		primitives, _, err := primitive.Walk(absDir, harnessRoot)
		if err != nil {
			return err
		}
		_, err = primitive.Project(absDir, primitives)
		return err
	})

	p.Add("project agnostic AGENTS.md at repo root", func(absDir string) error {
		_, err := agnostic.ProjectAgentsMD(absDir, agnostic.DefaultBody())
		return err
	})

	p.Add("merge sensor host_triggers into .claude/settings.json (additive)", func(absDir string) error {
		primitives, _, err := primitive.Walk(absDir, harnessRoot)
		if err != nil {
			return err
		}
		_, err = claudecode.ProjectHooks(absDir, primitives)
		return err
	})

	p.Add("project opt-in adapters from keystone.json (cursor / aider / continue)", func(absDir string) error {
		return runOptInAdapters(absDir, harnessRoot)
	})

	return p, nil
}

func planDown_2_2(absDir string) (*Plan, error) {
	p := &Plan{}

	p.Add("remove .keystone/INDEX.lite.json", func(absDir string) error {
		return removeIfExists(filepath.Join(absDir, legacyKeystoneDir, config.IndexLiteName))
	})

	p.Add("remove generated .claude/rules/ (guide shims)", func(absDir string) error {
		return removeIfExists(filepath.Join(absDir, ".claude", "rules"))
	})

	p.Add("strip keystone-managed hook entries from .claude/settings.json", func(absDir string) error {
		// Re-projecting with an empty primitive set strips every
		// `statusMessage: "keystone:*"` entry and leaves user-authored
		// hooks intact. Idempotent.
		_, err := claudecode.ProjectHooks(absDir, nil)
		return err
	})

	p.Add("remove generated cross-host artifacts (AGENTS.md, CONVENTIONS.md, .aider.conf.yml, .cursor/rules, .continue/rules)", func(absDir string) error {
		for _, rel := range []string{
			"AGENTS.md",
			"CONVENTIONS.md",
			".aider.conf.yml",
			filepath.Join(".cursor", "rules"),
			filepath.Join(".continue", "rules"),
		} {
			if err := removeIfExists(filepath.Join(absDir, rel)); err != nil {
				return err
			}
		}
		return nil
	})

	return p, nil
}

// rebuildIndex writes INDEX.json + INDEX.lite.json from a fresh walk
// of the harness tree. Mirrors what `keystone index` does at the
// command layer — duplicated here so the migration is self-contained
// and replayable without spawning the binary.
func rebuildIndex(absDir, harnessRoot, indexDir string) error {
	primitives, _, err := primitive.Walk(absDir, harnessRoot)
	if err != nil {
		return err
	}
	idx := primitive.Build(primitives, time.Now())
	keystoneDir := filepath.Join(absDir, indexDir)
	if err := primitive.Write(filepath.Join(keystoneDir, config.IndexName), idx); err != nil {
		return err
	}
	return primitive.WriteLite(filepath.Join(keystoneDir, config.IndexLiteName), primitive.BuildLite(idx))
}

// runOptInAdapters fires every cross-host adapter the user has
// enabled in keystone.json. Missing keystone.json is fine — fresh
// installs land on 2.2 templates that include the adapters list.
func runOptInAdapters(absDir, harnessRoot string) error {
	cfg, err := config.ReadProjectConfig(absDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	primitives, _, err := primitive.Walk(absDir, harnessRoot)
	if err != nil {
		return err
	}
	if cfg.HasAdapter(config.AdapterCursor) {
		if _, err := cursor.ProjectRules(absDir, primitives); err != nil {
			return fmt.Errorf("cursor: %w", err)
		}
	}
	if cfg.HasAdapter(config.AdapterAider) {
		if _, err := aider.ProjectAider(absDir, agnostic.DefaultBody()); err != nil {
			return fmt.Errorf("aider: %w", err)
		}
	}
	if cfg.HasAdapter(config.AdapterContinue) {
		if _, err := continueide.ProjectRules(absDir, primitives); err != nil {
			return fmt.Errorf("continue: %w", err)
		}
	}
	return nil
}

// removeIfExists is os.RemoveAll with the not-exists error swallowed —
// makes Down idempotent.
func removeIfExists(absPath string) error {
	if err := os.RemoveAll(absPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
