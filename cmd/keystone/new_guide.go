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
	corpusRel := filepath.Join("references", topic, name+".md")
	rulePath := filepath.Join(projectDir, charterRoot, ruleRel)
	corpusPath := filepath.Join(projectDir, charterRoot, corpusRel)

	title := titleize(name)
	id := topic + "/" + name
	ruleBody := fmt.Sprintf(`---
kind: rule
id: %s
description: TODO — one-line description of what this rule governs.
# paths:            # apply only when a touched file matches; omit to apply everywhere
#   - "src/**/*.go"
corpus:
  - %s
---

# %s

One-sentence framing of what this rule governs.

## Rules

- Directive one.
- Directive two.

For reasoning, see [`+"`%s`"+`](%s).
`, id, id, title, filepath.ToSlash(corpusRel), filepath.ToSlash(corpusRel))

	if err := writeSkeleton(rulePath, ruleBody); err != nil {
		return err
	}
	return writeSkeleton(corpusPath, referenceBody(id, title, ruleRel))
}

// referenceBody is a reference skeleton: the reasoning behind its rule.
func referenceBody(id, title, ruleRel string) string {
	return fmt.Sprintf(`---
kind: reference
id: %s
description: TODO — one-line description of the reasoning captured here.
---

# %s — reasoning

Long-form explanation of why the paired rule exists.

## Anti-patterns

Failure modes the rule guards against.

## Sources

Source material — papers, books, posts.

Back to the rule: [`+"`%s`"+`](%s).
`, id, title, filepath.ToSlash(ruleRel), filepath.ToSlash(ruleRel))
}

// runNewReference handles `keystone new reference <topic>/<name>`: only the
// reference, for when the rule already exists or the reasoning stands alone.
func runNewReference(args []string) error {
	projectDir, charterRoot, remaining, err := parseDirAndCharterRoot(args)
	if err != nil {
		return err
	}
	if len(remaining) != 1 {
		return fmt.Errorf("`keystone new reference` requires exactly one argument: <topic>/<name>")
	}
	topic, name, err := splitTopicName(remaining[0])
	if err != nil {
		return err
	}
	path := filepath.Join(projectDir, charterRoot, "references", topic, name+".md")
	ruleRel := filepath.Join("rules", topic, name+".md")
	return writeSkeleton(path, referenceBody(topic+"/"+name, titleize(name), ruleRel))
}

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
