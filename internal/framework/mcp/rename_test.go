package mcp

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

func TestListFilter_Category(t *testing.T) {
	hook := primitive.Primitive{Frontmatter: primitive.Frontmatter{Kind: "hook", ID: "lint"}}
	if !(listFilter{category: "constraint"}).match(hook) || (listFilter{category: "capability"}).match(hook) {
		t.Error("category filter: a hook is a constraint")
	}
	if !(listFilter{kind: "sensor"}).match(hook) {
		t.Error("a 4.x kind name still filters")
	}
}

// TestLoadReference_BothIDForms — a corpus: entry finds a 5.0 reference id
// and a 4.x one that carries the corpus/ prefix.
func TestLoadReference_BothIDForms(t *testing.T) {
	root := t.TempDir()
	for rel, id := range map[string]string{"go/errors.md": "go/errors", "domain/shape.md": "corpus/domain/shape"} {
		path := filepath.Join(root, ".agents/references", rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("---\nid: "+id+"\ndescription: x\n---\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, ref := range []string{"go/errors", "domain/shape", "corpus/domain/shape"} {
		if _, _, err := loadReference(root, ref); err != nil {
			t.Errorf("corpus: %s did not resolve: %v", ref, err)
		}
	}
}

// TestWriteGenerators_Use5Kinds — every keystone_new_* verb is a 5.0 kind
// (the CLI refuses the 4.x names) or a keystone-only scaffold.
func TestWriteGenerators_Use5Kinds(t *testing.T) {
	for _, g := range writeGenerators() {
		if g.verb == "adapter" || g.verb == "policy" {
			continue
		}
		if !is5Kind(g.verb) {
			t.Errorf("keystone_new_%s: not a 5.0 kind", g.verb)
		}
	}
}

// is5Kind: verb is a 5.0 kind name itself, not a 4.x alias of one.
func is5Kind(verb string) bool {
	k, legacy := primitive.NormalizeKind(verb)
	if legacy || string(k) != verb {
		return false
	}
	return primitive.CategoryOf(k) != ""
}
