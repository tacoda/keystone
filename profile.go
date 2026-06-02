package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// writeInstallProfile renders sel as harness/corpus/state/INSTALL_PROFILE.md
// under destDir. Overwrites any existing file (the file is install-scoped —
// re-running init should reset it).
func writeInstallProfile(destDir string, sel Selections) error {
	path := filepath.Join(destDir, "harness", "corpus", "state", "INSTALL_PROFILE.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	var b strings.Builder
	fmt.Fprintf(&b, "---\n")
	fmt.Fprintf(&b, "created: %s\n", time.Now().UTC().Format("2006-01-02"))
	fmt.Fprintf(&b, "keystone_version: %s\n", version)
	fmt.Fprintf(&b, "---\n\n")
	fmt.Fprintf(&b, "# Install Profile\n\n")
	fmt.Fprintf(&b, "Selections captured by `keystone init`. Read by the **bootstrap** action; safe to edit by hand.\n\n")
	fmt.Fprintf(&b, "## Selections\n\n")
	fmt.Fprintf(&b, "| Category | Value(s) |\n")
	fmt.Fprintf(&b, "|---|---|\n")

	// Iterate in catalog order, not map order, so the file is stable.
	for _, c := range categories {
		values := sel[c.ID]
		if len(values) == 0 {
			fmt.Fprintf(&b, "| %s | _(unset)_ |\n", c.ID)
			continue
		}
		fmt.Fprintf(&b, "| %s | %s |\n", c.ID, strings.Join(values, ", "))
	}

	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "  wrote: %s\n", path)
	return nil
}

// readInstalledAgents parses the agent row of INSTALL_PROFILE.md. Returns the
// list of agent IDs already recorded, or an empty slice if the row is unset.
// Returns an error if the profile file is missing or unreadable.
func readInstalledAgents(destDir string) ([]string, error) {
	path := filepath.Join(destDir, "harness", "corpus", "state", "INSTALL_PROFILE.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "| agent ") && !strings.HasPrefix(line, "| agent|") {
			continue
		}
		// Row shape: "| agent | claude-code, cursor |"
		cells := strings.Split(line, "|")
		if len(cells) < 3 {
			continue
		}
		val := strings.TrimSpace(cells[2])
		if val == "" || val == "_(unset)_" {
			return []string{}, nil
		}
		parts := strings.Split(val, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			if v := strings.TrimSpace(p); v != "" {
				out = append(out, v)
			}
		}
		return out, nil
	}
	return []string{}, nil
}

// appendAgentsToProfile rewrites INSTALL_PROFILE.md's agent row to include the
// new agents alongside the existing ones, preserving every other row.
func appendAgentsToProfile(destDir string, newAgents []string) error {
	path := filepath.Join(destDir, "harness", "corpus", "state", "INSTALL_PROFILE.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	existing, err := readInstalledAgents(destDir)
	if err != nil {
		return err
	}
	merged := append([]string{}, existing...)
	seen := map[string]bool{}
	for _, a := range merged {
		seen[a] = true
	}
	for _, a := range newAgents {
		if !seen[a] {
			merged = append(merged, a)
			seen[a] = true
		}
	}

	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "| agent ") || strings.HasPrefix(line, "| agent|") {
			lines[i] = fmt.Sprintf("| agent | %s |", strings.Join(merged, ", "))
			break
		}
	}

	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "  updated: %s\n", path)
	return nil
}
