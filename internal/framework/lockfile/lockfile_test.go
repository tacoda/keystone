package lockfile

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/tacoda/keystone/internal/framework/manifest"
)

func TestPolicyLock_ResolvedTier(t *testing.T) {
	tests := []struct {
		name string
		tier string
		want string
	}{
		{"default empty", "", manifest.TierOrg},
		{"explicit org", manifest.TierOrg, manifest.TierOrg},
		{"explicit team", manifest.TierTeam, manifest.TierTeam},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PolicyLock{Tier: tt.tier}
			if got := p.ResolvedTier(); got != tt.want {
				t.Errorf("ResolvedTier() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRead_MissingReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	lf, err := Read(dir)
	if err != nil {
		t.Fatalf("Read on missing file: %v", err)
	}
	if lf == nil {
		t.Fatal("Read returned nil lockfile")
	}
	if lf.Version != Version {
		t.Errorf("Version = %d, want %d", lf.Version, Version)
	}
	if lf.Policies == nil {
		t.Error("Policies map is nil, want empty map")
	}
}

func TestWriteRead_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	in := &Lockfile{
		Keystone: KeystoneInfo{
			Version:   "0.15.0",
			Installed: "2026-06-08",
			Agents:    []string{"claude-code", "codex"},
		},
		Policies: map[string]PolicyLock{
			"acme": {
				SourceRef:       "git+https://github.com/acme/policy.git#v1.0.0",
				ResolvedSHA:     "deadbeef00000000000000000000000000000000",
				PolicyVersion:   "1.0.0",
				KeystoneVersion: "0.15.0",
				Tier:            manifest.TierOrg,
				Strict:          manifest.StrictSpec{Guides: []string{"data-handling"}},
				Required:        manifest.StrictSpec{Actions: []string{"release"}},
				Files:           map[string]string{"harness/policies/acme/guides/x.md": "sha256:abc"},
			},
		},
	}
	if err := Write(dir, in); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, File)); err != nil {
		t.Fatalf("expected file at %s: %v", File, err)
	}

	out, err := Read(dir)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !reflect.DeepEqual(in.Keystone, out.Keystone) {
		t.Errorf("Keystone differs:\n  in:  %+v\n  out: %+v", in.Keystone, out.Keystone)
	}
	if !reflect.DeepEqual(in.Policies, out.Policies) {
		t.Errorf("Policies differ:\n  in:  %+v\n  out: %+v", in.Policies, out.Policies)
	}
}

func TestHashFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	h, err := HashFile(path)
	if err != nil {
		t.Fatalf("HashFile: %v", err)
	}
	want := "sha256:2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if h != want {
		t.Errorf("HashFile = %q, want %q", h, want)
	}
}

func TestHashFilesUnder(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sub, "a.md"), []byte("A"), 0o644); err != nil {
		t.Fatalf("write a: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sub, "b.md"), []byte("B"), 0o644); err != nil {
		t.Fatalf("write b: %v", err)
	}

	got, err := HashFilesUnder(dir, "sub")
	if err != nil {
		t.Fatalf("HashFilesUnder: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("got %d files, want 2: %+v", len(got), got)
	}
	if _, ok := got["sub/a.md"]; !ok {
		t.Errorf("missing sub/a.md in %+v", got)
	}
	if _, ok := got["sub/b.md"]; !ok {
		t.Errorf("missing sub/b.md in %+v", got)
	}
}

func TestHashFilesUnder_MissingDir(t *testing.T) {
	dir := t.TempDir()
	got, err := HashFilesUnder(dir, "missing")
	if err != nil {
		t.Errorf("HashFilesUnder on missing dir returned error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %d files for missing dir, want 0", len(got))
	}
}
