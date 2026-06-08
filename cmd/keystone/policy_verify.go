package main

import (
	"fmt"
	"os"

	"github.com/tacoda/keystone/internal/framework/loader"
	"github.com/tacoda/keystone/internal/framework/lockfile"
)

// verifyPolicies reads the lockfile at installDir and walks the cascade,
// returning the combined result.
func verifyPolicies(installDir string) (*loader.VerifyResult, error) {
	lf, err := lockfile.Read(installDir)
	if err != nil {
		return nil, err
	}
	return loader.Verify(installDir, lf.Policies)
}

// printVerifyReport renders a result to stdout. The bool return is whether
// any hard violations were found (gaps are advisory and do not flip it).
func printVerifyReport(installDir string, res *loader.VerifyResult) bool {
	if !res.HasErrors() && !res.HasGaps() {
		fmt.Fprintf(os.Stdout, "✓ policy verify clean — no strict items overridden, no required items unmet (%s)\n", installDir)
		return false
	}
	if res.HasErrors() {
		fmt.Fprintf(os.Stdout, "✗ policy verify found %d strict violation(s) in %s\n\n", len(res.Violations), installDir)
		for _, v := range res.Violations {
			fmt.Fprintln(os.Stdout, "  "+v.String())
			fmt.Fprintln(os.Stdout)
		}
		fmt.Fprintln(os.Stdout, "Strict policy items cannot be overridden by lower tiers. Remove the offending file(s) or raise the change with the policy author.")
		if res.HasGaps() {
			fmt.Fprintln(os.Stdout)
		}
	}
	if res.HasGaps() {
		fmt.Fprintf(os.Stdout, "? policy verify found %d required item(s) that need to be defined in %s\n\n", len(res.Gaps), installDir)
		for _, g := range res.Gaps {
			fmt.Fprintln(os.Stdout, "  "+g.String())
		}
		fmt.Fprintln(os.Stdout)
		fmt.Fprintln(os.Stdout, "These items are declared `required` by a policy but no tier (project, team, or org) has defined them. Add them at the project level.")
	}
	return res.HasErrors()
}
