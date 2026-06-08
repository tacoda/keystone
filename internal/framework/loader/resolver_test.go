package loader

import (
	"strings"
	"testing"
)

func TestParsePolicyRef(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		wantURL   string
		wantRev   string
		wantError string // substring; empty means no error expected
	}{
		{
			name:    "https no rev",
			raw:     "git+https://github.com/acme/policy.git",
			wantURL: "https://github.com/acme/policy.git",
		},
		{
			name:    "https with tag",
			raw:     "git+https://github.com/acme/policy.git#v1.2.0",
			wantURL: "https://github.com/acme/policy.git",
			wantRev: "v1.2.0",
		},
		{
			name:    "ssh with branch",
			raw:     "git+ssh://git@github.com/acme/policy.git#main",
			wantURL: "ssh://git@github.com/acme/policy.git",
			wantRev: "main",
		},
		{
			name:      "empty",
			raw:       "",
			wantError: "empty policy ref",
		},
		{
			name:      "unsupported scheme",
			raw:       "https://github.com/acme/policy.git",
			wantError: "unsupported policy ref",
		},
		{
			name:      "git+ prefix only",
			raw:       "git+",
			wantError: "empty URL",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref, err := ParsePolicyRef(tt.raw)
			if tt.wantError != "" {
				if err == nil {
					t.Errorf("ParsePolicyRef(%q) expected error containing %q, got nil", tt.raw, tt.wantError)
					return
				}
				if !strings.Contains(err.Error(), tt.wantError) {
					t.Errorf("ParsePolicyRef(%q) error %q does not contain %q", tt.raw, err.Error(), tt.wantError)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParsePolicyRef(%q) unexpected error: %v", tt.raw, err)
			}
			if ref.URL != tt.wantURL {
				t.Errorf("ParsePolicyRef(%q).URL = %q, want %q", tt.raw, ref.URL, tt.wantURL)
			}
			if ref.Rev != tt.wantRev {
				t.Errorf("ParsePolicyRef(%q).Rev = %q, want %q", tt.raw, ref.Rev, tt.wantRev)
			}
			if ref.Raw != tt.raw {
				t.Errorf("ParsePolicyRef(%q).Raw = %q, want %q", tt.raw, ref.Raw, tt.raw)
			}
			if ref.Scheme != "git" {
				t.Errorf("ParsePolicyRef(%q).Scheme = %q, want %q", tt.raw, ref.Scheme, "git")
			}
		})
	}
}
