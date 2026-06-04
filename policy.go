package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// runPolicy dispatches `keystone policy <subcommand> ...`. `list` and `remove`
// are slated for a later slice.
func runPolicy(args []string) error {
	if len(args) == 0 {
		printPolicyUsage(os.Stderr)
		return fmt.Errorf("policy requires a subcommand")
	}
	switch args[0] {
	case "add":
		return runPolicyAdd(args[1:])
	case "update":
		return runPolicyUpdate(args[1:])
	case "help", "--help", "-h":
		printPolicyUsage(os.Stdout)
		return nil
	default:
		return fmt.Errorf("unknown policy subcommand %q (try: add, update)", args[0])
	}
}

func printPolicyUsage(w *os.File) {
	fmt.Fprint(w, `keystone policy — manage installed org policies

Usage:
  keystone policy add <ref> [--dir <path>]
  keystone policy update <name> [<new-ref>] [--dir <path>] [--force]
  keystone policy help

Commands:
  add       Install a new org policy into an existing harness. Errors if a
            policy with the same name is already recorded in the lockfile.
  update    Re-resolve a policy and replace its content. Refuses to overwrite
            local edits unless --force.
`)
}

// installPolicies resolves, validates, and copies every policy in refs into
// destDir. Updates harness/.keystone.lock with one entry per installed policy.
// Stops at the first failure (no rollback in v1) — partial state on disk is
// possible if a later policy errors after an earlier one succeeded.
//
// Returns the map of policy-name → PolicyLock for the policies installed in
// this call, suitable for rendering into INSTALL_PROFILE.md.
func installPolicies(destDir string, refs []string) (map[string]PolicyLock, error) {
	if len(refs) == 0 {
		return nil, nil
	}

	lf, err := ensureLockfile(destDir)
	if err != nil {
		return nil, err
	}

	installed := map[string]PolicyLock{}
	for _, raw := range refs {
		name, lock, err := installOnePolicy(destDir, raw)
		if err != nil {
			return nil, fmt.Errorf("policy %q: %w", raw, err)
		}
		lf.Policies[name] = lock
		installed[name] = lock
	}

	if err := writeLockfile(destDir, lf); err != nil {
		return nil, err
	}
	fmt.Fprintf(os.Stdout, "  wrote: %s\n", filepath.Join(destDir, KeystoneLockfile))
	return installed, nil
}

// installOnePolicy handles one ref end-to-end: parse → resolve → load manifest
// → validate content → copy into destDir → hash installed files. Returns the
// policy's manifest name and the lockfile entry to record.
func installOnePolicy(destDir, raw string) (string, PolicyLock, error) {
	ref, err := parsePolicyRef(raw)
	if err != nil {
		return "", PolicyLock{}, err
	}

	fmt.Fprintf(os.Stdout, "▸ installing policy %s\n", raw)

	resolved, err := resolvePolicy(ref)
	if err != nil {
		return "", PolicyLock{}, err
	}
	defer os.RemoveAll(resolved.LocalDir)

	manifest, err := loadManifest(resolved.LocalDir)
	if err != nil {
		return "", PolicyLock{}, err
	}

	if _, err := validatePolicyContent(resolved.LocalDir, manifest); err != nil {
		return "", PolicyLock{}, err
	}

	srcFS := os.DirFS(resolved.LocalDir)
	if err := copyTree(srcFS, PolicyContentRoot, destDir, overwrite); err != nil {
		return "", PolicyLock{}, fmt.Errorf("copy policy content: %w", err)
	}

	namespaceDir := filepath.Join("harness", "policies", manifest.Namespace())
	fileHashes, err := hashFilesUnder(destDir, namespaceDir)
	if err != nil {
		return "", PolicyLock{}, fmt.Errorf("hash installed files: %w", err)
	}

	lock := PolicyLock{
		SourceRef:       raw,
		ResolvedSHA:     resolved.ResolvedSHA,
		PolicyVersion:   manifest.Version,
		KeystoneVersion: version,
		Files:           fileHashes,
	}
	return manifest.Name, lock, nil
}
