package migrations

// Version 2.4 — read-surface parity. The framework's source-of-truth
// is unchanged: `.keystone/harness/*.md` files. What 2.4 adds is
// equal-shape READ access from two surfaces:
//
//   - MCP server: new `keystone_show` tool + `tags:` filter on
//     `keystone_list_primitives`. The agent (inside a session) now
//     queries the composed primitive graph the same way `keystone
//     show` exposes it on the CLI.
//   - Web dashboard: `/harness/primitives` gains tag-cloud + tag
//     dropdown filter; the detail page renders `tags:`, `includes:`,
//     `model:`, `host_triggers:`, and reverse `included_by:` (via
//     extended `primitive.IncomingRefs`). The human (outside the
//     session) sees the same graph the agent does.
//
// Both surfaces are READ-only over the file source of truth. No
// schema changes; no data transforms; no new on-disk artifacts.
//
// Migration is a no-op by design. The Up plan is empty — `keystone
// migrate up` records 2.4 in the lockfile so subsequent index runs
// stop reporting it as pending, but writes nothing else.

func init() {
	Register(Migration{
		Version: "2.4",
		Up:      planUp_2_4,
		Down:    planDown_2_4,
	})
}

func planUp_2_4(absDir string) (*Plan, error) {
	// No-op. 2.4 is exclusively additive at the binary layer (MCP +
	// web). Existing 2.3 file content remains valid; the new read
	// surfaces parse what's already there.
	return &Plan{}, nil
}

func planDown_2_4(absDir string) (*Plan, error) {
	// No-op. Nothing to undo on disk.
	return &Plan{}, nil
}
