package main

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/tacoda/keystone/internal/framework/budget"
	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/loader"
	"github.com/tacoda/keystone/internal/framework/lockfile"
	"github.com/tacoda/keystone/internal/framework/policies"
	"github.com/tacoda/keystone/internal/framework/scaffold"
)

// runDoctor handles `keystone doctor [--dir <path>] [--charter-root <name>]
// [--paths-only] [--policies-only] [--drift-only]`.
//
// Runs three independent checks against an existing install:
//
//  1. Path conformance — scan every markdown file under <charter-root>/
//     (excluding vendored policies) and flag any inter-charter link with
//     '../' or './' segments. See docs/conventions.md for the rule.
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
	charterRoot := config.DefaultCharterRoot

	hadErrors := false

	if runPaths {
		if fix {
			fixed, err := fixPathConventions(absDir, charterRoot)
			if err != nil {
				return fmt.Errorf("path fix: %w", err)
			}
			fmt.Fprintf(os.Stdout, "✓ paths: rewrote %d link(s) to charter-root-relative form\n", fixed)
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

// runBudgetReport walks the charter, estimates per-file token use, and
// renders the per-port breakdown vs the budgets declared in keystone.json.
// Always informational — over-budget ports print a warning but never
// set a non-zero exit. Projects that want stricter enforcement can wrap
// `keystone doctor --budget` in a script that greps for the warning marker.
func runBudgetReport(projectDir, charterRoot string) error {
	alloc, err := walkCharterBudget(projectDir, charterRoot)
	if err != nil {
		return err
	}
	cfg, _ := config.ReadProjectConfig(projectDir) // nil cfg = no budgets

	reps := alloc.Report(cfg, 5)
	if len(reps) == 0 {
		fmt.Fprintf(os.Stdout, "  no markdown content under %s/ — nothing to count\n", charterRoot)
		return nil
	}

	overBudget := 0
	fmt.Fprintln(os.Stdout, "  budget: per-port token estimate (whitespace-approximate)")
	for _, r := range reps {
		marker := "•"
		budgetCol := "no cap"
		if r.MaxTokens > 0 {
			budgetCol = fmt.Sprintf("%d / %d", r.Tokens, r.MaxTokens)
			if r.IsOverBudget() {
				marker = "!"
				overBudget++
			} else {
				budgetCol += fmt.Sprintf("  (%d%% used)", 100*r.Tokens/r.MaxTokens)
			}
		} else {
			budgetCol = fmt.Sprintf("%d tokens", r.Tokens)
		}
		fmt.Fprintf(os.Stdout, "    %s %-10s %s\n", marker, r.Port, budgetCol)
		for _, f := range r.TopFiles {
			fmt.Fprintf(os.Stdout, "        %5d  %s\n", f.Tokens, f.Path)
		}
	}
	if overBudget > 0 {
		fmt.Fprintf(os.Stdout, "  ⚠ %d port(s) over their declared budget — top contributors above\n", overBudget)
	} else if anyBudgetDeclared(cfg) {
		fmt.Fprintln(os.Stdout, "  ✓ every port with a declared budget is within cap")
	} else {
		fmt.Fprintln(os.Stdout, "  ℹ no budgets declared in keystone.json — add a `budgets` block to enforce caps")
	}
	return nil
}

// walkCharterBudget scans every .md file under <charterRoot>/, classifies
// it by port (skipping non-port paths like README, learning/, archive/),
// estimates its tokens, and records the result in an Allocator.
func walkCharterBudget(projectDir, charterRoot string) (*budget.Allocator, error) {
	alloc := budget.NewAllocator()
	root := filepath.Join(projectDir, charterRoot)
	skip := filepath.Join(charterRoot, policies.PolicyRoot)

	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		rel, _ := filepath.Rel(projectDir, p)
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if strings.HasPrefix(rel, filepath.ToSlash(skip)) {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(p) != ".md" {
			return nil
		}
		port := budget.PortForPath(rel, charterRoot)
		if port == "" {
			return nil
		}
		body, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		alloc.Add(port, rel, budget.Estimate(body))
		return nil
	})
	return alloc, err
}

// anyBudgetDeclared reports whether cfg.Budgets has any non-zero
// MaxTokens entry — used to decide between the "no budgets declared"
// and "everything within cap" closing messages.
func anyBudgetDeclared(cfg *config.ProjectConfig) bool {
	if cfg == nil {
		return false
	}
	for _, spec := range cfg.Budgets {
		if spec.MaxTokens > 0 || spec.MaxTokensPerLoad > 0 {
			return true
		}
	}
	return false
}

func printDoctorUsage(w *os.File) {
	fmt.Fprint(w, `keystone doctor — audit an existing charter install

Usage:
  keystone doctor [--dir <path>] [--charter-root <name>] [--paths-only|--policies-only|--drift-only]

Runs three independent checks by default:

  paths   — every inter-charter link must be charter-root-relative;
            '../' and './' segments are forbidden (see docs/conventions.md
            for the rule). Violations exit non-zero.
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
  --fix                   Rewrite path violations in place (paths check only).
                          Each '../' or './' link is resolved against the
                          source file's directory and replaced with the
                          charter-root-relative form.
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
	skip := filepath.Join(charterRoot, policies.PolicyRoot)
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
		for i, line := range strings.Split(string(body), "\n") {
			for _, m := range markdownLink.FindAllStringSubmatch(line, -1) {
				target := strings.TrimSpace(m[1])
				if hasForbiddenSegment(target) {
					hits = append(hits, pathViolation{File: rel, Line: i + 1, Link: target})
				}
			}
		}
		return nil
	})
	return hits, err
}

// fixPathConventions walks every markdown file under <charter-root>/
// (excluding vendored policies) and rewrites markdown links with '../'
// or './' segments to their charter-root-relative form. Returns the
// total count of rewritten links.
//
// Resolution: for a link `target` in a file at `<charterRoot>/<rel>`,
// the charter-root-relative form is path.Join(dir(rel), target). Go's
// path.Join normalizes the '..' and '.' segments correctly.
func fixPathConventions(projectDir, charterRoot string) (int, error) {
	root := filepath.Join(projectDir, charterRoot)
	skip := filepath.Join(charterRoot, policies.PolicyRoot)
	count := 0

	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		rel, _ := filepath.Rel(projectDir, p)
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if strings.HasPrefix(rel, filepath.ToSlash(skip)) {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(p) != ".md" {
			return nil
		}
		body, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		// Path of this file's directory relative to charter root.
		insideCharter := strings.TrimPrefix(rel, charterRoot+"/")
		fileDir := path.Dir(insideCharter)
		if fileDir == "." {
			fileDir = ""
		}

		fileChanged := false
		newBody := markdownLink.ReplaceAllStringFunc(string(body), func(match string) string {
			submatches := markdownLink.FindStringSubmatch(match)
			if len(submatches) < 2 {
				return match
			}
			target := strings.TrimSpace(submatches[1])
			if !hasForbiddenSegment(target) {
				return match
			}
			// Preserve any #anchor or ?query suffix verbatim.
			suffix := ""
			if i := strings.IndexAny(target, "#?"); i >= 0 {
				suffix = target[i:]
				target = target[:i]
			}
			resolved := path.Join(fileDir, target)
			fileChanged = true
			count++
			// Splice the new target back into the original [text](target)
			// match, preserving the leading [...]( and trailing ).
			openParen := strings.IndexByte(match, '(')
			closeParen := strings.LastIndexByte(match, ')')
			if openParen < 0 || closeParen < 0 {
				return match
			}
			return match[:openParen+1] + resolved + suffix + match[closeParen:]
		})
		if fileChanged {
			if err := os.WriteFile(p, []byte(newBody), 0o644); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}

// hasForbiddenSegment returns true for inter-charter link targets that
// contain '../' or './' segments. URL-style targets are exempt.
func hasForbiddenSegment(target string) bool {
	if target == "" {
		return false
	}
	if strings.Contains(target, "://") {
		return false // url-style — out of scope
	}
	if strings.HasPrefix(target, "#") || strings.HasPrefix(target, "mailto:") {
		return false
	}
	// Strip url-style fragment + query suffix so '../X.md#anchor' still
	// flags as a violation.
	if i := strings.IndexAny(target, "#?"); i >= 0 {
		target = target[:i]
	}
	for _, seg := range strings.Split(target, "/") {
		if seg == ".." || seg == "." {
			return true
		}
	}
	return false
}

func printPathReport(hits []pathViolation, charterRoot string) {
	if len(hits) == 0 {
		fmt.Fprintf(os.Stdout, "✓ paths: every inter-charter link is charter-root-relative (no ../ or ./ segments)\n")
		return
	}
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].File == hits[j].File {
			return hits[i].Line < hits[j].Line
		}
		return hits[i].File < hits[j].File
	})
	fmt.Fprintf(os.Stdout, "✗ paths: %d inter-charter link(s) use forbidden ../ or ./ segments\n", len(hits))
	for _, h := range hits {
		fmt.Fprintf(os.Stdout, "    %s:%d  %s\n", h.File, h.Line, h.Link)
	}
	fmt.Fprintf(os.Stdout, "  Convention: inter-charter links are written relative to the charter root\n")
	fmt.Fprintf(os.Stdout, "  (e.g. `corpus/process/spec.md`, not `../../corpus/process/spec.md`).\n")
	fmt.Fprintf(os.Stdout, "  See docs/conventions.md for the rule.\n")
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
	lf, err := lockfile.Read(projectDir, charterRoot)
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
			if err := policies.Reset(d.Policy, projectDir, charterRoot); err != nil {
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

	err := fs.WalkDir(scaffold.Templates, "charter", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel := strings.TrimPrefix(path, "charter/")
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
	fmt.Fprintln(os.Stdout, "  changed upstream. Compare with `keystone init --reset --i-understand-this-is-destructive`")
	fmt.Fprintln(os.Stdout, "  on a copy to see what changed.")
}
