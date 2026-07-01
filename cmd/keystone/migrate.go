package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/lockfile"
	"github.com/tacoda/keystone/internal/framework/migrations"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// runMigrate dispatches `keystone migrate <subcommand>`:
//
//	keystone migrate                   → up to latest
//	keystone migrate up [<version>]    → up to <version> (default: latest)
//	keystone migrate down [<version>]  → down to <version> (default: previous)
//	keystone migrate status            → show applied + pending
//
// Common flags: --dir <path>, --dry-run.
//
// Every direction obeys the iron laws in package migrations: core
// framework files only, never user-edited content, every change shown,
// every problem surfaced. On a clean run, the lockfile's
// migrations_applied slice is updated to reflect the new state.
func runMigrate(args []string) error {
	if len(args) > 0 {
		switch args[0] {
		case "--help", "-h":
			printMigrateUsage(os.Stdout)
			return nil
		case "up":
			return runMigrateUp(args[1:])
		case "down":
			return runMigrateDown(args[1:])
		case "status":
			return runMigrateStatus(args[1:])
		case "-", "--":
			return fmt.Errorf("unknown subcommand %q", args[0])
		}
		if strings.HasPrefix(args[0], "-") {
			// no subcommand; treat all args as up's args
			return runMigrateUp(args)
		}
		return fmt.Errorf("unknown subcommand %q (expected: up | down | status)", args[0])
	}
	return runMigrateUp(nil)
}

type migrateFlags struct {
	dir    string
	dryRun bool
	target string // explicit version, or "" for default
}

func parseMigrateFlags(args []string) (*migrateFlags, error) {
	f := &migrateFlags{dir: "."}
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--help" || a == "-h":
			return nil, errHelp
		case a == "--dir":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("flag %s requires a value", a)
			}
			f.dir = args[i+1]
			i++
		case strings.HasPrefix(a, "--dir="):
			f.dir = strings.TrimPrefix(a, "--dir=")
		case a == "--dry-run":
			f.dryRun = true
		case strings.HasPrefix(a, "-"):
			return nil, fmt.Errorf("unknown flag %s", a)
		default:
			if f.target != "" {
				return nil, fmt.Errorf("unexpected positional argument %q", a)
			}
			f.target = a
		}
	}
	return f, nil
}

var errHelp = fmt.Errorf("help requested")

func runMigrateUp(args []string) error {
	f, err := parseMigrateFlags(args)
	if err != nil {
		if err == errHelp {
			printMigrateUsage(os.Stdout)
			return nil
		}
		return err
	}
	absDir, err := resolveDir(f.dir)
	if err != nil {
		return err
	}

	applied := readAppliedTolerant(absDir)
	pending := migrations.Pending(applied)
	if f.target != "" {
		// Filter pending to those at-or-below the target version.
		filtered := pending[:0]
		for _, m := range pending {
			if migrations.CompareVersion(m.Version, f.target) <= 0 {
				filtered = append(filtered, m)
			}
		}
		pending = filtered
	}

	if len(pending) == 0 {
		fmt.Fprintln(os.Stdout, "✓ keystone migrate up: nothing to do (already at target)")
		return nil
	}

	fmt.Fprintf(os.Stdout, "keystone migrate up — %d migration(s) pending:\n", len(pending))
	for _, m := range pending {
		fmt.Fprintf(os.Stdout, "  • %s\n", m.Version)
	}
	fmt.Fprintln(os.Stdout)

	for _, m := range pending {
		if err := runMigrationDirection(absDir, m, "up", f.dryRun); err != nil {
			return fmt.Errorf("up %s: %w", m.Version, err)
		}
		if !f.dryRun {
			applied = append(applied, m.Version)
			if err := writeApplied(absDir, applied); err != nil {
				return fmt.Errorf("record %s applied: %w", m.Version, err)
			}
		}
	}

	if !f.dryRun {
		regenerateGenerated(absDir)
	}
	fmt.Fprintln(os.Stdout, "\n✓ keystone migrate up complete")
	return nil
}

