package charter

import (
	"strings"
	"testing"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

func TestExplain_ActivationByKind(t *testing.T) {
	cases := []struct {
		p    primitive.Primitive
		want string // substring expected in Activation
	}{
		{primitive.Primitive{Frontmatter: primitive.Frontmatter{Kind: "guide", ID: "g", Globs: []string{"cmd/**/*.go"}}}, "cmd/**/*.go"},
		{primitive.Primitive{Frontmatter: primitive.Frontmatter{Kind: "sensor", ID: "s", Mode: "computational", Event: "Stop", Run: "go build"}}, "computational check"},
		{primitive.Primitive{Frontmatter: primitive.Frontmatter{Kind: "sensor", ID: "r", Mode: "inferential", Event: "on-gate", Returns: "verdict"}}, "inferential check"},
		{primitive.Primitive{Frontmatter: primitive.Frontmatter{Kind: "tool", ID: "t", Transport: "http", Event: "on-commit"}}, "side-effect on on-commit"},
		{primitive.Primitive{Frontmatter: primitive.Frontmatter{Kind: "tool", ID: "t2", Transport: "cli"}}, "on demand"},
		{primitive.Primitive{Frontmatter: primitive.Frontmatter{Kind: "skill", ID: "k", Triggers: []string{"do x"}}}, "auto-activates on triggers"},
		{primitive.Primitive{Frontmatter: primitive.Frontmatter{Kind: "agent", ID: "a"}}, "subagent"},
		{primitive.Primitive{Frontmatter: primitive.Frontmatter{Kind: "command", ID: "c"}}, "unit of work"},
	}
	for _, c := range cases {
		e := Explain(c.p)
		if !strings.Contains(e.Activation, c.want) {
			t.Errorf("Explain(%s).Activation = %q, want substring %q", c.p.Kind, e.Activation, c.want)
		}
	}
}

func TestExplain_Links(t *testing.T) {
	e := Explain(primitive.Primitive{Frontmatter: primitive.Frontmatter{
		Kind: "guide", ID: "g", Corpus: []string{"x"}, Deps: []string{"y"},
	}})
	joined := strings.Join(e.Links, " ")
	if !strings.Contains(joined, "corpus: x") || !strings.Contains(joined, "deps: y") {
		t.Errorf("links = %v", e.Links)
	}
}
