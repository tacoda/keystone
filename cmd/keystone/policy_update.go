package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tacoda/keystone/internal/framework/loader"
	"github.com/tacoda/keystone/internal/framework/lockfile"
	"github.com/tacoda/keystone/internal/framework/manifest"
)

// runPolicyUpdate handles `keystone policy update <name> [<new-ref>] [--dir <path>] [--harness-root <name>]`.
//
// Re-resolves the named policy (using the lockfile's stored ref, or the new
// ref if supplied), validates the new content, refuses to clobber locally
// edited files unless --force, then replaces the namespace tree and rewrites
// the lockfile entry.
func runPolicyUpdate(args []string) error {
	harnessRoot, args, err := extractHarnessRoot(args)
	if err != nil {
		return err
	}
	var force bool
	dir := "."
	var positional []string

	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--force" || a == "-force":
			force = true
		case a == "--help" || a == "-h":
			printPolicyUpdateUsage(os.Stdout)
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

	var name, newRef string
	switch len(positional) {
	case 0:
		return fmt.Errorf("policy update requires a policy name (e.g. `keystone policy update acme`)")
	case 1:
		name = positional[0]
	case 2:
		name = positional[0]
		newRef = positional[1]
	default:
		return fmt.Errorf("policy update takes at most two positional arguments (<name> [<new-ref>]); use --dir for the install directory")
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}

	lf, err := lockfile.Read(absDir, harnessRoot)
	if err != nil {
		return err
	}
	existing, ok := lf.Policies[name]
	if !ok {
		return fmt.Errorf("policy %q is not installed (not recorded in %s)", name, lockfile.RelPath(harnessRoot))
	}

	sourceRef := existing.SourceRef
	if newRef != "" {
		sourceRef = newRef
	}

	ref, err := loader.ParsePolicyRef(sourceRef)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "▸ updating policy %s (%s)\n", name, sourceRef)

	resolved, err := loader.ResolvePolicy(ref)
	if err != nil {
		return err
	}
	defer os.RemoveAll(resolved.LocalDir)

	mf, err := manifest.Load(resolved.LocalDir)
	if err != nil {
		return err
	}
	if mf.Name != name {
		return fmt.Errorf("policy at %s declares name %q, but you're updating %q", sourceRef, mf.Name, name)
	}
	if _, err := manifest.ValidateContent(resolved.LocalDir, mf); err != nil {
		return err
	}

	namespaceDir := filepath.Join(harnessRoot, "policies", mf.Namespace())
	if !force {
		dirty, err := dirtyFiles(absDir, namespaceDir, existing.Files)
		if err != nil {
			return err
		}
		if len(dirty) > 0 {
			return fmt.Errorf(
				"policy %q has local changes — pass --force to discard:\n  %s",
				name, strings.Join(dirty, "\n  "),
			)
		}
	}

	if existing.ResolvedSHA == resolved.ResolvedSHA && newRef == "" {
		fmt.Fprintf(os.Stdout, "  already at %s — nothing to do\n", resolved.ResolvedSHA[:displaySHALen(resolved.ResolvedSHA)])
		return nil
	}

	if err := os.RemoveAll(filepath.Join(absDir, namespaceDir)); err != nil {
		return fmt.Errorf("remove old namespace: %w", err)
	}

	srcFS := os.DirFS(resolved.LocalDir)
	srcRoot := filepath.Join(manifest.PolicyContentRoot, "harness")
	dest := filepath.Join(absDir, harnessRoot)
	if err := copyTree(srcFS, srcRoot, dest, overwrite); err != nil {
		return fmt.Errorf("copy policy content: %w", err)
	}

	newHashes, err := lockfile.HashFilesUnder(absDir, namespaceDir)
	if err != nil {
		return fmt.Errorf("hash installed files: %w", err)
	}

	lf.Policies[name] = lockfile.PolicyLock{
		SourceRef:       sourceRef,
		ResolvedSHA:     resolved.ResolvedSHA,
		PolicyVersion:   mf.Version,
		KeystoneVersion: version,
		Tier:            mf.ResolvedTier(),
		Strict:          mf.Strict,
		Required:        mf.Required,
		Files:           newHashes,
	}
	if err := lockfile.Write(absDir, harnessRoot, lf); err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "✓ updated %s → %s (%s)\n",
		name, mf.Version, resolved.ResolvedSHA[:displaySHALen(resolved.ResolvedSHA)])

	res, verr := verifyPolicies(absDir, harnessRoot)
	if verr != nil {
		return fmt.Errorf("policy verify: %w", verr)
	}
	if printVerifyReport(absDir, res) {
		return fmt.Errorf("policy %q updated but strict cascade is violated — resolve the shadowing file(s) above", name)
	}
	return nil
}

// dirtyFiles compares the on-disk state of namespaceDir against the per-file
// hashes recorded in expected, returning a sorted list of "<path> (state)"
// entries for any file that differs. State is one of: modified, added, removed.
func dirtyFiles(installDir, namespaceDir string, expected map[string]string) ([]string, error) {
	current, err := lockfile.HashFilesUnder(installDir, namespaceDir)
	if err != nil {
		return nil, err
	}
	var diffs []string
	for path, want := range expected {
		got, found := current[path]
		if !found {
			diffs = append(diffs, path+" (removed)")
			continue
		}
		if got != want {
			diffs = append(diffs, path+" (modified)")
		}
	}
	for path := range current {
		if _, found := expected[path]; !found {
			diffs = append(diffs, path+" (added)")
		}
	}
	sort.Strings(diffs)
	return diffs, nil
}

// displaySHALen returns 8 if sha is at least 8 chars long, otherwise its full
// length — guards the SHA-truncation in user-facing output against short refs.
func displaySHALen(sha string) int {
	if len(sha) >= 8 {
		return 8
	}
	return len(sha)
}

func printPolicyUpdateUsage(w *os.File) {
	fmt.Fprint(w, `keystone policy update — update an installed org policy

Usage:
  keystone policy update <name> [<new-ref>] [--dir <path>] [--harness-root <name>] [--force]

Re-resolves the policy using the ref recorded in <harness-root>/keystone.lock.json,
or the new ref if supplied. For a moving ref (e.g. a branch like #main) this
picks up new commits; for a pinned tag it's a no-op unless the tag was moved.
Refuses to overwrite files that have been edited since install — pass --force
to discard local changes.

Flags:
  --dir <path>           Directory containing the harness (defaults to cwd).
  --harness-root <name>  Harness directory name (defaults to "harness").
  --force                Discard local edits to policy files when updating.
`)
}
