package patch

import (
	"testing"
	"testing/fstest"
)

func TestLooksLikeVersion(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"0.1.0", true},
		{"1.2.3", true},
		{"0", true},
		{"", false},
		{".", false},
		{"0.", false},
		{".1", false},
		{"1.x", false},
		{"templates", false},
		{"README.md", false},
	}
	for _, tt := range tests {
		if got := looksLikeVersion(tt.in); got != tt.want {
			t.Errorf("looksLikeVersion(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestCompareSemver(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"0.1.0", "0.2.0", -1},
		{"0.2.0", "0.1.0", 1},
		{"0.1.0", "0.1.0", 0},
		{"1.0.0", "0.99.0", 1},
		{"0.10.0", "0.9.0", 1}, // numeric, not lexical
		{"0.1", "0.1.0", 0},    // shorter pads with zero
	}
	for _, tt := range tests {
		if got := compareSemver(tt.a, tt.b); got != tt.want {
			t.Errorf("compareSemver(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestLoad_FiltersByVersionAndExtension(t *testing.T) {
	assets := fstest.MapFS{
		"patches/0.1.0/001-first.json": &fstest.MapFile{Data: []byte(`{
"description": "first",
"operations": [{"type": "add_file", "path": "a.md", "content": "A"}]
}`)},
		"patches/0.2.0/001-second.json": &fstest.MapFile{Data: []byte(`{
"description": "second",
"operations": []
}`)},
		"patches/0.2.0/skipped.yaml": &fstest.MapFile{Data: []byte("ignored: yes")},
		"patches/templates/x.json":   &fstest.MapFile{Data: []byte(`{"description":"x","operations":[]}`)},
		"patches/README.md":          &fstest.MapFile{Data: []byte("docs")},
	}

	got, err := Load(assets, "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d patches, want 2: %+v", len(got), got)
	}
	if got[0].Version != "0.1.0" || got[0].ID != "001-first" {
		t.Errorf("got[0] = %+v, want version 0.1.0 / id 001-first", got[0])
	}
	if got[1].Version != "0.2.0" || got[1].ID != "001-second" {
		t.Errorf("got[1] = %+v, want version 0.2.0 / id 001-second", got[1])
	}
}

func TestLoad_FromVersionExcludesLowerAndEqual(t *testing.T) {
	assets := fstest.MapFS{
		"patches/0.1.0/001-a.json": &fstest.MapFile{Data: []byte(`{"description":"a","operations":[]}`)},
		"patches/0.2.0/001-b.json": &fstest.MapFile{Data: []byte(`{"description":"b","operations":[]}`)},
		"patches/0.3.0/001-c.json": &fstest.MapFile{Data: []byte(`{"description":"c","operations":[]}`)},
	}

	got, err := Load(assets, "0.2.0")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d patches, want 1", len(got))
	}
	if got[0].Version != "0.3.0" {
		t.Errorf("got[0].Version = %q, want 0.3.0", got[0].Version)
	}
}
