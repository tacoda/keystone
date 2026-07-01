package primitive

import (
	"path/filepath"
	"strings"
)

// MatchPath reports whether a project-relative POSIX path matches a glob
// that may contain `**` (matches any number of path segments, including
// zero), plus `*` and `?` within a single segment (via filepath.Match).
//
// This is the doublestar semantics guides declare in `globs:` — e.g.
// `cmd/**/*.go` matches `cmd/keystone/root.go`, and `internal/**`
// matches everything under internal/. Unlike filepath.Match alone, `*`
// never crosses a `/`.
func MatchPath(pattern, path string) bool {
	pattern = strings.TrimPrefix(pattern, "./")
	path = strings.TrimPrefix(path, "./")
	return matchSegments(strings.Split(pattern, "/"), strings.Split(path, "/"))
}

func matchSegments(pat, name []string) bool {
	for len(pat) > 0 {
		if pat[0] == "**" {
			return matchDoubleStar(pat[1:], name)
		}
		if len(name) == 0 {
			return false
		}
		if ok, _ := filepath.Match(pat[0], name[0]); !ok {
			return false
		}
		pat, name = pat[1:], name[1:]
	}
	return len(name) == 0
}

// matchDoubleStar matches `rest` against `name` where a `**` preceded
// `rest`: the `**` consumes zero or more leading segments of name.
func matchDoubleStar(rest, name []string) bool {
	if len(rest) == 0 {
		return true // trailing ** matches any remainder (incl. none)
	}
	for i := 0; i <= len(name); i++ {
		if matchSegments(rest, name[i:]) {
			return true
		}
	}
	return false
}
