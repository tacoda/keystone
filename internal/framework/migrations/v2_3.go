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

// Version 2.3 — framework abstractions level up. No data migration
// required; the contract is purely additive:
//
//   - New primitive kind: `concern`. Lives under
//     `.keystone/harness/concerns/`. Reusable frontmatter + body
//     fragments composed into other primitives via `includes:`.
//   - `Includes []string` field on every primitive. Walker resolves
//     after parse: list fields union (deduped, host's values last),
//     scalar fields host-wins, concerns are leaves (no nested
//     includes).
//   - `Tags []string` field on every primitive. Orthogonal taxonomy
//     surfaced by the new `keystone list --tag <tag>` command.
//   - `keystone list [<kind>] [--tag <tag>]...` — filtered primitive
//     listing with composed view.
//   - `keystone show <kind> <id>` — single-primitive descriptor +
//     forward & reverse association lookup (includes / included_by,
//     traces / traced_by, host hooks). `--json` for structured
//     output.
//   - Sensor `severity:` now wires through to host hook behavior in
//     the claudecode adapter:
//       must (default) → command runs as-is (exit 2 blocks)
//       should         → wrapper converts non-zero exit to 0 + stderr warning
//       may            → wrapper converts non-zero exit to 0 silently
//   - claudecode `.claude/settings.json` JSON marshal disables HTML
//     escaping so shell metacharacters round-trip cleanly inside
//     hook command strings.
//
// All additions are optional fields and new files. Existing 2.2
// content remains valid. This migration's Up plan replays the
// projection pipeline once so the `.claude/`, `.cursor/`,
// `AGENTS.md`, etc. surfaces pick up the new fields.

func init() {
	Register(Migration{
		Version: "2.3",
		Up:      planUp_2_3,
		Down:    planDown_2_3,
	})
}

func planUp_2_3(absDir string) (*Plan, error) {
	p := &Plan{}

	harnessRoot := legacyHarnessRoot
	harnessAbs := filepath.Join(absDir, harnessRoot)
	if !dirExists(harnessAbs) {
		// Fresh install — `keystone init` scaffolds from 2.3 templates.
		return p, nil
	}

	// Ensure the concerns directory exists. It's optional content (a
	// project may have zero concerns) but the canonical path needs to
	// be there so a future `keystone new concern` finds its home.
	concernsDir := filepath.Join(harnessAbs, "concerns")
	p.Add("ensure .keystone/harness/concerns/ exists", func(_ string) error {
		return os.MkdirAll(concernsDir, 0o755)
	})

	p.Add("re-walk + re-compose + rebuild INDEX.json + INDEX.lite.json", func(absDir string) error {
		return rebuildIndexComposed(absDir, harnessRoot, legacyKeystoneDir)
	})

	p.Add("re-project primitives → .claude/{agents,commands,skills,rules}", func(absDir string) error {
		primitives, _, err := primitive.Walk(absDir, harnessRoot)
		if err != nil {
			return err
		}
		composed, _ := primitive.Compose(primitives)
		_, err = primitive.Project(absDir, composed)
		return err
	})

	p.Add("project agnostic AGENTS.md at repo root", func(absDir string) error {
		_, err := agnostic.ProjectAgentsMD(absDir, agnostic.DefaultBody())
		return err
	})

	p.Add("merge sensor host_triggers into .claude/settings.json with severity wrappers", func(absDir string) error {
		primitives, _, err := primitive.Walk(absDir, harnessRoot)
		if err != nil {
			return err
		}
		composed, _ := primitive.Compose(primitives)
		_, err = claudecode.ProjectHooks(absDir, composed)
		return err
	})

	p.Add("project opt-in adapters from keystone.json (cursor / aider / continue)", func(absDir string) error {
		return runOptInAdaptersComposed(absDir, harnessRoot)
	})

	return p, nil
}

func planDown_2_3(absDir string) (*Plan, error) {
	p := &Plan{}

	// 2.3 is additive only. Down is best-effort: leave the concerns
	// dir in place (deleting it could lose user-authored content),
	// but re-project the pipeline so .claude/settings.json drops the
	// severity-wrap behavior. 2.2's projection rewrites every
	// keystone-managed entry with unwrapped commands.
	p.Add("re-walk WITHOUT compose, rebuild INDEX (drops merged tags / includes effects)", func(absDir string) error {
		primitives, _, err := primitive.Walk(absDir, legacyHarnessRoot)
		if err != nil {
			return err
		}
		idx := primitive.Build(primitives, time.Now())
		keystoneDir := filepath.Join(absDir, legacyKeystoneDir)
		if err := primitive.Write(filepath.Join(keystoneDir, config.IndexName), idx); err != nil {
			return err
		}
		return primitive.WriteLite(filepath.Join(keystoneDir, config.IndexLiteName), primitive.BuildLite(idx))
	})

	p.Add("re-project hooks without severity wrappers (uncomposed primitives)", func(absDir string) error {
		primitives, _, err := primitive.Walk(absDir, legacyHarnessRoot)
		if err != nil {
			return err
		}
		// Skip Compose — the 2.2 contract didn't merge concern fields.
		// Sensors that gained `severity:` in 2.3 still parse fine; the
		// adapter just won't read it.
		_, err = claudecode.ProjectHooks(absDir, primitives)
		return err
	})

	return p, nil
}

// rebuildIndexComposed mirrors v2_2's rebuildIndex but routes the
// primitive slice through Compose before indexing, so the descriptor
// records what the agent actually sees after `includes:` resolution.
func rebuildIndexComposed(absDir, harnessRoot, indexDir string) error {
	primitives, _, err := primitive.Walk(absDir, harnessRoot)
	if err != nil {
		return err
	}
	composed, _ := primitive.Compose(primitives)
	idx := primitive.Build(composed, time.Now())
	keystoneDir := filepath.Join(absDir, indexDir)
	if err := primitive.Write(filepath.Join(keystoneDir, config.IndexName), idx); err != nil {
		return err
	}
	return primitive.WriteLite(filepath.Join(keystoneDir, config.IndexLiteName), primitive.BuildLite(idx))
}

// runOptInAdaptersComposed mirrors v2_2's helper but composes the
// primitive slice before handing it to each adapter — so the cursor /
// continue rule shims see merged globs + tags, and the aider
// CONVENTIONS.md reflects the current `agnostic.DefaultBody()`.
func runOptInAdaptersComposed(absDir, harnessRoot string) error {
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
	composed, _ := primitive.Compose(primitives)
	if cfg.HasAdapter(config.AdapterCursor) {
		if _, err := cursor.ProjectRules(absDir, composed); err != nil {
			return fmt.Errorf("cursor: %w", err)
		}
	}
	if cfg.HasAdapter(config.AdapterAider) {
		if _, err := aider.ProjectAider(absDir, agnostic.DefaultBody()); err != nil {
			return fmt.Errorf("aider: %w", err)
		}
	}
	if cfg.HasAdapter(config.AdapterContinue) {
		if _, err := continueide.ProjectRules(absDir, composed); err != nil {
			return fmt.Errorf("continue: %w", err)
		}
	}
	return nil
}
