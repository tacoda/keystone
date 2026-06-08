package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// runNewGuide handles `keystone new guide <topic>/<name>`. Scaffolds a
// guide and a paired corpus stub, both linking each other via the
// harness-root-relative path convention.
func runNewGuide(args []string) error {
	projectDir, harnessRoot, remaining, err := parseDirAndHarnessRoot(args)
	if err != nil {
		return err
	}
	if len(remaining) != 1 {
		return fmt.Errorf("`keystone new guide` requires exactly one argument: <topic>/<name>")
	}
	topic, name, err := splitTopicName(remaining[0])
	if err != nil {
		return err
	}

	guideRel := filepath.Join("guides", topic, name+".md")
	corpusRel := filepath.Join("corpus", topic, name+".md")
	guidePath := filepath.Join(projectDir, harnessRoot, guideRel)
	corpusPath := filepath.Join(projectDir, harnessRoot, corpusRel)

	title := titleize(name)
	guideBody := fmt.Sprintf(`# %s — rules

One-sentence framing of what this guide governs.

- Non-negotiable rule one.
- Non-negotiable rule two.
- Strongly-preferred rule three.

For reasoning, see [`+"`%s`"+`](%s).
`, title, filepath.ToSlash(corpusRel), filepath.ToSlash(corpusRel))

	corpusBody := fmt.Sprintf(`# %s — reasoning

Long-form explanation of why the rules in the paired guide exist.

## Anti-patterns

Failure modes the rules guard against.

## References

Source material — papers, books, posts.

Back to the rules: [`+"`%s`"+`](%s).
`, title, filepath.ToSlash(guideRel), filepath.ToSlash(guideRel))

	if err := writeSkeleton(guidePath, guideBody); err != nil {
		return err
	}
	if err := writeSkeleton(corpusPath, corpusBody); err != nil {
		return err
	}
	return nil
}

// runNewCorpus handles `keystone new corpus <topic>/<name>`. Scaffolds
// only the corpus file; useful when the guide already exists or the
// corpus stands alone (rare).
func runNewCorpus(args []string) error {
	projectDir, harnessRoot, remaining, err := parseDirAndHarnessRoot(args)
	if err != nil {
		return err
	}
	if len(remaining) != 1 {
		return fmt.Errorf("`keystone new corpus` requires exactly one argument: <topic>/<name>")
	}
	topic, name, err := splitTopicName(remaining[0])
	if err != nil {
		return err
	}
	corpusRel := filepath.Join("corpus", topic, name+".md")
	guideRel := filepath.Join("guides", topic, name+".md")
	corpusPath := filepath.Join(projectDir, harnessRoot, corpusRel)
	title := titleize(name)
	body := fmt.Sprintf(`# %s — reasoning

Long-form explanation.

## Anti-patterns

Failure modes.

## References

Source material.

Back to the rules: [`+"`%s`"+`](%s).
`, title, filepath.ToSlash(guideRel), filepath.ToSlash(guideRel))
	return writeSkeleton(corpusPath, body)
}

// splitTopicName parses a `<topic>/<name>` argument.
func splitTopicName(s string) (string, string, error) {
	i := strings.IndexByte(s, '/')
	if i < 0 {
		return "", "", fmt.Errorf("expected <topic>/<name>, got %q", s)
	}
	topic, name := s[:i], s[i+1:]
	if topic == "" || name == "" {
		return "", "", fmt.Errorf("topic and name must both be non-empty in %q", s)
	}
	if strings.ContainsAny(name, "/\\") {
		return "", "", fmt.Errorf("name %q must not contain a slash", name)
	}
	return topic, name, nil
}

// titleize converts a kebab-case name into a Title Case heading.
// `data-handling` → `Data handling`.
func titleize(name string) string {
	parts := strings.Split(name, "-")
	if len(parts) == 0 {
		return name
	}
	if len(parts[0]) > 0 {
		parts[0] = strings.ToUpper(parts[0][:1]) + parts[0][1:]
	}
	return strings.Join(parts, " ")
}
