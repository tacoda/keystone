// Package loader resolves <port>/<name> against a cascade of project +
// policy layers. The project's harness/ always wins by default; among
// policies, policies nested deeper in keystone.json's policies[] tree refine
// the outer policies they're nested in. A policy's strict declarations lock
// items absolutely — nothing else can override them.
//
// The Loader interface is the framework's read-side abstraction over the
// cascade. Phase 3 will wire it into the runtime once vendored policies land;
// for now it is a defined stub with in-memory fixture tests so the contract
// is locked before the full implementation is built.
package loader

import "io/fs"

// Policy represents one source layer in the cascade. Root is the content
// root of that layer — harness/ for the project layer, the vendored policy
// directory for policy layers.
type Policy struct {
	Name string // "project" for the project layer; the policy's name otherwise
	Root fs.FS
}

// Cascade is the ordered set of policy layers, with the project at index 0
// followed by policies in pre-order tree-walk order (outer policies before
// their children). The Loader resolves against Project first; among policies,
// deeper-nested policies refine the outer policies they're nested in.
type Cascade struct {
	Project Policy
	Policies []Policy
}

// Origin describes which layer in the cascade won a resolution.
type Origin struct {
	Policy string // policy name, or "project" for the project layer
	Path   string // relative path within that policy's Root
}
