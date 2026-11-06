package toolrun

import (
	"context"
	"strings"
	"testing"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

var echo = primitive.Primitive{Frontmatter: primitive.Frontmatter{
	Kind: "tool", ID: "echo", Run: `printf '%s|%s' "$KEYSTONE_ARG_TARGET" "$KEYSTONE_ARG_MODE"`,
	Args: []primitive.Arg{{Name: "target", Required: true}, {Name: "mode"}},
}}

func TestRun_ArgsBecomeEnv(t *testing.T) {
	out, err := Run(context.Background(), t.TempDir(), echo, map[string]string{"target": "a b"})
	if err != nil || out != "a b|" {
		t.Errorf("out = %q, err = %v", out, err)
	}
}

func TestRun_RejectsBadArgs(t *testing.T) {
	for args, want := range map[string]map[string]string{
		"missing required argument --target": {"mode": "x"},
		"unknown argument --nope":            {"target": "x", "nope": "y"},
	} {
		if _, err := Run(context.Background(), t.TempDir(), echo, want); err == nil || !strings.Contains(err.Error(), args) {
			t.Errorf("want %q, got %v", args, err)
		}
	}
}
