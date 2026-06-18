// Package migrations is the registry of versioned forward/backward
// transforms keystone applies to a project on `keystone migrate`. Each
// Migration is one atomic step keyed by a SemVer-ish version string;
// every Migration owns both an Up (apply) and a Down (revert) Plan so
// upgrades can be rolled back when something fails downstream.
//
// State of which Migrations have been applied lives in
// .keystone/lockfile.json (field migrations_applied) — the CLI dispatch
// reads + writes that field after every transition.
//
// # Iron laws
//
// Every migration step authored in this package MUST obey the
// following invariants. They exist so users can trust `keystone migrate`
// will never silently mutate work they authored.
//
//  1. Core-framework only. Migrations edit framework-owned files:
//     keystone.json (schema fields), the lockfile, vendored policy
//     manifests, generated INDEX.json / .claude projections, and the
//     scaffolded harness directory layout. They do not edit any other
//     file the framework didn't ship.
//  2. Never user-edited files. The body of a user-authored primitive
//     (guides/, corpus/, sensors/, actions/, etc. — any file the user
//     can hand-edit) is off-limits. Frontmatter inside those files is
//     equally off-limits; if a schema change makes old frontmatter
//     wrong, surface it through `keystone lint` for the user to fix,
//     don't auto-rewrite.
//  3. Folder and file renames are allowed. Moving a directory or
//     renaming a file does not change the file's content; this is the
//     primary mechanism a migration uses to lift legacy layouts onto
//     the current one.
//  4. Every change is shown. Before execution, the plan prints every
//     step it will run. --dry-run prints without writing. After
//     execution, the per-step result is reported. No step mutates state
//     without leaving a visible trace.
//  5. Problems surface. Unexpected state (both legacy AND target paths
//     exist, missing prerequisite, ambiguous schema) returns an error
//     that names the conflict and asks the user to resolve by hand.
//     Migrations never auto-recover; they stop and report.
//  6. No breaking changes between properly migrated versions. Once a
//     consumer has run every Up in the registry, every keystone command
//     on that version must work without further user intervention. A
//     consumer who has NOT migrated to the latest version sees a soft
//     warning on every command run ("N pending migration(s) — run
//     `keystone migrate up`") but commands still proceed; the
//     framework's runtime readers (config, lockfile, etc.) keep
//     backward-compat fallbacks for the previous version's schema so
//     unmigrated installs degrade gracefully rather than crash.
package migrations

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Migration is one versioned transform with paired Up/Down plans.
//
// Up runs when migrating forward to this version; Down reverts to the
// version immediately below it in the registry. Plans are computed
// against the on-disk state of absDir at call time, not pre-cached, so a
// dry-run shows what would happen now.
type Migration struct {
	Version string
	Up      func(absDir string) (*Plan, error)
	Down    func(absDir string) (*Plan, error)
}

// Plan is the ordered list of steps a single migration direction
// performs. Steps are independent in failure terms — if one fails, the
// whole plan stops; partially-applied state must be reconcilable by
// re-running.
type Plan struct {
	Steps []Step
}

// Step is one filesystem-mutating operation inside a Plan. Desc is what
// the CLI prints; Op is the closure that does the work. Op must be
// idempotent — repeating it on already-converged state is a no-op.
type Step struct {
	Desc string
	Op   func(absDir string) error
}

// Add appends a step to a plan. Used by migration authors when building
// the Up/Down plans.
func (p *Plan) Add(desc string, op func(absDir string) error) {
	p.Steps = append(p.Steps, Step{Desc: desc, Op: op})
}

// Execute runs every step in order. The first error short-circuits and
// is wrapped with the step index + description for diagnosis.
func (p *Plan) Execute(absDir string) error {
	for i, s := range p.Steps {
		if err := s.Op(absDir); err != nil {
			return fmt.Errorf("step %d (%s): %w", i+1, s.Desc, err)
		}
	}
	return nil
}

var registry []Migration

// Register adds a migration to the package-level registry. Called from
// per-version init() funcs in this package. Re-sorts by ascending
// version so callers always see the registry in chronological order.
func Register(m Migration) {
	registry = append(registry, m)
	sort.Slice(registry, func(i, j int) bool {
		return CompareVersion(registry[i].Version, registry[j].Version) < 0
	})
}

// All returns the registry sorted ascending by version. Safe to mutate
// the returned slice; the underlying registry stays intact.
func All() []Migration {
	out := make([]Migration, len(registry))
	copy(out, registry)
	return out
}

// Find returns the migration for the given version, or nil if it isn't
// registered.
func Find(version string) *Migration {
	for i := range registry {
		if registry[i].Version == version {
			return &registry[i]
		}
	}
	return nil
}

// Pending returns every registered migration with a version strictly
// greater than the most recent entry in applied. When applied is empty,
// every registered migration is pending.
func Pending(applied []string) []Migration {
	current := ""
	if len(applied) > 0 {
		current = applied[len(applied)-1]
	}
	var out []Migration
	for _, m := range All() {
		if current == "" || CompareVersion(m.Version, current) > 0 {
			out = append(out, m)
		}
	}
	return out
}

// CompareVersion returns negative if a < b, zero if equal, positive if
// a > b. Compares dotted numeric segments left-to-right; missing
// segments are treated as zero. Non-numeric segments compare as zero.
//
// "2.0" < "2.0.1" < "2.1" < "2.10" < "3.0".
func CompareVersion(a, b string) int {
	as := strings.Split(a, ".")
	bs := strings.Split(b, ".")
	n := len(as)
	if len(bs) > n {
		n = len(bs)
	}
	for i := 0; i < n; i++ {
		ai := segInt(as, i)
		bi := segInt(bs, i)
		if ai != bi {
			return ai - bi
		}
	}
	return 0
}

func segInt(parts []string, i int) int {
	if i >= len(parts) {
		return 0
	}
	n, _ := strconv.Atoi(parts[i])
	return n
}
