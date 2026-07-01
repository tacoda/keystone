package primitive

import "testing"

func TestMatchPath(t *testing.T) {
	cases := []struct {
		pat, path string
		want      bool
	}{
		{"cmd/**/*.go", "cmd/keystone/root.go", true},
		{"cmd/**/*.go", "cmd/keystone/sub/deep.go", true},
		{"cmd/**/*.go", "cmd/keystone/notes.md", false}, // suffix must match
		{"cmd/**/*.go", "internal/x.go", false},         // wrong prefix
		{"internal/**", "internal/framework/web/sse.go", true},
		{"internal/**", "internal", true}, // trailing ** matches zero remaining segments
		{"go.mod", "go.mod", true},
		{"go.mod", "go.sum", false},
		{"**/*_test.go", "internal/framework/primitive/glob_test.go", true},
		{"**/*_test.go", "internal/framework/primitive/glob.go", false},
		{"*.go", "root.go", true},
		{"*.go", "cmd/root.go", false}, // * does not cross /
		{".charter/**/*.md", ".charter/guides/idioms/go/stdlib-first.md", true},
	}
	for _, c := range cases {
		if got := MatchPath(c.pat, c.path); got != c.want {
			t.Errorf("MatchPath(%q, %q) = %v, want %v", c.pat, c.path, got, c.want)
		}
	}
}
