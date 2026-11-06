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
// Shared by `keystone project`, init, watch, and migrate so they never drift.
func projectHosts(absDir string, primitives []primitive.Primitive, w io.Writer) (projection.ApplyResult, error) {
	if w == nil {
		w = io.Discard
	}
	outs, warns, mode := planHosts(absDir, primitives)
	lf, err := lockfile.Read(absDir)
	if err != nil {
		return projection.ApplyResult{}, err
	}
	res, err := projection.Apply(absDir, outs, lf.Projections, mode)
	if err != nil {
		return res, err
	}
	lf.Projections = res.Owned
	if err := lockfile.Write(absDir, lf); err != nil {
		return res, err
	}
	for _, d := range res.Wrote {
		fmt.Fprintf(w, "  wrote: %s\n", d)
	}
	for _, d := range res.Pruned {
		fmt.Fprintf(w, "  pruned: %s\n", d)
	}
	for _, x := range warns {
		fmt.Fprintf(w, "  ⚠ %s: %s\n", x.Host, x.Message)
	}
	return res, nil
}

// planHosts builds the projection plan for the project's enabled hosts and
// its configured link mode.
func planHosts(absDir string, primitives []primitive.Primitive) ([]projection.Output, []projection.Warning, projection.LinkMode) {
	cfg, err := config.ReadProjectConfig(absDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		fmt.Fprintf(os.Stderr, "keystone: read keystone.json: %v\n", err)
	}
	linkMode := ""
	if cfg != nil {
		linkMode = cfg.LinkMode
	}
	outs, warns := projection.Plan(primitives, projection.Enabled())
	return outs, warns, projection.ResolveLinkMode(linkMode)
}