func runMigrateDown(args []string) error {
	f, err := parseMigrateFlags(args)
	if err != nil {
		if err == errHelp {
			printMigrateUsage(os.Stdout)
			return nil
		}
		return err
	}
	absDir, err := resolveDir(f.dir)
	if err != nil {
		return err
	}

	applied := readAppliedTolerant(absDir)
	if len(applied) == 0 {
		fmt.Fprintln(os.Stdout, "✓ keystone migrate down: nothing to revert (no migrations recorded)")
		return nil
	}

	// Determine which migrations to revert. If target is empty, revert one.
	// Otherwise, revert every migration with version > target.
	var toRevert []migrations.Migration
	if f.target == "" {
		latest := applied[len(applied)-1]
		m := migrations.Find(latest)
		if m == nil {
			return fmt.Errorf("applied migration %s not found in registry — cannot revert without its Down plan", latest)
		}
		toRevert = []migrations.Migration{*m}
	} else {
		for i := len(applied) - 1; i >= 0; i-- {
			v := applied[i]
			if migrations.CompareVersion(v, f.target) <= 0 {
				break
			}
			m := migrations.Find(v)
			if m == nil {
				return fmt.Errorf("applied migration %s not found in registry — cannot revert without its Down plan", v)
			}
			toRevert = append(toRevert, *m)
		}
	}
	if len(toRevert) == 0 {
		fmt.Fprintln(os.Stdout, "✓ keystone migrate down: nothing to revert (already at target)")
		return nil
	}

	fmt.Fprintf(os.Stdout, "keystone migrate down — %d migration(s) to revert (newest first):\n", len(toRevert))
	for _, m := range toRevert {
		fmt.Fprintf(os.Stdout, "  • %s\n", m.Version)
	}
	fmt.Fprintln(os.Stdout)

	for _, m := range toRevert {
		if err := runMigrationDirection(absDir, m, "down", f.dryRun); err != nil {
			return fmt.Errorf("down %s: %w", m.Version, err)
		}
		if !f.dryRun {
			applied = applied[:len(applied)-1]
			if err := writeApplied(absDir, applied); err != nil {
				return fmt.Errorf("record %s reverted: %w", m.Version, err)
			}
		}
	}

	if !f.dryRun {
		regenerateGenerated(absDir)
	}
	fmt.Fprintln(os.Stdout, "\n✓ keystone migrate down complete")
	return nil
}

func runMigrateStatus(args []string) error {
	f, err := parseMigrateFlags(args)
	if err != nil {
		if err == errHelp {
			printMigrateUsage(os.Stdout)
			return nil
		}
		return err
	}
	absDir, err := resolveDir(f.dir)
	if err != nil {
		return err
	}

	applied := readAppliedTolerant(absDir)
	pending := migrations.Pending(applied)

	current := "(none)"
	if len(applied) > 0 {
		current = applied[len(applied)-1]
	}
	fmt.Fprintf(os.Stdout, "keystone migrate status\n")
	fmt.Fprintf(os.Stdout, "  current:  %s\n", current)
	if len(applied) > 0 {
		fmt.Fprintf(os.Stdout, "  applied:  %s\n", strings.Join(applied, ", "))
	} else {
		fmt.Fprintln(os.Stdout, "  applied:  (none)")
	}
	if len(pending) == 0 {
		fmt.Fprintln(os.Stdout, "  pending:  (none — up to date)")
	} else {
		var versions []string
		for _, m := range pending {
			versions = append(versions, m.Version)
		}
		fmt.Fprintf(os.Stdout, "  pending:  %s\n", strings.Join(versions, ", "))
	}
	return nil
}

func runMigrationDirection(absDir string, m migrations.Migration, dir string, dryRun bool) error {
	var (
		plan *migrations.Plan
		err  error
	)
	switch dir {
	case "up":
		plan, err = m.Up(absDir)
	case "down":
		plan, err = m.Down(absDir)
	default:
		return fmt.Errorf("invalid direction %q", dir)
	}
	if err != nil {
		return err
	}
	printPlan(os.Stdout, m.Version, dir, plan)
	if dryRun {
		fmt.Fprintln(os.Stdout, "  (dry-run: no changes written)")
		return nil
	}
	return plan.Execute(absDir)
}

func printPlan(w io.Writer, version, dir string, p *migrations.Plan) {
	if len(p.Steps) == 0 {
		fmt.Fprintf(w, "═══ %s %s ═══  (no-op)\n", version, dir)
		return
	}
	fmt.Fprintf(w, "═══ %s %s ═══\n", version, dir)
	for i, s := range p.Steps {
		fmt.Fprintf(w, "  %d. %s\n", i+1, s.Desc)
	}
}

// readAppliedTolerant returns the applied-migrations list from the
// lockfile, tolerating every historical location: 4.0 (.charter/
// lockfile.json), 3.0 (.harness/lockfile.json), 2.0–2.4 (.keystone/
// lockfile.json), and pre-2.0 (harness/keystone.lock.json). On any read
// error the list comes back empty.
//
// Filesystem-derived fallback: when no lockfile entry is found, the
// on-disk layout is consulted. A .keystone/ tree without any recorded
// migrations is treated as an implicit "2.0 applied" so those installs
// don't show a spurious 2.0-pending warning. A bare harness/ at the
// project root leaves applied empty — that IS a pre-2.0 install that
// hasn't been migrated yet.
func readAppliedTolerant(absDir string) []string {
	candidates := []string{
		// 4.0+ canonical: the lockfile lives at <root>/lockfile.json.
		filepath.Join(absDir, config.DefaultCharterRoot, config.LockfileName),
		// 3.0 legacy: the single root was .harness/.
		filepath.Join(absDir, ".harness", "lockfile.json"),
		// 2.0–2.4 legacy: under .keystone/.
		filepath.Join(absDir, ".keystone", "lockfile.json"),
		// pre-2.0: bare harness/ dir with the old lockfile name.
		filepath.Join(absDir, config.DefaultCharterRoot, "keystone.lock.json"),
		filepath.Join(absDir, "harness", "keystone.lock.json"),
	}
	for _, path := range candidates {
		if applied, ok := readAppliedAt(path); ok {
			if len(applied) > 0 {
				return applied
			}
			break
		}
	}
	if dirExists(filepath.Join(absDir, ".keystone", "harness")) {
		return []string{"2.0"}
	}
	return nil
}

