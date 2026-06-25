package mcp

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

func TestToolMCPName(t *testing.T) {
	cases := map[string]string{
		"grep-symbols":    "keystone_tool_grep_symbols",
		"keystone:fmt":    "keystone_tool_keystone_fmt",
		"a/b/c":           "keystone_tool_a_b_c",
	}
	for id, want := range cases {
		if got := toolMCPName(id); got != want {
			t.Errorf("toolMCPName(%q) = %q, want %q", id, got, want)
		}
	}
}

// TestExecTool — the handler runs the tool's run: script with declared args
// exported as KEYSTONE_ARG_<NAME>.
func TestExecTool(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"target": "hello"}
	p := primitive.Primitive{Frontmatter: primitive.Frontmatter{
		Kind: "tool", ID: "echo", Run: `printf '%s' "$KEYSTONE_ARG_TARGET"`,
		Args: []primitive.Arg{{Name: "target"}},
	}}
	out, err := execTool(context.Background(), t.TempDir(), p, req)
	if err != nil {
		t.Fatalf("execTool: %v\n%s", err, out)
	}
	if strings.TrimSpace(out) != "hello" {
		t.Errorf("execTool output = %q, want %q", out, "hello")
	}
}

// TestExecTool_NonZeroSurfaces — a failing run: returns an error.
func TestExecTool_NonZeroSurfaces(t *testing.T) {
	req := mcp.CallToolRequest{}
	p := primitive.Primitive{Frontmatter: primitive.Frontmatter{Kind: "tool", ID: "boom", Run: "exit 3"}}
	if _, err := execTool(context.Background(), t.TempDir(), p, req); err == nil {
		t.Error("expected non-zero run: to surface an error")
	}
}
