package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tacoda/keystone/internal/framework/loader"
	"github.com/tacoda/keystone/internal/framework/lockfile"
	"github.com/tacoda/keystone/internal/framework/manifest"
)

// runPolicyAdd handles `keystone policy add <ref> [--dir <path>]`.
//
// Resolves the policy from <ref>, validates its content, refuses to install
// if a policy with the same name is already recorded in the lockfile, then
// copies the namespace tree and writes the lockfile entry.
func runPolicyAdd(args []string) error {
	dir := "."
	var positional []string

	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--help" || a == "-h":
			printPolicyAddUsage(os.Stdout)
			return nil
		case a == "--dir":
			if i+1 >= len(args) {
				return fmt.Errorf("flag %s requires a value", a)
			}
			dir = args[i+1]
			i++
		case strings.HasPrefix(a, "--dir="):
			dir = strings.TrimPrefix(a, "--dir=")
		case strings.HasPrefix(a, "-"):
			return fmt.Errorf("unknown flag %s", a)
		default:
			positional = append(positional, a)
		}
	}

	if len(positional) != 1 {
		return fmt.Errorf("policy add requires exactly one ref (e.g. `keystone policy add git+https://github.com/acme/policy.git#v1.0.0`)")
	}
	raw := positional[0]

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

	lf, err := ensureLockfile(absDir)
	if err != nil {
		return err
	}

	ref, err := loader.ParsePolicyRef(raw)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "▸ installing policy %s\n", raw)

	resolved, err := loader.ResolvePolicy(ref)
	if err != nil {
		return err
	}
	defer os.RemoveAll(resolved.LocalDir)

	mf, err := manifest.Load(resolved.LocalDir)
	if err != nil {
		return err
	}

	if _, exists := lf.Policies[mf.Name]; exists {
		return fmt.Errorf(
			"policy %q is already installed (recorded in %s); use `keystone policy update %s` to re-resolve or change the ref",
			mf.Name, lockfile.File, mf.Name,
		)
	}

	if _, err := manifest.ValidateContent(resolved.LocalDir, mf); err != nil {
		return err
	}

	srcFS := os.DirFS(resolved.LocalDir)
	if err := copyTree(srcFS, manifest.PolicyContentRoot, absDir, overwrite); err != nil {
		return fmt.Errorf("copy policy content: %w", err)
	}

	namespaceDir := filepath.Join("harness", "policies", mf.Namespace())
	fileHashes, err := lockfile.HashFilesUnder(absDir, namespaceDir)
	if err != nil {
		return fmt.Errorf("hash installed files: %w", err)
	}

	lf.Policies[mf.Name] = lockfile.PolicyLock{
		SourceRef:       raw,
		ResolvedSHA:     resolved.ResolvedSHA,
		PolicyVersion:   mf.Version,
		KeystoneVersion: version,
		Tier:            mf.ResolvedTier(),
		Strict:          mf.Strict,
		Required:        mf.Required,
		Files:           fileHashes,
	}
	if err := lockfile.Write(absDir, lf); err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "✓ installed %s %s (%s)\n",
		mf.Name, mf.Version, resolved.ResolvedSHA[:displaySHALen(resolved.ResolvedSHA)])

	res, verr := verifyPolicies(absDir)
	if verr != nil {
		return fmt.Errorf("policy verify: %w", verr)
	}
	if printVerifyReport(absDir, res) {
		return fmt.Errorf("policy %q installed but strict cascade is violated — resolve the shadowing file(s) above", mf.Name)
	}
	return nil
}

func printPolicyAddUsage(w *os.File) {
	fmt.Fprint(w, `keystone policy add — install an org policy into an existing harness

Usage:
  keystone policy add <ref> [--dir <path>]

Fetches and installs a policy from <ref>. v1 supports git+<url>[#<rev>]:
  keystone policy add git+https://github.com/acme/policy.git#v1.2.0

Errors out if a policy with the same name is already recorded in
harness/keystone.lock.json — use 'keystone policy update' to re-resolve, or
remove the policy first to re-add it.

Flags:
  --dir <path>   Directory containing harness/ (defaults to cwd).
`)
}