func readAppliedAt(path string) ([]string, bool) {
	if _, err := os.Stat(path); err != nil {
		return nil, false
	}
	lf, err := lockfile.ReadFromPath(path)
	if err != nil {
		return nil, false
	}
	return lf.MigrationsApplied, true
}

func writeApplied(absDir string, applied []string) error {
	// Write to the lockfile at its current location. Post-4.0 the charter
	// (and its lockfile) lives at <root>/.charter/; a pre-3.0 install still
	// has it under .keystone/. The 3.0/4.0 Up relocates it before this runs.
	path := filepath.Join(absDir, config.DefaultCharterRoot, config.LockfileName)
	if !dirExists(filepath.Join(absDir, config.DefaultCharterRoot)) {
		path = filepath.Join(absDir, ".keystone", "lockfile.json")
	}
	lf, err := lockfile.ReadFromPath(path)
	if err != nil {
		// Lockfile may not exist yet (fresh install). Build a minimal one.
		lf = &lockfile.Lockfile{Version: lockfile.Version}
	}
	lf.MigrationsApplied = applied
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return lockfile.WriteToPath(path, lf)
}

func resolveDir(dir string) (string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolve dir: %w", err)
	}
	if info, err := os.Stat(absDir); err != nil {
		return "", fmt.Errorf("dir %s: %w", absDir, err)
	} else if !info.IsDir() {
		return "", fmt.Errorf("dir %s is not a directory", absDir)
	}
	return absDir, nil
}

// dirExists / fileExists are also used by other commands in this package
// (mcp.go); kept here so the migrate refactor doesn't strand them.
func dirExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

// regenerateGenerated refreshes INDEX.json + .claude/ projections after
// a successful Up or Down run. Errors are reported but not fatal — the
// structural migration has already landed and is the load-bearing change.
func regenerateGenerated(absDir string) {
	charterRoot := config.DefaultCharterRoot
	primitives, _, walkErr := primitive.Walk(absDir, charterRoot)
	if walkErr != nil {
		return
	}
	idx := primitive.Build(primitives, time.Now())
	outPath := filepath.Join(absDir, config.KeystoneDir(charterRoot), config.IndexName)
	if err := primitive.Write(outPath, idx); err != nil {
		fmt.Fprintf(os.Stderr, "! keystone migrate: index write failed: %v\n", err)
	} else {
		fmt.Fprintf(os.Stdout, "  wrote: %s (%d primitive(s))\n", relTo(absDir, outPath), len(primitives))
	}
	if _, err := primitive.Project(absDir, primitives); err != nil {
		fmt.Fprintf(os.Stderr, "! keystone migrate: projection failed: %v\n", err)
	}
}

func relTo(base, target string) string {
	if rel, err := filepath.Rel(base, target); err == nil {
		return rel
	}
	return target
}

func printMigrateUsage(w *os.File) {
	fmt.Fprint(w, `keystone migrate — versioned forward + backward transforms

Usage:
  keystone migrate [up [<version>] | down [<version>] | status]
                   [--dir <path>] [--dry-run]

Subcommands:
  up [<version>]    Apply every pending migration up to <version>
                    (default: latest). Records applied migrations in
                    .charter/lockfile.json.
  down [<version>]  Revert migrations newer than <version> (default:
                    one step). Pops entries off the applied list as
                    each Down plan succeeds.
  status            Show current version, applied list, and pending
                    migrations.

With no subcommand, runs `+"`up`"+` to the latest version.

Iron laws every migration obeys:
  • Edits framework-owned files only (keystone.json schema fields, the
    lockfile, vendored policy manifests, generated INDEX.json /
    .claude projections, scaffolded charter directory layout).
  • Never edits user-authored primitive content. Frontmatter changes
    surface through `+"`keystone lint`"+` for the user to apply.
  • Renames folders and files freely; rewrites the contents of
    framework-owned files only.
  • Prints every step before execution. --dry-run prints without
    writing.
  • Stops and reports any unexpected state (conflict, missing
    prerequisite, ambiguous schema) for the user to resolve by hand.

Flags:
  --dir <path>      Project root (defaults to cwd).
  --dry-run         Print the plan; don't apply it.
`)
}
