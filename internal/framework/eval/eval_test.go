package eval

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

// TestHookCommand — an eval runs a hook's run:, the command hook fire runs;
// only a 4.x sensor without run: falls back to its first code block.
func TestHookCommand(t *testing.T) {
	dir := t.TempDir()
	body := "---\nid: lint\n---\n```bash\nlegacy-cmd\n```\n"
	if err := os.WriteFile(filepath.Join(dir, "lint.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	withRun := primitive.Primitive{Frontmatter: primitive.Frontmatter{Run: "go vet ./..."}, Path: "lint.md"}
	if got := hookCommand(dir, withRun); got != "go vet ./..." {
		t.Errorf("with run: got %q", got)
	}
	legacy := primitive.Primitive{Path: "lint.md"}
	if got := hookCommand(dir, legacy); got != "legacy-cmd" {
		t.Errorf("4.x fallback: got %q", got)
	}
}
