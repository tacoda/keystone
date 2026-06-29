package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// workDirRel is where document instances live — tracked in git as
// durable team memory (filled plans, reviews, ADRs, retros). Routed
// through config so the standardized harness root is the only place to
// change the location.
var workDirRel = filepath.Join(config.KeystoneDir(config.DefaultHarnessRoot), "work")

var gateLineRE = regexp.MustCompile(`(?m)^gate:.*$`)

// runDocument dispatches `keystone document <list|promote>`.
func runDocument(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: keystone document <list|promote> [args]")
	}
	switch args[0] {
	case "list":
		return runDocumentList(args[1:])
	case "promote":
		return runDocumentPromote(args[1:])
	default:
		return fmt.Errorf("unknown document subcommand %q (use: list, promote)", args[0])
	}
}

// runDocumentList enumerates document instances under .keystone/work/,
// printing id, current gate, and path.
func runDocumentList(args []string) error {
	dir := "."
	if len(args) == 1 {
		dir = args[0]
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	workAbs := filepath.Join(absDir, workDirRel)
	if _, err := os.Stat(workAbs); os.IsNotExist(err) {
		fmt.Fprintln(os.Stdout, "no documents — .keystone/work/ does not exist yet")
		return nil
	}
	count := 0
	if err := filepath.WalkDir(workAbs, listWalkFn(absDir, &count)); err != nil {
		return err
	}
	if count == 0 {
		fmt.Fprintln(os.Stdout, "no documents under .keystone/work/")
	}
	return nil
}

// listWalkFn returns the WalkDir callback that prints each document
// instance and increments count. Split out to keep runDocumentList flat.
func listWalkFn(absDir string, count *int) fs.WalkDirFunc {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		if printDocEntry(absDir, path) {
			*count++
		}
		return nil
	}
}

// printDocEntry prints one instance line (id, gate, path) and reports
// whether the file was a parseable document. Non-document files are
// skipped silently.
func printDocEntry(absDir, path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	fm, ok, _ := primitive.Parse(string(data))
	if !ok {
		return false
	}
	gate := fm.Gate
	if gate == "" {
		gate = "(unset)"
	}
	rel, _ := filepath.Rel(absDir, path)
	fmt.Fprintf(os.Stdout, "  %-28s [%s]  %s\n", fm.ID, gate, rel)
	return true
}

// runDocumentPromote advances one document instance to a new gate. The
// transition must move forward through the instance's own `gates:`
// order; backward or unknown gates are rejected. The change is printed
// (old → new) — never a silent overwrite.
func runDocumentPromote(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: keystone document promote <path> <gate>")
	}
	path, target := args[0], args[1]
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	fmText, body, ok := primitive.SplitFrontmatter(string(data))
	if !ok {
		return fmt.Errorf("%s has no frontmatter", path)
	}
	fm, _, err := primitive.Parse(string(data))
	if err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	current, err := validateGateTransition(fm, target)
	if err != nil {
		return err
	}
	out := "---\n" + rewriteGate(fmText, target) + "---\n" + body
	if err := os.WriteFile(path, []byte(out), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	fmt.Fprintf(os.Stdout, "  %s: %s → %s\n", fm.ID, current, target)

	// Framework hook: the document crossed a gate. Fired after the
	// transition is committed, so a hook failure is logged, not fatal —
	// the gate has already advanced.
	if err := runHookFire([]string{"on-gate", "--type", fm.Type, "--command", target}); err != nil {
		fmt.Fprintf(os.Stderr, "on-gate hooks: %v\n", err)
	}
	return nil
}

// validateGateTransition checks that target is a forward move through the
// document's own `gates:` order, returning the resolved current gate.
func validateGateTransition(fm primitive.Frontmatter, target string) (string, error) {
	if len(fm.Gates) == 0 {
		return "", fmt.Errorf("document declares no `gates:` — not promotable")
	}
	current := fm.Gate
	if current == "" {
		current = fm.Gates[0]
	}
	curIdx := indexOf(fm.Gates, current)
	tgtIdx := indexOf(fm.Gates, target)
	if tgtIdx < 0 {
		return "", fmt.Errorf("gate %q is not in this document's gates %v", target, fm.Gates)
	}
	if curIdx < 0 {
		return "", fmt.Errorf("current gate %q is not in gates %v — fix the file first", current, fm.Gates)
	}
	if tgtIdx <= curIdx {
		return "", fmt.Errorf("cannot promote %q → %q: not a forward transition in %v", current, target, fm.Gates)
	}
	return current, nil
}

// rewriteGate returns the frontmatter block with `gate:` set to target,
// inserting the line after the first line when absent. fmText ends with a
// trailing newline (per SplitFrontmatter), so the result always does too.
func rewriteGate(fmText, target string) string {
	if gateLineRE.MatchString(fmText) {
		out := gateLineRE.ReplaceAllString(fmText, "gate: "+target)
		return ensureTrailingNewline(out)
	}
	nl := strings.IndexByte(fmText, '\n')
	if nl < 0 {
		return fmText + "\ngate: " + target + "\n"
	}
	return fmText[:nl+1] + "gate: " + target + "\n" + fmText[nl+1:]
}

func ensureTrailingNewline(s string) string {
	if strings.HasSuffix(s, "\n") {
		return s
	}
	return s + "\n"
}

func indexOf(ss []string, want string) int {
	for i, s := range ss {
		if s == want {
			return i
		}
	}
	return -1
}
