package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tacoda/keystone/internal/framework/loader"
	"github.com/tacoda/keystone/internal/framework/manifest"
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
	case "verify":
		return runPolicyVerify(args[1:])
	case "help", "--help", "-h":
		printPolicyUsage(os.Stdout)
		return nil
	default:
		return fmt.Errorf("unknown policy subcommand %q (try: add, update, verify)", args[0])
	}
}

func printPolicyUsage(w *os.File) {
	fmt.Fprint(w, `keystone policy — manage installed policies

Usage:
  keystone policy add <ref> [--dir <path>]
  keystone policy update <name> [<new-ref>] [--dir <path>] [--force]
  keystone policy verify [--dir <path>]
  keystone policy help

Commands:
  add       Install a new policy into an existing harness. Errors if a
            policy with the same name is already recorded in the lockfile.
  update    Re-resolve a policy and replace its content. Refuses to overwrite
            local edits unless --force.
  verify    Walk every installed policy's strict items and report any file in
            a lower tier (team policies, project tree) that overrides them.
            Exits non-zero on any violation.
`)
}

// runPolicyVerify handles `keystone policy verify [--dir <path>]`. Reports
// any strict-cascade violations in the install dir. Exits non-zero when
// violations exist; returns nil on a clean tree.
func runPolicyVerify(args []string) error {
	dir := "."
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--help" || a == "-h":
			fmt.Fprint(os.Stdout, `keystone policy verify — audit the strict cascade

Usage:
  keystone policy verify [--dir <path>]

For each installed policy with strict items, walks lower tiers (team
policies + project tree) and reports any file that overrides a strict
item. Exits non-zero on any violation.

Flags:
  --dir <path>   Directory containing harness/ (defaults to cwd).
`)
			return nil
		case a == "--dir":
			if i+1 >= len(args) {
				return fmt.Errorf("flag %s requires a value", a)
			}
			dir = args[i+1]
			i++
		default:
			return fmt.Errorf("unknown argument %s", a)
		}
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}
	if _, err := os.Stat(filepath.Join(absDir, "harness")); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no harness/ in %s — run `keystone init` first", absDir)
		}
		return err
	}

	res, err := verifyPolicies(absDir)
	if err != nil {
		return err
	}
	if printVerifyReport(absDir, res) {
		return fmt.Errorf("strict cascade is violated")
	}
	return nil
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
	ref, err := loader.ParsePolicyRef(raw)
	if err != nil {
		return "", PolicyLock{}, err
	}

	fmt.Fprintf(os.Stdout, "▸ installing policy %s\n", raw)

	resolved, err := loader.ResolvePolicy(ref)
	if err != nil {
		return "", PolicyLock{}, err
	}
	defer os.RemoveAll(resolved.LocalDir)

	mf, err := manifest.Load(resolved.LocalDir)
	if err != nil {
		return "", PolicyLock{}, err
	}

	if _, err := manifest.ValidateContent(resolved.LocalDir, mf); err != nil {
		return "", PolicyLock{}, err
	}

	srcFS := os.DirFS(resolved.LocalDir)
	if err := copyTree(srcFS, manifest.PolicyContentRoot, destDir, overwrite); err != nil {
		return "", PolicyLock{}, fmt.Errorf("copy policy content: %w", err)
	}

	namespaceDir := filepath.Join("harness", "policies", mf.Namespace())
	fileHashes, err := hashFilesUnder(destDir, namespaceDir)
	if err != nil {
		return "", PolicyLock{}, fmt.Errorf("hash installed files: %w", err)
	}

	lock := PolicyLock{
		SourceRef:       raw,
		ResolvedSHA:     resolved.ResolvedSHA,
		PolicyVersion:   mf.Version,
		KeystoneVersion: version,
		Tier:            mf.ResolvedTier(),
		Strict:          mf.Strict,
		Required:        mf.Required,
		Files:           fileHashes,
	}
	return mf.Name, lock, nil
}
