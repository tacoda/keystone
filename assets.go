// Package keystone holds the embedded asset filesystem.
//
// The //go:embed directive requires the embedded directories to be siblings
// or descendants of the Go source file holding the directive. Because the
// embedded content (harness/, targets/, optional/, migrations/) lives at the
// repository root, this package must live at the root too. The rest of the
// binary lives under cmd/keystone/ and imports Assets from here.
//
// Phase 2 of the 1.0 plan will move the embedded content into
// internal/framework/scaffold/templates/, at which point this root package
// can fold into the framework runtime.
package keystone

import "embed"

//go:embed all:harness all:targets all:optional all:migrations
var Assets embed.FS
