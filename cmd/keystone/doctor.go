package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/loader"
	"github.com/tacoda/keystone/internal/framework/lockfile"
	"github.com/tacoda/keystone/internal/framework/policies"
	"github.com/tacoda/keystone/internal/framework/projection"
	"github.com/tacoda/keystone/internal/framework/scaffold"
)

// runDoctor handles `keystone doctor [--dir <path>] [--charter-root <name>]
// [--paths-only] [--policies-only] [--drift-only]`.
//
// Runs three independent checks against an existing install:
//
//  1. Path conformance — scan every markdown file under <charter-root>/
//     (excluding vendored policies) and flag any relative link that names
//     no file.
//  2. Policy drift — load keystone.json + lockfile, run loader.Verify;
//     drifted policies get reset and the user is told to re-run
//     `keystone install`.
//  3. Template drift — diff the user's scaffolded defaults against the
//     binary's current templates; report which user files have changed
//     so the author can decide whether to refresh.
//
// Exit codes:
//
//	0 — every check clean (or only template drift, which is informational).
//	0 — policy drift was auto-reset; install needed to repopulate.
//	1 — any path violation or strict-cascade violation.
func runDoctor(args []string) error {
	dir := "."
	runAll := true
	var runPaths, runPoliciesCheck, runDriftCheck, runBudget, fix bool

	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--help" || a == "-h":
			printDoctorUsage(os.Stdout)
			return nil
		case a == "--dir":
			if i+1 >= len(args) {
				return fmt.Errorf("flag %s requires a value", a)
			}
			dir = args[i+1]
			i++
		case strings.HasPrefix(a, "--dir="):
			dir = strings.TrimPrefix(a, "--dir=")
		case a == "--paths-only":
			runAll = false
			runPaths = true
		case a == "--policies-only":
			runAll = false
			runPoliciesCheck = true
		case a == "--drift-only":
			runAll = false
			runDriftCheck = true
		case a == "--budget":
			runAll = false
			runBudget = true
		case a == "--fix":
			fix = true
		case strings.HasPrefix(a, "-"):
			return fmt.Errorf("unknown flag %s", a)
		default:
			return fmt.Errorf("unexpected positional argument %q", a)
		}
	}
	if runAll {
		runPaths, runPoliciesCheck, runDriftCheck = true, true, true
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}
	charterRoot := config.DefaultAgentsRoot

	hadErrors := false

	if runPaths {
		if fix {
			return fmt.Errorf("--fix no longer rewrites links: 5.0 links are relative to their own file, so fix a broken one by hand")
		}
		violations, err := checkPathConventions(absDir, charterRoot)
		if err != nil {
			return fmt.Errorf("path check: %w", err)
		}
		printPathReport(violations, charterRoot)
		if len(violations) > 0 {
			hadErrors = true
		}
	}

	if runPoliciesCheck {
		errs, err := checkPolicyIntegrity(absDir, charterRoot)
		if err != nil {
			return fmt.Errorf("policy check: %w", err)
		}
		if errs > 0 {
			hadErrors = true
		}
	}

	if runDriftCheck {
		drifts, err := checkTemplateDrift(absDir, charterRoot)
		if err != nil {
			return fmt.Errorf("drift check: %w", err)
		}
		printTemplateDrift(drifts, charterRoot)
		// Template drift is informational — never sets hadErrors.
	}

	if runBudget {
		if err := runBudgetReport(absDir, charterRoot); err != nil {
			return fmt.Errorf("budget check: %w", err)
		}
		// Budget over-runs are warnings, not errors — never sets hadErrors.
	}

	if hadErrors {
		return fmt.Errorf("doctor found issues — see above")
	}
	return nil
}

// runBudgetReport prints the context budget summary. Always informational
// in doctor — an exceeded cap prints but never sets a non-zero exit;
// `keystone verify` and `keystone budget` enforce caps.
func runBudgetReport(projectDir, _ string) error {
	r, err := measureBudget(projectDir)
	if err != nil {
		return err
	}
	printBudgetSummary(os.Stdout, r)
	return nil
}

