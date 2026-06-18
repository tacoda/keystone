package loader

import (
	"io/fs"
)

// Loader resolves <port>/<name> against a Cascade. Phase 1 ships the
// interface and a default project-first / policies-in-order implementation;
// Phase 3 extends it with strict-cascade enforcement and depth limits.
type Loader interface {
	// Resolve returns the winning file for a given port and name, or
	// fs.ErrNotExist if no layer provides it. The returned Origin describes
	// which layer won.
	//
	// `port` is the port directory (e.g., "guides/process" or "sensors").
	// `name` is the bare item name without the .md extension.
	Resolve(port, name string) (fs.File, Origin, error)
}

// New returns the default Loader for the given Cascade. The default loader
// walks Project first, then each Policy in slice order; the first layer
// whose Root contains <port>/<name>.md wins. The framework never composes
// overlapping files.
func New(c Cascade) Loader {
	return &defaultLoader{cascade: c}
}

type defaultLoader struct {
	cascade Cascade
}

func (l *defaultLoader) Resolve(port, name string) (fs.File, Origin, error) {
	rel := port + "/" + name + ".md"

	if l.cascade.Project.Root != nil {
		if f, err := l.cascade.Project.Root.Open(rel); err == nil {
			return f, Origin{Policy: l.cascade.Project.Name, Path: rel}, nil
		}
	}
	for _, p := range l.cascade.Policies {
		if p.Root == nil {
			continue
		}
		if f, err := p.Root.Open(rel); err == nil {
			return f, Origin{Policy: p.Name, Path: rel}, nil
		}
	}
	return nil, Origin{}, fs.ErrNotExist
}
