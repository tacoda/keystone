package config

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseShorthand(t *testing.T) {
	tests := []struct {
		spec        string
		wantSource  string
		wantVersion string
	}{
		{"tacoda/tacoda-org@0.2.0", "tacoda/tacoda-org", "0.2.0"},
		{"tacoda/tacoda-org", "tacoda/tacoda-org", ""},
		{"github.com/acme/policies@v1.0", "github.com/acme/policies", "v1.0"},
		{"gitlab.com/acme/policies@main", "gitlab.com/acme/policies", "main"},
		{"acme/policies@abc1234", "acme/policies", "abc1234"},
		{"", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.spec, func(t *testing.T) {
			gotSource, gotVersion := ParseShorthand(tt.spec)
			if gotSource != tt.wantSource || gotVersion != tt.wantVersion {
				t.Errorf("ParseShorthand(%q) = (%q, %q), want (%q, %q)",
					tt.spec, gotSource, gotVersion, tt.wantSource, tt.wantVersion)
			}
		})
	}
}

func TestDefaultPolicyName(t *testing.T) {
	tests := []struct {
		source string
		want   string
	}{
		{"tacoda/tacoda-org", "tacoda-org"},
		{"github.com/acme/policies", "policies"},
		{"gitlab.com/acme/policies/", "policies"},
		{"single", "single"},
	}
	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			if got := DefaultPolicyName(tt.source); got != tt.want {
				t.Errorf("DefaultPolicyName(%q) = %q, want %q", tt.source, got, tt.want)
			}
		})
	}
}

func TestExpandSource(t *testing.T) {
	tests := []struct {
		source string
		want   string
	}{
		{"tacoda/tacoda-org", "https://github.com/tacoda/tacoda-org.git"},
		{"github.com/tacoda/tacoda-org", "https://github.com/tacoda/tacoda-org.git"},
		{"gitlab.com/acme/policies", "https://gitlab.com/acme/policies.git"},
		{"github.com/tacoda/tacoda-org.git", "https://github.com/tacoda/tacoda-org.git"},
	}
	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			if got := ExpandSource(tt.source); got != tt.want {
				t.Errorf("ExpandSource(%q) = %q, want %q", tt.source, got, tt.want)
			}
		})
	}
}

func TestValidateSource(t *testing.T) {
	tests := []struct {
		source    string
		wantError string // substring; empty = no error
	}{
		{"tacoda/tacoda-org", ""},
		{"github.com/acme/policies", ""},
		{"gitlab.com/acme/team/policies", ""},
		{"", "source is empty"},
		{"git+https://github.com/acme/policies.git", "0.x-style URL"},
		{"https://github.com/acme/policies.git", "0.x-style URL"},
		{"acme", "does not match"},
		{"Acme/Policies", "does not match"},
	}
	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			err := ValidateSource(tt.source)
			if tt.wantError == "" {
				if err != nil {
					t.Errorf("ValidateSource(%q) unexpected error: %v", tt.source, err)
				}
				return
			}
			if err == nil {
				t.Errorf("ValidateSource(%q) expected error containing %q, got nil", tt.source, tt.wantError)
				return
			}
			if !strings.Contains(err.Error(), tt.wantError) {
				t.Errorf("ValidateSource(%q) error %q does not contain %q", tt.source, err.Error(), tt.wantError)
			}
		})
	}
}

// TestProjectConfig_ResolvedHarnessRoot pins 2.0's fixed-layout
// contract: the harness path is no longer per-project, and the
// deprecated HarnessRoot field is ignored when present in keystone.json.
func TestProjectConfig_ResolvedHarnessRoot(t *testing.T) {
	cases := []struct {
		name string
		cfg  *ProjectConfig
	}{
		{"empty", &ProjectConfig{}},
		{"legacy-field-ignored", &ProjectConfig{HarnessRoot: "playbook"}},
		{"nil", nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.cfg.ResolvedHarnessRoot(); got != DefaultHarnessRoot {
				t.Errorf("ResolvedHarnessRoot() = %q, want %q (fixed at 2.0)", got, DefaultHarnessRoot)
			}
		})
	}
}

func TestReadProjectConfig_Missing(t *testing.T) {
	dir := t.TempDir()
	_, err := ReadProjectConfig(dir)
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("ReadProjectConfig missing should return os.ErrNotExist, got %v", err)
	}
}

func TestWriteReadProjectConfig_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	in := &ProjectConfig{
		Version:          SchemaVersion,
		FrameworkVersion: "1.0.0",
		HarnessRoot:      "playbook",
		Policies: []PolicyNode{
			{
				Name:    "acme-org",
				Source:  "github.com/acme/policies",
				Version: "v2.0.0",
				Strict:  map[string][]string{"guides": {"data-handling"}},
				Children: []PolicyNode{
					{
						Name:    "acme-platform",
						Source:  "acme/platform-policies",
						Version: "v1.4.0",
					},
				},
			},
		},
		Budgets: map[string]BudgetSpec{
			"guides": {MaxTokens: 6000},
		},
	}
	if err := WriteProjectConfig(dir, in); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ProjectConfigFile)); err != nil {
		t.Fatalf("expected file at %s: %v", ProjectConfigFile, err)
	}

	out, err := ReadProjectConfig(dir)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("config differs:\n  in:  %+v\n  out: %+v", in, out)
	}
}

func TestProjectConfig_RejectsDuplicateNames(t *testing.T) {
	cfg := &ProjectConfig{
		Version: SchemaVersion,
		Policies: []PolicyNode{
			{Name: "a", Source: "x/a", Version: "v1"},
			{Name: "a", Source: "y/a", Version: "v2"},
		},
	}
	err := cfg.validate()
	if err == nil || !strings.Contains(err.Error(), "duplicate plugin name") {
		t.Errorf("expected duplicate-name error, got %v", err)
	}
}

func TestProjectConfig_RejectsDuplicateNamesAcrossDepth(t *testing.T) {
	cfg := &ProjectConfig{
		Version: SchemaVersion,
		Policies: []PolicyNode{
			{
				Name: "a", Source: "x/a", Version: "v1",
				Children: []PolicyNode{
					{Name: "a", Source: "y/a", Version: "v2"},
				},
			},
		},
	}
	err := cfg.validate()
	if err == nil || !strings.Contains(err.Error(), "duplicate plugin name") {
		t.Errorf("expected duplicate-name error across depth, got %v", err)
	}
}
