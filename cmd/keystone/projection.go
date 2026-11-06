package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/lockfile"
	"github.com/tacoda/keystone/internal/framework/primitive"
	"github.com/tacoda/keystone/internal/framework/projection"
)

// projectHosts plans and applies every host projection for the walked
// primitives, records ownership in the lockfile, and logs to w (nil = quiet).
// overwrite lets it replace paths keystone did not create (--overwrite).
// Shared by `keystone project`, init, watch, and migrate so they never drift.
func projectHosts(absDir string, primitives []primitive.Primitive, w io.Writer, overwrite bool) (projection.ApplyResult, error) {
	if w == nil {
		w = io.Discard
	}
	lf, err := lockfile.Read(absDir)
	if err != nil {
		return projection.ApplyResult{}, err
	}
	pr, mode, err := planHosts(absDir, primitives, lf.ManagedEntries)
	if err != nil {
		return projection.ApplyResult{}, err
	}
	res, applyErr := projection.Apply(absDir, pr.Outputs, lf.Projections, projection.Options{Mode: mode, Overwrite: overwrite})
	if err := saveProjectionState(absDir, lf, res, applyErr); err != nil {
		return res, err
	}
	if applyErr != nil {
		return res, applyErr
	}
	printProjection(w, res, pr)
	return res, nil
}

// saveProjectionState records what Apply owns and manages in the lockfile.
// Merged entries are recorded even when a later write failed: an entry
// keystone added but did not record would look like the user's and never
// leave.
func saveProjectionState(absDir string, lf *lockfile.Lockfile, res projection.ApplyResult, applyErr error) error {
	if applyErr == nil {
		lf.Projections = res.Owned
	}
	if applyErr != nil && len(res.Managed) == 0 {
		return nil
	}
	recordManaged(lf, res.Managed)
	return lockfile.Write(absDir, lf)
}

func recordManaged(lf *lockfile.Lockfile, managed map[string][]string) {
	for dest, entries := range managed {
		if lf.ManagedEntries == nil {
			lf.ManagedEntries = map[string][]string{}
		}
		lf.ManagedEntries[dest] = entries
	}
}

// printProjection logs what changed, every constraint a host cannot
// enforce, and the plan's warnings.
func printProjection(w io.Writer, res projection.ApplyResult, pr projection.PlanResult) {
	for _, d := range res.Wrote {
		fmt.Fprintf(w, "  wrote: %s\n", d)
	}
	for _, d := range res.Pruned {
		fmt.Fprintf(w, "  pruned: %s\n", d)
	}
	for _, o := range pr.Outputs {
		if o.Strategy == projection.Report {
			fmt.Fprintf(w, "  ⚠ %s cannot enforce %ss — %s\n", o.Host, o.Kind, projection.ReportAdvice(o.Kind))
			if o.Note != "" {
				fmt.Fprintf(w, "      %s\n", o.Note)
			}
		}
	}
	for _, x := range pr.Warnings {
		fmt.Fprintf(w, "  ⚠ %s: %s\n", x.Host, x.Message)
	}
}

// planHosts builds the projection plan for the project's enabled hosts
// (keystone.json `adapters`) and its configured link mode.
func planHosts(absDir string, primitives []primitive.Primitive, managed map[string][]string) (projection.PlanResult, projection.LinkMode, error) {
	cfg, err := config.ReadProjectConfig(absDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		fmt.Fprintf(os.Stderr, "keystone: read keystone.json: %v\n", err)
	}
	var linkMode string
	var adapters []string
	if cfg != nil {
		linkMode, adapters = cfg.LinkMode, cfg.Adapters
	}
	hosts, err := projection.HostsFor(absDir, adapters)
	if err != nil {
		return projection.PlanResult{}, "", err
	}
	return projection.PlanWithState(primitives, hosts, managed), projection.ResolveLinkMode(linkMode), nil
}
