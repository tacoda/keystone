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
	pr, mode := planHosts(absDir, primitives)
	lf, err := lockfile.Read(absDir)
	if err != nil {
		return projection.ApplyResult{}, err
	}
	res, err := projection.Apply(absDir, pr.Outputs, lf.Projections, projection.Options{Mode: mode, Overwrite: overwrite})
	if err != nil {
		return res, err
	}
	lf.Projections = res.Owned
	if err := lockfile.Write(absDir, lf); err != nil {
		return res, err
	}
	printProjection(w, res, pr)
	return res, nil
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
		}
	}
	for _, x := range pr.Warnings {
		fmt.Fprintf(w, "  ⚠ %s: %s\n", x.Host, x.Message)
	}
}

// planHosts builds the projection plan for the project's enabled hosts
// (keystone.json `adapters`) and its configured link mode.
func planHosts(absDir string, primitives []primitive.Primitive) (projection.PlanResult, projection.LinkMode) {
	cfg, err := config.ReadProjectConfig(absDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		fmt.Fprintf(os.Stderr, "keystone: read keystone.json: %v\n", err)
	}
	var linkMode string
	var adapters []string
	if cfg != nil {
		linkMode, adapters = cfg.LinkMode, cfg.Adapters
	}
	return projection.Plan(primitives, projection.Enabled(adapters)), projection.ResolveLinkMode(linkMode)
}
