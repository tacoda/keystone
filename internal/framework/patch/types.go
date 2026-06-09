// Package patch loads and applies framework patches. Each patch is a JSON
// file under patches/<version>/<NNN>-<name>.json declaring a list of
// operations against the consumer's harness/ tree or keystone config files.
//
// Scope at 1.0 covers two kinds of forward-only change: config-schema bumps
// (keystone.json, lockfile) and updates to framework-scaffolded scaffold
// prose (READMEs under harness/, sensor and action playbooks scaffolded by
// keystone init). The rules content a user wrote — guide bodies, corpus
// entries, domain invariants, custom adapters — is owned by the user's git
// and stays out of patches. The 0.x notion of patches as wholesale content
// migrations is dead. See docs/ports/patch.md for the bright line.
package patch

// Patch is one loaded patch file. ID is the filename without extension;
// Version is the parent directory name (e.g. "0.6.0").
type Patch struct {
	Version    string `json:"-"`
	ID         string `json:"-"`
	SourcePath string `json:"-"`

	Description string      `json:"description"`
	Operations  []Operation `json:"operations"`
}

// Operation is the raw shape of a single op in a patch file. Op-specific
// fields are union-style; only the ones matching Type are populated.
type Operation struct {
	Type string `json:"type"`
	Path string `json:"path"`

	// add_file
	Content string `json:"content,omitempty"`

	// frontmatter_set
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`

	// ensure_section
	AfterHeading string `json:"after_heading,omitempty"`
	Heading      string `json:"heading,omitempty"`
	Body         string `json:"body,omitempty"`

	// replace_block (reuses Heading)
	Match       string `json:"match,omitempty"`
	Replacement string `json:"replacement,omitempty"`

	// move_dir: relocate every file under Path (source) into To (destination),
	// preserving subpath structure. Idempotent — already-moved files no-op;
	// destination files with diverged content surface as conflicts. After all
	// files are moved, the source directory is removed if empty.
	//
	// move_file: relocate a single file from Path to To. Same idempotency
	// semantics as move_dir, scoped to one file.
	//
	// delete_dir: remove Path if it is empty (after a prior move_dir, for
	// instance). Conflicts if Path still contains files.
	//
	// delete_file: remove the single file at Path. Idempotent — missing
	// target no-ops.
	To string `json:"to,omitempty"`
}

// OpStatus classifies the planned effect of a single operation.
type OpStatus int

const (
	OpNoop     OpStatus = iota // target state already present; skip silently
	OpCreate                   // file doesn't exist; will be written
	OpChange                   // file exists; will be modified
	OpConflict                 // can't apply automatically — user must intervene
)

// OpResult is what PlanOperation returns: the read-out of current state, the
// proposed new state (nil for no-op/conflict), a status, and a human-readable
// note used for user-facing messages.
type OpResult struct {
	Op         Operation
	TargetPath string // absolute path on disk (single-file ops); empty for multi-file
	Status     OpStatus
	Current    []byte
	Proposed   []byte
	Note       string

	// MovePlans is populated only by move_dir; one entry per file the op will
	// touch. Aggregate status above reflects the worst sub-status across these
	// (any conflict → conflict; all noop → noop; otherwise change).
	MovePlans []MovePlan
}

// MovePlan is one file-level action inside a move_dir operation.
type MovePlan struct {
	SourceAbs string
	DestAbs   string
	RelPath   string // for display
	Status    OpStatus
	Note      string
}
