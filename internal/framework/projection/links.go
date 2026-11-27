package projection

import (
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	mdLink    = regexp.MustCompile(`\]\(([^)\s]+)\)`)
	urlScheme = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9+.-]*:`)
)

// RebaseLinks rewrites each relative Markdown link in body, written for a
// file at src, so it resolves from dest instead (both repo-relative). A
// shim lives in another folder than its source; without this its links
// point at files that aren't there. URLs, anchors, absolute paths, links
// that leave the repo, and code blocks are left as they are.
func RebaseLinks(body []byte, src, dest string) []byte {
	lines := strings.SplitAfter(string(body), "\n")
	inFence := false
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inFence = !inFence
			continue
		}
		if !inFence {
			lines[i] = mdLink.ReplaceAllStringFunc(line, func(m string) string {
				return "](" + rebase(m[2:len(m)-1], src, dest) + ")"
			})
		}
	}
	return []byte(strings.Join(lines, ""))
}

// IsRelativeLink reports a link target that names a file relative to the
// linking file: not empty, not an anchor, not absolute, not a URL.
func IsRelativeLink(target string) bool {
	switch {
	case target == "", strings.HasPrefix(target, "#"), strings.HasPrefix(target, "/"):
		return false
	}
	return !urlScheme.MatchString(target)
}

// RelativeLinkTargets returns each relative Markdown link target in body
// (written for a file at repo-relative src), resolved to a repo-relative
// path. Code fences are skipped; URLs, anchors, absolute paths, and links
// that leave the repo are not returned.
func RelativeLinkTargets(body []byte, src string) []string {
	var out []string
	seen := map[string]bool{}
	lines := strings.SplitAfter(string(body), "\n")
	inFence := false
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		for _, m := range mdLink.FindAllStringSubmatch(line, -1) {
			target := m[1]
			if !IsRelativeLink(target) {
				continue
			}
			file, _, _ := strings.Cut(target, "#")
			abs := path.Clean(path.Join(path.Dir(src), file))
			if abs == ".." || strings.HasPrefix(abs, "../") {
				continue // leaves the repo
			}
			if !seen[abs] {
				seen[abs] = true
				out = append(out, filepath.ToSlash(abs))
			}
		}
	}
	return out
}

// rebase moves one link target from src's folder to dest's.
func rebase(target, src, dest string) string {
	if !IsRelativeLink(target) {
		return target
	}
	file, frag, _ := strings.Cut(target, "#")
	abs := path.Clean(path.Join(path.Dir(src), file))
	if abs == ".." || strings.HasPrefix(abs, "../") {
		return target // leaves the repo
	}
	rel, err := filepath.Rel(filepath.FromSlash(path.Dir(dest)), filepath.FromSlash(abs))
	if err != nil {
		return target
	}
	out := filepath.ToSlash(rel)
	if frag != "" {
		out += "#" + frag
	}
	return out
}
