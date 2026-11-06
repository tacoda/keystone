package agnostic

import (
	"strings"
	"testing"
)

// TestRenderPointer_NonImportHosts pins the frozen 4.x pointer the v2.x
// migrations still render through the frozen Cursor adapter.
func TestRenderPointer_NonImportHosts(t *testing.T) {
	cursor := RenderPointer(CursorProfile())
	if strings.Contains(cursor, "@CHARTER.md") {
		t.Error("cursor has no import mechanism; pointer must not use @import")
	}
	for _, want := range []string{"**No subagents**", "You **must** read"} {
		if !strings.Contains(cursor, want) {
			t.Errorf("cursor pointer missing %q", want)
		}
	}
}
