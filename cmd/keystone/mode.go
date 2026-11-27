package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// modesRuleID is where the pacing mode lives: the process/modes rule's
// `**Active mode:**` line is the source of truth (see the keystone-mode
// skill). The workflow starter pack plants it; a charter without the
// rule has no modes to switch.
const modesRuleID = "process/modes"

var modeChoices = []string{"paired", "solo", "autopilot"}

var modeSummaries = map[string]string{
	"paired":    "user at the keyboard; confirm at every gate",
	"solo":      "user around; ask only at genuine forks",
	"autopilot": "user away; pause only on iron-law violations or destructive actions",
}

func modeCmd() *cobra.Command {
	c := &cobra.Command{
		Use:                "mode [paired|solo|autopilot]",
		Short:              "Read or set the charter pacing mode (paired / solo / autopilot)",
		DisableFlagParsing: true,
		RunE:               runAndForward(runMode),
	}
	return c
}

// runMode handles `keystone mode [mode] [--dir D] [--json]`. No argument
// prints the active mode; an argument switches it by rewriting the
// `**Active mode:**` line of rules/process/modes.md in place (plus the
// frontmatter description, which feeds the index), then re-indexes.
func runMode(args []string) error {
	mode, asJSON := "", false
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--json":
			asJSON = true
		case a == "--dir":
			i++
			if i >= len(args) {
				return fmt.Errorf("--dir requires a value")
			}
		case strings.HasPrefix(a, "--dir="):
		case strings.HasPrefix(a, "-"):
			return fmt.Errorf("unknown flag %q", a)
		default:
			if mode != "" {
				return fmt.Errorf("mode takes one value; also got %q", a)
			}
			mode = a
		}
	}
	absDir, err := filepath.Abs(dirArg(args))
	if err != nil {
		return err
	}
	path, err := modesRulePath(absDir)
	if err != nil {
		return err
	}
	if mode == "" {
		cur, err := currentMode(path)
		if err != nil {
			return err
		}
		if asJSON {
			return json.NewEncoder(os.Stdout).Encode(map[string]string{"mode": cur})
		}
		fmt.Fprintf(os.Stdout, "active mode: %s\n", cur)
		for _, m := range modeChoices {
			fmt.Fprintf(os.Stdout, "  %-9s %s\n", m, modeSummaries[m])
		}
		return nil
	}
	if !validMode(mode) {
		return fmt.Errorf("unknown mode %q (allowed: %s)", mode, strings.Join(modeChoices, ", "))
	}
	prev, err := currentMode(path)
	if err != nil {
		return err
	}
	if err := setActiveMode(absDir, path, mode); err != nil {
		return err
	}
	if asJSON {
		return json.NewEncoder(os.Stdout).Encode(map[string]string{"mode": mode, "previous": prev})
	}
	fmt.Fprintf(os.Stdout, "✓ active mode: %s → %s (%s)\n", prev, mode, relTo(absDir, path))
	return nil
}

func validMode(m string) bool {
	for _, c := range modeChoices {
		if m == c {
			return true
		}
	}
	return false
}

// modesRulePath resolves the process/modes rule's file under absDir.
func modesRulePath(absDir string) (string, error) {
	prims, _, err := primitive.Walk(absDir, config.DefaultAgentsRoot)
	if err != nil {
		return "", err
	}
	for i := range prims {
		if prims[i].Kind == string(primitive.KindRule) && prims[i].ID == modesRuleID {
			return filepath.Join(absDir, filepath.FromSlash(prims[i].Path)), nil
		}
	}
	return "", fmt.Errorf("no %s rule under %s — the pacing mode lives at .agents/rules/%s.md (planted by the workflow starter pack)", modesRuleID, config.DefaultAgentsRoot, modesRuleID)
}

var activeModeLine = regexp.MustCompile(`(?m)^\*\*Active mode:\s*([^*]+?)\*\*\s*$`)

// currentMode reads the `**Active mode:**` line of the modes rule.
func currentMode(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	m := activeModeLine.FindSubmatch(data)
	if m == nil {
		return "", fmt.Errorf("%s has no '**Active mode:**' line", path)
	}
	return string(m[1]), nil
}

// setActiveMode rewrites the `**Active mode:**` line in the body and the
// `description:` line in the frontmatter (the index carries the
// description), leaving the rest of the file untouched, then re-indexes.
func setActiveMode(absDir, path, mode string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	out := activeModeLine.ReplaceAll(data, []byte("**Active mode: "+mode+"**"))
	fm, body, ok := primitive.SplitFrontmatter(string(out))
	if ok {
		lines := strings.Split(fm, "\n")
		for i, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "description:") && strings.Contains(line, "Active mode:") {
				lines[i] = "description: 'Active mode: " + mode + "'"
			}
		}
		out = []byte("---\n" + strings.Join(lines, "\n") + "---\n" + body)
	}
	if err := os.WriteFile(path, out, 0o644); err != nil {
		return err
	}
	return reindex(absDir)
}