func printDoctorUsage(w *os.File) {
	fmt.Fprint(w, `keystone doctor — audit an existing charter install

Usage:
  keystone doctor [--dir <path>] [--charter-root <name>] [--paths-only|--policies-only|--drift-only]

Runs three independent checks by default:

  paths   — every relative Markdown link under .agents/ resolves to a
            file, read from the linking file's folder (code blocks are
            skipped). A broken link exits non-zero.
  policies — walks vendored policies, detects drift, resets drifted policies
            so a follow-up 'keystone install' can repopulate from cache.
  drift   — diffs the user's scaffolded defaults against the binary's
            current templates; reports which user files have changed.
            Informational only — never sets a non-zero exit.

Flags:
  --paths-only            Run only the path-convention check.
  --policies-only          Run only the policy-integrity check.
  --drift-only            Run only the template-drift check.
  --budget                Run only the per-port budget report (whitespace-
                          approximate token count vs keystone.json's
                          budgets block). Always informational — never
                          exits non-zero.
  --dir <path>            Project root (defaults to cwd).
  --charter-root <name>   Override the charter folder (defaults to the
                          value in keystone.json, then "charter").
`)
}

// --- Path conventions check ---------------------------------------------

type pathViolation struct {
	File string // path relative to project root
	Line int
	Link string // the offending link target
}

// markdownLink matches markdown link targets: [text](target) — captures
// the target. URL-style targets (http, https, mailto, etc.) are filtered
// out at violation time.
var markdownLink = regexp.MustCompile(`\[[^\]]*\]\(([^)]+)\)`)

func checkPathConventions(projectDir, charterRoot string) ([]pathViolation, error) {
	root := filepath.Join(projectDir, charterRoot)
	skip := filepath.FromSlash(config.PoliciesDir)
	var hits []pathViolation

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		rel, _ := filepath.Rel(projectDir, path)
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if strings.HasPrefix(rel, filepath.ToSlash(skip)) {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		hits = append(hits, brokenLinks(rel, filepath.Dir(path), string(body))...)
		return nil
	})
	return hits, err
}

// brokenLinks lists the links in body (a file at rel, in folder dir) that
// name no file. Code blocks are skipped.
func brokenLinks(rel, dir, body string) []pathViolation {
	var hits []pathViolation
	inFence := false
	for i, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inFence = !inFence
			continue
		}
		if !inFence {
			hits = append(hits, lineBrokenLinks(rel, dir, i+1, line)...)
		}
	}
	return hits
}

func lineBrokenLinks(rel, dir string, n int, line string) []pathViolation {
	var hits []pathViolation
	for _, m := range markdownLink.FindAllStringSubmatch(line, -1) {
		if target := strings.TrimSpace(m[1]); brokenLink(dir, target) {
			hits = append(hits, pathViolation{File: rel, Line: n, Link: target})
		}
	}
	return hits
}

// brokenLink reports a relative link target that names no file when read
// from dir. URLs, anchors, and absolute paths are not checked.
func brokenLink(dir, target string) bool {
	if !projection.IsRelativeLink(target) {
		return false
	}
	if i := strings.IndexAny(target, "#?"); i >= 0 {
		target = target[:i]
	}
	_, err := os.Stat(filepath.Join(dir, filepath.FromSlash(target)))
	return err != nil
}

func printPathReport(hits []pathViolation, charterRoot string) {
	if len(hits) == 0 {
		fmt.Fprintf(os.Stdout, "✓ paths: every relative link resolves\n")
		return
	}
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].File == hits[j].File {
			return hits[i].Line < hits[j].Line
		}
		return hits[i].File < hits[j].File
	})
	fmt.Fprintf(os.Stdout, "✗ paths: %d link(s) name a file that doesn't exist\n", len(hits))
	for _, h := range hits {
		fmt.Fprintf(os.Stdout, "    %s:%d  %s\n", h.File, h.Line, h.Link)
	}
	fmt.Fprintf(os.Stdout, "  A link is read from the linking file's folder. Projected host copies are\n")
	fmt.Fprintf(os.Stdout, "  rebased by keystone project; fix the source in .agents/.\n")
}

