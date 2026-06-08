package main

import (
	"github.com/tacoda/keystone/internal/framework/lockfile"
)

// ensureLockfile loads the lockfile and backfills its keystone section from
// INSTALL_PROFILE.md if the lockfile is empty (i.e., the install was created
// before the lockfile existed in 0.x). The returned lockfile is not yet
// persisted — callers that mutate it must call lockfile.Write.
//
// This is a transitional helper for the 0.x → 1.0 path. At 1.0 the
// INSTALL_PROFILE.md fallback can be dropped.
func ensureLockfile(installDir string) (*lockfile.Lockfile, error) {
	lf, err := lockfile.Read(installDir)
	if err != nil {
		return nil, err
	}
	if lf.Keystone.Version != "" {
		return lf, nil
	}
	if v, perr := readKeystoneVersionFromProfile(installDir); perr == nil && v != "" {
		lf.Keystone.Version = v
	}
	if agents, perr := readInstalledAgentsFromProfile(installDir); perr == nil && len(agents) > 0 {
		lf.Keystone.Agents = agents
	}
	return lf, nil
}
