package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tacoda/keystone/internal/framework/lockfile"
)

// writeInstallProfile renders sel as <harnessRoot>/corpus/state/INSTALL_PROFILE.md
// under destDir. Overwrites any existing file (the file is install-scoped —
// re-running init should reset it).
//
// The profile is the human-readable record. Machine state (keystone version,
// agents, plugins) lives in <harnessRoot>/keystone.lock.json — written separately.
func writeInstallProfile(destDir, harnessRoot string, sel Selections) error {
	path := filepath.Join(destDir, harnessRoot, "corpus", "state", "INSTALL_PROFILE.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	var b strings.Builder
	fmt.Fprintf(&b, "---\n")
	fmt.Fprintf(&b, "created: %s\n", time.Now().UTC().Format("2006-01-02"))
	fmt.Fprintf(&b, "---\n\n")
	fmt.Fprintf(&b, "# Install Profile\n\n")
	lockPath := filepath.ToSlash(filepath.Join(harnessRoot, "keystone.lock.json"))
	fmt.Fprintf(&b, "Selections captured by `keystone init`. Read by the **bootstrap** action; safe to edit by hand. Machine state (keystone version, agents, plugins) lives in [`%s`](%s) at the repo root.\n\n", lockPath, lockPath)
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

// readInstalledAgents returns the list of agent IDs recorded in the lockfile.
// Falls back to parsing INSTALL_PROFILE.md for installs that predate the
// lockfile.
func readInstalledAgents(destDir, harnessRoot string) ([]string, error) {
	lf, err := lockfile.Read(destDir, harnessRoot)
	if err != nil {
		return nil, err
	}
	if len(lf.Keystone.Agents) > 0 {
		return append([]string{}, lf.Keystone.Agents...), nil
	}
	return readInstalledAgentsFromProfile(destDir, harnessRoot)
}

// readKeystoneVersion returns the binary version that last touched the install,
// from the lockfile. Falls back to INSTALL_PROFILE.md frontmatter for installs
// that predate the lockfile. Returns "" if neither source has a value.
func readKeystoneVersion(destDir, harnessRoot string) (string, error) {
	lf, err := lockfile.Read(destDir, harnessRoot)
	if err != nil {
		return "", err
	}
	if lf.Keystone.Version != "" {
		return lf.Keystone.Version, nil
	}
	return readKeystoneVersionFromProfile(destDir, harnessRoot)
}

// updateKeystoneVersion sets the binary version in the lockfile, creating
// the file if needed. Backfills install state from INSTALL_PROFILE.md when
// the lockfile is empty.
func updateKeystoneVersion(destDir, harnessRoot, newVersion string) error {
	lf, err := ensureLockfile(destDir, harnessRoot)
	if err != nil {
		return err
	}
	lf.Keystone.Version = newVersion
	return lockfile.Write(destDir, harnessRoot, lf)
}

// appendInstalledAgents adds newAgents to the lockfile's agent list, preserving
// existing entries and order. Backfills from INSTALL_PROFILE.md when the
// lockfile is empty so old installs get a lockfile on first agent-add.
func appendInstalledAgents(destDir, harnessRoot string, newAgents []string) error {
	lf, err := ensureLockfile(destDir, harnessRoot)
	if err != nil {
		return err
	}
	if lf.Keystone.Version == "" {
		lf.Keystone.Version = version
	}
	if lf.Keystone.Installed == "" {
		lf.Keystone.Installed = time.Now().UTC().Format("2006-01-02")
	}

	seen := map[string]bool{}
	for _, a := range lf.Keystone.Agents {
		seen[a] = true
	}
	for _, a := range newAgents {
		if !seen[a] {
			lf.Keystone.Agents = append(lf.Keystone.Agents, a)
			seen[a] = true
		}
	}
	if err := lockfile.Write(destDir, harnessRoot, lf); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "  updated: %s\n", filepath.Join(destDir, lockfile.RelPath(harnessRoot)))
	return nil
}

// readKeystoneVersionFromProfile parses the `keystone_version:` frontmatter
// from INSTALL_PROFILE.md. Used as a backward-compat fallback for installs
// created before the lockfile existed. Returns "" if the field is missing.
func readKeystoneVersionFromProfile(destDir, harnessRoot string) (string, error) {
	path := filepath.Join(destDir, harnessRoot, "corpus", "state", "INSTALL_PROFILE.md")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "keystone_version:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "keystone_version:")), nil
		}
	}
	return "", nil
}

// readInstalledAgentsFromProfile parses the agent row of INSTALL_PROFILE.md.
// Used as a backward-compat fallback for installs created before the lockfile
// existed.
func readInstalledAgentsFromProfile(destDir, harnessRoot string) ([]string, error) {
	path := filepath.Join(destDir, harnessRoot, "corpus", "state", "INSTALL_PROFILE.md")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "| agent ") && !strings.HasPrefix(line, "| agent|") {
			continue
		}
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
