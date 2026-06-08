package manifest

import (
	"strings"
	"testing"
)

func TestStrictSpec_IsEmpty(t *testing.T) {
	tests := []struct {
		name string
		s    StrictSpec
		want bool
	}{
		{"empty", StrictSpec{}, true},
		{"only guides", StrictSpec{Guides: []string{"a"}}, false},
		{"only playbooks", StrictSpec{Playbooks: []string{"a"}}, false},
		{"only actions", StrictSpec{Actions: []string{"a"}}, false},
		{"only sensors", StrictSpec{Sensors: []string{"a"}}, false},
		{"mixed", StrictSpec{Guides: []string{"a"}, Sensors: []string{"b"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.IsEmpty(); got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManifest_Namespace(t *testing.T) {
	m := &Manifest{Name: "acme-platform"}
	if got := m.Namespace(); got != "acme-platform" {
		t.Errorf("Namespace() = %q, want %q", got, "acme-platform")
	}
}

func TestManifest_validate(t *testing.T) {
	tests := []struct {
		name      string
		manifest  Manifest
		wantError string // substring; empty means no error expected
	}{
		{
			name:     "valid minimal",
			manifest: Manifest{Name: "acme", Version: "1.0.0"},
		},
		{
			name:     "valid with strict guides",
			manifest: Manifest{Name: "acme", Version: "1.0.0", Strict: StrictSpec{Guides: []string{"data-handling"}}},
		},
		{
			name:     "valid with strict sensors",
			manifest: Manifest{Name: "acme", Version: "1.0.0", Strict: StrictSpec{Sensors: []string{"rubocop"}}},
		},
		{
			name:      "missing name",
			manifest:  Manifest{Version: "1.0.0"},
			wantError: "missing required field 'name'",
		},
		{
			name:      "invalid name with uppercase",
			manifest:  Manifest{Name: "Acme", Version: "1.0.0"},
			wantError: "must match",
		},
		{
			name:      "invalid name starting with digit",
			manifest:  Manifest{Name: "1acme", Version: "1.0.0"},
			wantError: "must match",
		},
		{
			name:      "missing version",
			manifest:  Manifest{Name: "acme"},
			wantError: "missing required field 'version'",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.manifest.validate()
			switch {
			case tt.wantError == "" && err != nil:
				t.Errorf("validate() unexpected error: %v", err)
			case tt.wantError != "" && err == nil:
				t.Errorf("validate() expected error containing %q, got nil", tt.wantError)
			case tt.wantError != "" && err != nil && !strings.Contains(err.Error(), tt.wantError):
				t.Errorf("validate() error %q does not contain %q", err.Error(), tt.wantError)
			}
		})
	}
}