// --- Policy integrity check ---------------------------------------------

func checkPolicyIntegrity(projectDir, charterRoot string) (int, error) {
	cfg, err := config.ReadProjectConfig(projectDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stdout, "  no %s — skipping policy check\n", config.ProjectConfigFile)
			return 0, nil
		}
		return 0, err
	}
	if len(cfg.Policies) == 0 {
		fmt.Fprintf(os.Stdout, "✓ policies: keystone.json declares no policies — nothing to verify\n")
		return 0, nil
	}
	lf, err := lockfile.Read(projectDir)
	if err != nil {
		return 0, err
	}
	expected := map[string]map[string]string{}
	for name, lock := range lf.Policies {
		expected[name] = lock.Files
	}
	res, err := loader.Verify(projectDir, cfg, expected)
	if err != nil {
		return 0, err
	}
	errs := 0
	if res.HasDrift() {
		fmt.Fprintf(os.Stdout, "▸ policies: drift detected — resetting %d policy(s)\n", len(res.Drift))
		for _, d := range res.Drift {
			fmt.Fprintf(os.Stdout, "    • %s: %d drifted file(s)\n", d.Policy, len(d.Files))
			for _, f := range d.Files {
				fmt.Fprintf(os.Stdout, "        - %s (%s)\n", f.Path, f.Kind)
			}
			if err := policies.Reset(d.Policy, projectDir); err != nil {
				return 0, err
			}
		}
		fmt.Fprintln(os.Stdout, "  re-run `keystone install` to repopulate from cache")
	}
	if res.HasErrors() {
		errs += len(res.Violations)
		fmt.Fprintf(os.Stdout, "✗ policies: %d strict-cascade violation(s)\n", len(res.Violations))
		for _, v := range res.Violations {
			fmt.Fprintln(os.Stdout, "    "+v.String())
		}
	}
	if !res.HasDrift() && !res.HasErrors() {
		fmt.Fprintf(os.Stdout, "✓ policies: all %d policy(s) clean (no drift, no strict violations)\n", len(cfg.Policies))
	}
	return errs, nil
}

// --- Template drift check -----------------------------------------------

type templateDiff struct {
	Rel string // path under charter root
}

func checkTemplateDrift(projectDir, charterRoot string) ([]templateDiff, error) {
	root := filepath.Join(projectDir, charterRoot)
	var diffs []templateDiff

	err := fs.WalkDir(scaffold.Templates, "agents", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel := strings.TrimPrefix(path, "agents/")
		userPath := filepath.Join(root, rel)
		userBytes, err := os.ReadFile(userPath)
		if err != nil {
			if os.IsNotExist(err) {
				// User deleted a default — informational, not a drift.
				return nil
			}
			return err
		}
		templateBytes, err := fs.ReadFile(scaffold.Templates, path)
		if err != nil {
			return err
		}
		if string(userBytes) != string(templateBytes) {
			diffs = append(diffs, templateDiff{Rel: rel})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(diffs, func(i, j int) bool { return diffs[i].Rel < diffs[j].Rel })
	return diffs, nil
}

func printTemplateDrift(diffs []templateDiff, charterRoot string) {
	if len(diffs) == 0 {
		fmt.Fprintf(os.Stdout, "✓ drift: every default file matches the binary's current templates\n")
		return
	}
	fmt.Fprintf(os.Stdout, "ℹ drift: %d scaffolded file(s) diverge from the current templates\n", len(diffs))
	for _, d := range diffs {
		fmt.Fprintf(os.Stdout, "    %s/%s\n", charterRoot, d.Rel)
	}
	fmt.Fprintln(os.Stdout, "  These are either intentional project customizations or files the templates")
	fmt.Fprintln(os.Stdout, "  changed upstream. Compare with `keystone init --overwrite`")
	fmt.Fprintln(os.Stdout, "  on a copy to see what changed.")
}
