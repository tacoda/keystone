package eval

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// BaselineDiff is the side-by-side result of a baseline run.
type BaselineDiff struct {
	Baseline Report     `json:"baseline"`
	Current  Report     `json:"current"`
	Changes  []Change   `json:"changes"`
	Summary  DiffStats  `json:"summary"`
}

type Change struct {
	ID        string `json:"id"`
	Level     string `json:"level"`
	WasStatus string `json:"was_status"`
	NowStatus string `json:"now_status"`
	Kind      string `json:"kind"` // "regression" | "fix" | "new" | "removed" | "stable"
}

type DiffStats struct {
	Regressions int `json:"regressions"`
	Fixes       int `json:"fixes"`
	NewEvals    int `json:"new"`
	Removed     int `json:"removed"`
	Stable      int `json:"stable"`
}

// RunWithBaseline materializes `ref` into a git worktree, runs evals
// there, then runs evals against the current tree, and diffs the
// two. Worktree is cleaned up on return.
//
// Requires `git` on PATH and a clean enough working tree for git
// worktree to operate. Errors are surfaced — no partial-state
// success.
func RunWithBaseline(ctx context.Context, projectDir, ref, filter string) (*BaselineDiff, error) {
	if _, err := exec.LookPath("git"); err != nil {
		return nil, fmt.Errorf("git not on PATH: %w", err)
	}

	// Create a sibling temp dir for the worktree.
	tmp, err := os.MkdirTemp("", "keystone-baseline-*")
	if err != nil {
		return nil, fmt.Errorf("temp dir: %w", err)
	}
	defer func() {
		_ = removeWorktree(ctx, projectDir, tmp)
		_ = os.RemoveAll(tmp)
	}()

	addCmd := exec.CommandContext(ctx, "git", "-C", projectDir, "worktree", "add", "--detach", tmp, ref)
	if out, err := addCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("git worktree add %s: %v\n%s", ref, err, string(out))
	}

	baselineSpecs, err := LoadAll(tmp)
	if err != nil {
		return nil, fmt.Errorf("baseline LoadAll: %w", err)
	}
	baselineRep := Run(ctx, tmp, baselineSpecs, filter)

	currentSpecs, err := LoadAll(projectDir)
	if err != nil {
		return nil, fmt.Errorf("current LoadAll: %w", err)
	}
	currentRep := Run(ctx, projectDir, currentSpecs, filter)

	diff := &BaselineDiff{
		Baseline: baselineRep,
		Current:  currentRep,
		Changes:  diffResults(baselineRep, currentRep),
	}
	for _, c := range diff.Changes {
		switch c.Kind {
		case "regression":
			diff.Summary.Regressions++
		case "fix":
			diff.Summary.Fixes++
		case "new":
			diff.Summary.NewEvals++
		case "removed":
			diff.Summary.Removed++
		case "stable":
			diff.Summary.Stable++
		}
	}
	return diff, nil
}

func removeWorktree(ctx context.Context, projectDir, path string) error {
	cmd := exec.CommandContext(ctx, "git", "-C", projectDir, "worktree", "remove", "--force", path)
	return cmd.Run()
}

func diffResults(baseline, current Report) []Change {
	type key struct{ id, level string }
	b := map[key]string{}
	for _, r := range baseline.Results {
		b[key{r.ID, r.Level}] = r.Status
	}
	c := map[key]string{}
	for _, r := range current.Results {
		c[key{r.ID, r.Level}] = r.Status
	}
	var changes []Change
	seen := map[key]bool{}
	for k, was := range b {
		seen[k] = true
		now, ok := c[k]
		if !ok {
			changes = append(changes, Change{ID: k.id, Level: k.level, WasStatus: was, NowStatus: "", Kind: "removed"})
			continue
		}
		kind := "stable"
		if was != now {
			if now == "fail" && was == "pass" {
				kind = "regression"
			} else if now == "pass" && was == "fail" {
				kind = "fix"
			} else {
				kind = "stable" // skip→pass etc. — not graded as regression
			}
		}
		changes = append(changes, Change{ID: k.id, Level: k.level, WasStatus: was, NowStatus: now, Kind: kind})
	}
	for k, now := range c {
		if seen[k] {
			continue
		}
		changes = append(changes, Change{ID: k.id, Level: k.level, WasStatus: "", NowStatus: now, Kind: "new"})
	}
	return changes
}

// RenderBaselineMarkdown formats a baseline diff as a human report.
func RenderBaselineMarkdown(d *BaselineDiff) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# keystone eval — baseline diff\n\n")
	fmt.Fprintf(&b, "- regressions: **%d**\n- fixes:       **%d**\n- new:         %d\n- removed:     %d\n- stable:      %d\n\n", d.Summary.Regressions, d.Summary.Fixes, d.Summary.NewEvals, d.Summary.Removed, d.Summary.Stable)
	if d.Summary.Regressions > 0 {
		fmt.Fprintln(&b, "## ⚠ regressions")
		for _, c := range d.Changes {
			if c.Kind == "regression" {
				fmt.Fprintf(&b, "- %s (%s): was `%s` → now `%s`\n", c.ID, c.Level, c.WasStatus, c.NowStatus)
			}
		}
		fmt.Fprintln(&b)
	}
	if d.Summary.Fixes > 0 {
		fmt.Fprintln(&b, "## ✓ fixes")
		for _, c := range d.Changes {
			if c.Kind == "fix" {
				fmt.Fprintf(&b, "- %s (%s): was `%s` → now `%s`\n", c.ID, c.Level, c.WasStatus, c.NowStatus)
			}
		}
		fmt.Fprintln(&b)
	}
	if d.Summary.NewEvals > 0 {
		fmt.Fprintln(&b, "## + new evals")
		for _, c := range d.Changes {
			if c.Kind == "new" {
				fmt.Fprintf(&b, "- %s (%s): %s\n", c.ID, c.Level, c.NowStatus)
			}
		}
		fmt.Fprintln(&b)
	}
	if d.Summary.Removed > 0 {
		fmt.Fprintln(&b, "## − removed evals")
		for _, c := range d.Changes {
			if c.Kind == "removed" {
				fmt.Fprintf(&b, "- %s (%s): was %s\n", c.ID, c.Level, c.WasStatus)
			}
		}
	}
	return b.String()
}

// resolveProjectDir is a small helper for callers that pass cwd —
// the eval engine works in absolute paths so paths join cleanly.
func resolveProjectDir(dir string) (string, error) {
	if filepath.IsAbs(dir) {
		return dir, nil
	}
	return filepath.Abs(dir)
}
