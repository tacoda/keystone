package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// runNewRule handles `keystone new rule <topic>/<name>`. Scaffolds a
// rule (a glob-scoped directive) and a paired corpus stub the rule
// cites via `corpus:`.
func runNewRule(args []string) error {
	projectDir, charterRoot, remaining, err := parseDirAndCharterRoot(args)
	if err != nil {
		return err
	}
	if len(remaining) != 1 {
		return fmt.Errorf("`keystone new rule` requires exactly one argument: <topic>/<name>")
	}
	topic, name, err := splitTopicName(remaining[0])
	if err != nil {
		return err
	}

	ruleRel := filepath.Join("rules", topic, name+".md")
	corpusRel := filepath.Join("corpus", topic, name+".md")
	rulePath := filepath.Join(projectDir, charterRoot, ruleRel)
	corpusPath := filepath.Join(projectDir, charterRoot, corpusRel)

	title := titleize(name)
	id := "rules/" + topic + "/" + name
	ruleBody := fmt.Sprintf(`---
kind: rule
id: %s
description: TODO — one-line description of what this rule governs.
severity: must
corpus:
  - %s
---

# %s — rules

One-sentence framing of what this rule governs.

## RULES

- Directive one.
- Directive two.

For reasoning, see [`+"`%s`"+`](%s).
`, id, "corpus/"+topic+"/"+name, title, filepath.ToSlash(corpusRel), filepath.ToSlash(corpusRel))

	corpusBody := fmt.Sprintf(`---
kind: corpus
id: corpus/%s
description: TODO — one-line description of the reasoning captured here.
---

# %s — reasoning

Long-form explanation of why the rules in the paired guide exist.

## Anti-patterns

Failure modes the rules guard against.

## References

Source material — papers, books, posts.

Back to the rules: [`+"`%s`"+`](%s).
`, id, title, filepath.ToSlash(ruleRel), filepath.ToSlash(ruleRel))

	if err := writeSkeleton(rulePath, ruleBody); err != nil {
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
	projectDir, charterRoot, remaining, err := parseDirAndCharterRoot(args)
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
	ruleRel := filepath.Join("rules", topic, name+".md")
	corpusPath := filepath.Join(projectDir, charterRoot, corpusRel)
	title := titleize(name)
	body := fmt.Sprintf(`# %s — reasoning

Long-form explanation.

## Anti-patterns

Failure modes.

## References

Source material.

Back to the rules: [`+"`%s`"+`](%s).
`, title, filepath.ToSlash(ruleRel), filepath.ToSlash(ruleRel))
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
