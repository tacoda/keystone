package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// runMigrate handles `keystone migrate [--dir <path>] [--dry-run]`.
//
// One-shot 1.x → 2.0 transformation. Idempotent: re-running on an
// already-2.0 install only refreshes INDEX.json + projections.
//
// Steps performed (in order):
//
//  1. Move `harness/` (1.x) to `.keystone/harness/` (2.0).
//  2. Lift the lockfile to `.keystone/lockfile.json`.
//  3. Rename vendored-policy dir `<harness>/plugins/` → `<harness>/policies/`.
//  4. Rename manifest files `keystone-plugin.json` → `keystone-policy.json`
//     inside every vendored policy.
//  5. Rewrite `keystone.json`: bump `version` to "2", rename `plugins`
//     field to `policies`, strip the deprecated `harness_root` field.
//  6. Regenerate `.keystone/INDEX.json`.
//  7. Regenerate `.claude/` host projections from canonical sources.
//
// Frontmatter on user-authored primitive files is NOT touched — the
// canonical shape ships in 2.0's templates; users mix-and-match as they
// see fit, and `keystone lint` reports anything still missing required
// fields.
func runMigrate(args []string) error {
	dir := "."
	dryRun := false
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--help" || a == "-h":
			printMigrateUsage(os.Stdout)
			return nil
		case a == "--dir":
			if i+1 >= len(args) {
				return fmt.Errorf("flag %s requires a value", a)
			}
			dir = args[i+1]
			i++
		case strings.HasPrefix(a, "--dir="):
			dir = strings.TrimPrefix(a, "--dir=")
		case a == "--dry-run":
			dryRun = true
		case strings.HasPrefix(a, "-"):
			return fmt.Errorf("unknown flag %s", a)
		default:
			return fmt.Errorf("unexpected positional argument %q", a)
		}
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}
	if info, err := os.Stat(absDir); err != nil {
		return fmt.Errorf("dir %s: %w", absDir, err)
	} else if !info.IsDir() {
		return fmt.Errorf("dir %s is not a directory", absDir)
	}

	plan, err := planMigration(absDir)
	if err != nil {
		return err
	}
	plan.print(os.Stdout)
	if dryRun {
		fmt.Fprintln(os.Stdout, "\n--dry-run: nothing changed.")
		return nil
	}
	if err := plan.execute(absDir); err != nil {
		return err
	}

	// Regenerate generated artifacts. Ignore errors so the structural
	// migration is still recorded as successful even if downstream
	// regeneration fails.
	harnessRoot := config.DefaultHarnessRoot
	primitives, _, walkErr := primitive.Walk(absDir, harnessRoot)
	if walkErr == nil {
		idx := primitive.Build(primitives, time.Now())
		outPath := filepath.Join(absDir, config.KeystoneDir(harnessRoot), config.IndexName)
		if err := primitive.Write(outPath, idx); err != nil {
			fmt.Fprintf(os.Stderr, "! keystone migrate: index write failed: %v\n", err)
		} else {
			fmt.Fprintf(os.Stdout, "  wrote: %s (%d primitive(s))\n", relTo(absDir, outPath), len(primitives))
		}
		if _, err := primitive.Project(absDir, primitives); err != nil {
			fmt.Fprintf(os.Stderr, "! keystone migrate: projection failed: %v\n", err)
		}
	}

	fmt.Fprintln(os.Stdout, "\n✓ keystone migrate complete")
	return nil
}

// migrationPlan records every step the migrator will (or did) perform.
// Each step is idempotent.
type migrationPlan struct {
	steps []migrationStep
}

type migrationStep struct {
	desc string
	op   func(absDir string) error
}

func (p *migrationPlan) add(desc string, op func(absDir string) error) {
	p.steps = append(p.steps, migrationStep{desc: desc, op: op})
}

func (p *migrationPlan) print(w io.Writer) {
	if len(p.steps) == 0 {
		fmt.Fprintln(w, "keystone migrate: already at 2.0 layout — nothing to do (INDEX.json + projections will be regenerated)")
		return
	}
	fmt.Fprintf(w, "keystone migrate — %d step(s):\n", len(p.steps))
	for i, s := range p.steps {
		fmt.Fprintf(w, "  %d. %s\n", i+1, s.desc)
	}
}

func (p *migrationPlan) execute(absDir string) error {
	for i, s := range p.steps {
		if err := s.op(absDir); err != nil {
			return fmt.Errorf("step %d (%s): %w", i+1, s.desc, err)
		}
	}
	return nil
}

