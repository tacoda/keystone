package charter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

func prim(kind, id, prov string) primitive.Primitive {
	return primitive.Primitive{
		Frontmatter: primitive.Frontmatter{Kind: kind, ID: id, Description: "d"},
		Provenance:  prov,
	}
}

func TestEffective_ProjectWinsRecordsShadow(t *testing.T) {
	entries := Effective([]primitive.Primitive{
		prim("guide", "idioms/go/x", "policy/acme"),
		prim("guide", "idioms/go/x", "project"),
		prim("guide", "idioms/go/y", "policy/acme"),
	})
	if len(entries) != 2 {
		t.Fatalf("expected 2 deduped entries, got %d", len(entries))
	}
	var x *Entry
	for i := range entries {
		if entries[i].Primitive.ID == "idioms/go/x" {
			x = &entries[i]
		}
	}
	if x == nil || x.Primitive.Provenance != "project" {
		t.Fatalf("project should win for x: %+v", x)
	}
	if len(x.Shadows) != 1 || x.Shadows[0] != "policy/acme" {
		t.Errorf("expected policy/acme shadowed, got %v", x.Shadows)
	}
}

func TestSignals_BuiltinsPlusCustomSortedDeduped(t *testing.T) {
	got := Signals([]string{"on-deploy", "pre-verify"}) // pre-verify is also a builtin
	seen := map[string]int{}
	for _, s := range got {
		seen[s]++
	}
	if seen["pre-verify"] != 1 {
		t.Errorf("pre-verify should appear once, got %d", seen["pre-verify"])
	}
	if seen["on-deploy"] != 1 {
		t.Error("custom signal on-deploy missing")
	}
	// sorted
	for i := 1; i < len(got); i++ {
		if got[i-1] > got[i] {
			t.Errorf("signals not sorted at %d: %q > %q", i, got[i-1], got[i])
		}
	}
}

func TestCoverage_GovernedVsUncharted(t *testing.T) {
	tmp := t.TempDir()
	// a guide governing **/*.go
	write(t, tmp, ".charter/guides/idioms/go/x.md",
		"---\nkind: guide\nid: idioms/go/x\ndescription: d\nglobs:\n  - \"**/*.go\"\n---\nbody\n")
	write(t, tmp, "cmd/main.go", "package main\n")
	write(t, tmp, "docs/readme.md", "# doc\n") // uncharted (no guide matches *.md)

	res, err := Coverage(tmp, ".charter")
	if err != nil {
		t.Fatal(err)
	}
	if res.Governed < 1 {
		t.Errorf("expected cmd/main.go governed, got %+v", res)
	}
	foundDoc := false
	for _, u := range res.Uncharted {
		if u == "docs/readme.md" {
			foundDoc = true
		}
	}
	if !foundDoc {
		t.Errorf("docs/readme.md should be uncharted, got %v", res.Uncharted)
	}
}

func write(t *testing.T, root, rel, body string) {
	t.Helper()
	abs := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(abs, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
