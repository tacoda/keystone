package mcp

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

func TestListFilter_Kind(t *testing.T) {
	hook := primitive.Primitive{Frontmatter: primitive.Frontmatter{Kind: "hook", ID: "lint"}}
	if !(listFilter{kind: "hook"}).match(hook) || (listFilter{kind: "skill"}).match(hook) {
		t.Error("kind filter: a hook matches hook, not skill")
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
		if !isCurrentKind(g.verb) {
			t.Errorf("keystone_new_%s: not a current kind", g.verb)
		}
	}
}

// isCurrentKind: verb is a current kind name. Corpus is deliberately
// both a current kind (6.0) and a 4.x alias for reference in frontmatter,
// so the check is KnownKinds membership; a 4.x-only name (guide, sensor,
// subagent) is not in KnownKinds.
func isCurrentKind(verb string) bool {
	return primitive.IsKnownKind(verb)
}