func planMigration(absDir string) (*migrationPlan, error) {
	plan := &migrationPlan{}

	legacyHarness := filepath.Join(absDir, "harness")
	newHarness := filepath.Join(absDir, ".keystone", "harness")
	keystoneDir := filepath.Join(absDir, ".keystone")

	legacyExists := dirExists(legacyHarness)
	newExists := dirExists(newHarness)

	if legacyExists && newExists {
		return nil, fmt.Errorf("both legacy harness/ and .keystone/harness/ exist — resolve by hand before migrating")
	}

	if legacyExists {
		plan.add(fmt.Sprintf("move %s → %s", "harness/", ".keystone/harness/"), func(_ string) error {
			if err := os.MkdirAll(keystoneDir, 0o755); err != nil {
				return err
			}
			return os.Rename(legacyHarness, newHarness)
		})
	}

	// Lockfile lift: legacy location was <harnessRoot>/keystone.lock.json.
	// After the harness move, it's already inside .keystone/harness/. Lift
	// it one level up to .keystone/lockfile.json (2.0 location).
	plan.add("lift lockfile to .keystone/lockfile.json", func(_ string) error {
		newLockPath := filepath.Join(keystoneDir, "lockfile.json")
		// Candidates we might find — try each.
		for _, cand := range []string{
			filepath.Join(newHarness, "keystone.lock.json"),
			filepath.Join(absDir, "keystone.lock.json"), // very early 1.x layouts
		} {
			if fileExists(cand) {
				if err := os.MkdirAll(keystoneDir, 0o755); err != nil {
					return err
				}
				return os.Rename(cand, newLockPath)
			}
		}
		// Already at 2.0 location, or fresh project: nothing to do.
		return nil
	})

	// Rename vendored-policies dir.
	pluginsDir := filepath.Join(newHarness, "plugins")
	policiesDir := filepath.Join(newHarness, "policies")
	plan.add("rename .keystone/harness/plugins/ → .keystone/harness/policies/", func(_ string) error {
		if !dirExists(pluginsDir) {
			return nil
		}
		if dirExists(policiesDir) {
			return fmt.Errorf("both plugins/ and policies/ exist in .keystone/harness/ — resolve by hand")
		}
		return os.Rename(pluginsDir, policiesDir)
	})

	// Rename manifest files inside each vendored policy.
	plan.add("rename keystone-plugin.json → keystone-policy.json in vendored policies", func(_ string) error {
		if !dirExists(policiesDir) {
			return nil
		}
		return filepath.WalkDir(policiesDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if filepath.Base(path) == "keystone-plugin.json" {
				dst := filepath.Join(filepath.Dir(path), "keystone-policy.json")
				if fileExists(dst) {
					return nil
				}
				return os.Rename(path, dst)
			}
			return nil
		})
	})

	// keystone.json rewrite.
	cfgPath := filepath.Join(absDir, config.ProjectConfigFile)
	if fileExists(cfgPath) {
		plan.add("rewrite keystone.json: version 2, plugins→policies, strip harness_root", func(_ string) error {
			return rewriteKeystoneJSON(cfgPath)
		})
	}

	return plan, nil
}

// rewriteKeystoneJSON reads the project config raw, applies 1.x → 2.0
// schema transforms, and writes the result back.
//
// Transforms applied (each idempotent):
//   - version: "1" → "2"
//   - rename top-level field `plugins` to `policies`
//   - remove deprecated `harness_root`
//
// All other keys are preserved as-is.
func rewriteKeystoneJSON(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return fmt.Errorf("parse %s: %w", config.ProjectConfigFile, err)
	}
	changed := false
	if v, ok := doc["version"]; ok {
		if s, _ := v.(string); s != "2" {
			doc["version"] = "2"
			changed = true
		}
	} else {
		doc["version"] = "2"
		changed = true
	}
	if _, ok := doc["plugins"]; ok {
		if _, alreadyPolicies := doc["policies"]; alreadyPolicies {
			delete(doc, "plugins") // ambiguous — favor the new name
		} else {
			doc["policies"] = doc["plugins"]
			delete(doc, "plugins")
		}
		changed = true
	}
	if _, ok := doc["harness_root"]; ok {
		delete(doc, "harness_root")
		changed = true
	}
	if !changed {
		return nil
	}
	// Re-encode with stable field order matching ProjectConfig.
	ordered := map[string]any{}
	for _, k := range []string{"version", "framework_version", "policies", "budgets"} {
		if v, ok := doc[k]; ok {
			ordered[k] = v
			delete(doc, k)
		}
	}
	for k, v := range doc {
		ordered[k] = v
	}
	out, err := json.MarshalIndent(ordered, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')
	return os.WriteFile(path, out, 0o644)
}

func dirExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

func relTo(base, target string) string {
	if rel, err := filepath.Rel(base, target); err == nil {
		return rel
	}
	return target
}

func printMigrateUsage(w *os.File) {
	fmt.Fprint(w, `keystone migrate — one-shot 1.x → 2.0 upgrade

Usage:
  keystone migrate [--dir <path>] [--dry-run]

Performs every structural change required to lift a 1.x install onto
the 2.0 layout. Idempotent: safe to re-run; a second run only
regenerates INDEX.json + .claude/ projections.

Steps:
  1. Move harness/ → .keystone/harness/
  2. Lift the lockfile to .keystone/lockfile.json
  3. Rename .keystone/harness/plugins/ → .keystone/harness/policies/
  4. Rename keystone-plugin.json → keystone-policy.json in each vendored policy
  5. Rewrite keystone.json (version 2, plugins→policies, drop harness_root)
  6. Regenerate .keystone/INDEX.json
  7. Regenerate .claude/ host projections

User-authored content (rule body text, action playbooks, etc.) is
never edited. Frontmatter on legacy files is preserved as-is — run
`+"`keystone lint`"+` afterward to see what still needs canonical
frontmatter.

Flags:
  --dir <path>    Project root (defaults to cwd).
  --dry-run       Print the plan; don't apply it.
`)
}
