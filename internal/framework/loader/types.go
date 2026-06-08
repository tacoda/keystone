// Package loader resolves <port>/<name> against a cascade of project +
// plugin layers. The cascade is the ordered set of content sources the
// framework walks at session start: the project's harness/ comes first, then
// each plugin in pre-order tree-walk order from keystone.json's plugins[].
//
// The Loader interface is the framework's read-side abstraction over the
// cascade. Phase 3 will wire it into the runtime once vendored plugins land;
// for now it is a defined stub with in-memory fixture tests so the contract
// is locked before the full implementation is built.
package loader

import "io/fs"

// Plugin represents one source layer in the cascade. Root is the content
// root of that layer — harness/ for the project layer, the vendored plugin
// directory for plugin layers.
type Plugin struct {
	Name string // "project" for the project layer; the plugin's name otherwise
	Root fs.FS
}

// Cascade is the ordered set of plugin layers, with the project at index 0
// followed by plugins in pre-order tree-walk order. The Loader resolves
// against Project first, then walks Plugins in slice order.
type Cascade struct {
	Project Plugin
	Plugins []Plugin
}

// Origin describes which layer in the cascade won a resolution.
type Origin struct {
	Plugin string // plugin name, or "project" for the project layer
	Path   string // relative path within that plugin's Root
}
