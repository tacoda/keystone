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
	dir, asJSON := ".", false
	pos, help, err := parseFlagsArgs(args,
		map[string]*string{"--dir": &dir},
		map[string]*bool{"--json": &asJSON},
		func() { printModeUsage(os.Stdout) })
	if help || err != nil {
		return err
	}
	if len(pos) > 1 {
		return fmt.Errorf("mode takes one value; also got %q", pos[1])
	}
	mode := ""
	if len(pos) == 1 {
		mode = pos[0]
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	path, err := modesRulePath(absDir)
	if err != nil {
		return err
	}
	if mode == "" {
		cur, err := currentMode(absDir, path)
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
	prev, err := currentMode(absDir, path)
	if err != nil {
		return err
	}
	store, err := setActiveMode(absDir, path, mode)
	if err != nil {
		return err
	}
	if asJSON {
		return json.NewEncoder(os.Stdout).Encode(map[string]string{"mode": mode, "previous": prev})
	}
	fmt.Fprintf(os.Stdout, "✓ active mode: %s → %s (%s)\n", prev, mode, store)
	return nil
}

func printModeUsage(w *os.File) {
	fmt.Fprint(w, `keystone mode [paired|solo|autopilot] [--dir D] [--json]

Read the active pacing mode, or set it. The source of truth is the
**Active mode:** line of .agents/rules/process/modes.md when that rule
exists (workflow pack); otherwise .keystone/mode.
`)
}

// currentMode reads the active mode: the modes rule's line when the rule
// exists, else the .keystone/mode state file, else the default (paired).
func currentMode(absDir, path string) (string, error) {
	if path != "" {
		return currentModeFromFile(path)
	}
	data, err := os.ReadFile(filepath.Join(absDir, filepath.FromSlash(modeStateFile)))
	if os.IsNotExist(err) {
		return "paired", nil
	}
	if err != nil {
		return "", err
	}
	if m := strings.TrimSpace(string(data)); validMode(m) {
		return m, nil
	}
	return "paired", nil
}

func validMode(m string) bool {
	for _, c := range modeChoices {
		if m == c {
			return true
		}
	}
	return false
}

// modeStateFile is the fallback store when the charter has no
// process/modes rule (a plain init without the workflow pack). The rule's
// `**Active mode:**` line stays the source of truth when it exists.
const modeStateFile = config.StateRoot + "/mode"

// modesRulePath resolves the process/modes rule's file under absDir, or
// "" when the charter has no such rule.
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
	return "", nil
}

var activeModeLine = regexp.MustCompile(`(?m)^\*\*Active mode:\s*([^*]+?)\*\*\s*$`)

// currentModeFromFile reads the `**Active mode:**` line of the modes rule.
func currentModeFromFile(path string) (string, error) {
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

// setActiveMode stores the mode: always in the .keystone/mode state file,
// and — when the process/modes rule exists — also in its `**Active mode:**`
// line and frontmatter description (the index carries the description),
// leaving the rest of the file untouched, then re-indexes. Returns the
// repo-relative store that changed.
func setActiveMode(absDir, path, mode string) (string, error) {
	state := filepath.Join(absDir, filepath.FromSlash(modeStateFile))
	if err := os.MkdirAll(filepath.Dir(state), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(state, []byte(mode+"\n"), 0o644); err != nil {
		return "", err
	}
	if path == "" {
		return modeStateFile, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
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
		return "", err
	}
	if err := reindex(absDir); err != nil {
		return "", err
	}
	return relTo(absDir, path), nil
}
