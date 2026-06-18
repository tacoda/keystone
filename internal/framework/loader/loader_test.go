package loader

import (
	"errors"
	"io"
	"io/fs"
	"testing"
	"testing/fstest"
)

// makeFS is a tiny constructor for in-memory fixtures.
func makeFS(files map[string]string) fs.FS {
	m := fstest.MapFS{}
	for path, content := range files {
		m[path] = &fstest.MapFile{Data: []byte(content)}
	}
	return m
}

func TestDefaultLoader_Resolve(t *testing.T) {
	project := makeFS(map[string]string{
		"guides/process/spec.md":     "project version of spec",
		"guides/process/orient.md":   "project version of orient",
		"sensors/build.md":           "project sensor",
		"corpus/principles/tdd.md":   "project tdd corpus",
	})
	universal := makeFS(map[string]string{
		"guides/process/spec.md":      "universal version of spec",
		"guides/principles/bdd.md":    "universal bdd guide",
		"corpus/principles/tdd.md":    "universal tdd corpus",
		"corpus/principles/bdd.md":    "universal bdd corpus",
	})
	team := makeFS(map[string]string{
		"sensors/rubocop.md":         "team rubocop sensor",
		"guides/principles/bdd.md":   "team bdd guide override attempt",
	})

	cascade := Cascade{
		Project: Policy{Name: "project", Root: project},
		Policies: []Policy{
			{Name: "universal", Root: universal},
			{Name: "team", Root: team},
		},
	}
	l := New(cascade)

	tests := []struct {
		name        string
		port        string
		item        string
		wantOrigin  string
		wantContent string
		wantErr     error
	}{
		{
			name:        "project wins over policy for same item",
			port:        "guides/process",
			item:        "spec",
			wantOrigin:  "project",
			wantContent: "project version of spec",
		},
		{
			name:        "policy provides item not in project",
			port:        "guides/principles",
			item:        "bdd",
			wantOrigin:  "universal",
			wantContent: "universal bdd guide",
		},
		{
			name:        "first policy wins over later policies for same item",
			port:        "guides/principles",
			item:        "bdd",
			wantOrigin:  "universal",
			wantContent: "universal bdd guide",
		},
		{
			name:        "later policy provides item missing from earlier",
			port:        "sensors",
			item:        "rubocop",
			wantOrigin:  "team",
			wantContent: "team rubocop sensor",
		},
		{
			name:        "project-only item resolves to project",
			port:        "guides/process",
			item:        "orient",
			wantOrigin:  "project",
			wantContent: "project version of orient",
		},
		{
			name:    "not found anywhere returns ErrNotExist",
			port:    "guides/process",
			item:    "release",
			wantErr: fs.ErrNotExist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, origin, err := l.Resolve(tt.port, tt.item)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Resolve(%q, %q) error = %v, want %v", tt.port, tt.item, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Resolve(%q, %q) unexpected error: %v", tt.port, tt.item, err)
			}
			defer f.Close()

			if origin.Policy != tt.wantOrigin {
				t.Errorf("Resolve(%q, %q) origin = %q, want %q", tt.port, tt.item, origin.Policy, tt.wantOrigin)
			}
			got, err := io.ReadAll(f)
			if err != nil {
				t.Fatalf("read file: %v", err)
			}
			if string(got) != tt.wantContent {
				t.Errorf("Resolve(%q, %q) content = %q, want %q", tt.port, tt.item, got, tt.wantContent)
			}
		})
	}
}

func TestDefaultLoader_NilRoot(t *testing.T) {
	// Nil project root with a populated policy should still resolve.
	policy := makeFS(map[string]string{
		"guides/process/spec.md": "policy spec",
	})
	cascade := Cascade{
		Project: Policy{Name: "project", Root: nil},
		Policies: []Policy{{Name: "p1", Root: policy}},
	}
	l := New(cascade)
	f, origin, err := l.Resolve("guides/process", "spec")
	if err != nil {
		t.Fatalf("Resolve unexpected error: %v", err)
	}
	defer f.Close()
	if origin.Policy != "p1" {
		t.Errorf("origin = %q, want %q", origin.Policy, "p1")
	}
}

func TestDefaultLoader_EmptyCascade(t *testing.T) {
	l := New(Cascade{})
	_, _, err := l.Resolve("guides/process", "spec")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("empty cascade Resolve err = %v, want %v", err, fs.ErrNotExist)
	}
}
