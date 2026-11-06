package primitive

import (
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Agent Skills spec limits (https://agentskills.io/specification).
const (
	maxSkillName          = 64
	maxSkillDescription   = 1024
	maxSkillCompatibility = 500
)

// specSkillKeys are the only top-level keys the spec allows; the reference
// validator (`skills-ref validate`) rejects any other.
var specSkillKeys = map[string]bool{
	"name": true, "description": true, "license": true,
	"compatibility": true, "metadata": true, "allowed-tools": true,
}

// keystoneSkillKeys are the 4.x keys `SpecSkillFrontmatter` knows how to
// carry into spec shape. Any other non-spec key blocks the rewrite rather
// than being dropped.
var keystoneSkillKeys = map[string]bool{
	"kind": true, "id": true, "triggers": true, "tools": true,
	"model": true, "tags": true, "includes": true,
}

var reSkillName = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// lintSkillSpec checks a SKILL.md against the Agent Skills spec. Spec
// violations are errors; 4.x keys are a warning, since keystone still reads
// them (migrations iron law 6) but strict hosts may not.
func lintSkillSpec(p Primitive, add func(FindingSeverity, string)) {
	if path.Base(p.Path) != "SKILL.md" {
		return
	}
	if sev, msg := skillNameFinding(p); msg != "" {
		add(sev, msg)
	}
	if n := len(p.Description); n > maxSkillDescription {
		add(FindingError, fmt.Sprintf("skill description is %d characters (spec max %d)", n, maxSkillDescription))
	}
	if n := len(p.Compatibility); n > maxSkillCompatibility {
		add(FindingError, fmt.Sprintf("skill compatibility is %d characters (spec max %d)", n, maxSkillCompatibility))
	}
	if extra := keysOutside(p.Keys, specSkillKeys); len(extra) > 0 {
		add(FindingWarning, fmt.Sprintf("keys outside the Agent Skills spec (%s) — strict hosts reject them; `keystone lint --fix` moves them", strings.Join(extra, ", ")))
	}
}

// skillNameFinding checks the spec's `name` rules against the skill's
// directory; msg is "" when valid.
func skillNameFinding(p Primitive) (FindingSeverity, string) {
	name, dir := p.Name, path.Base(path.Dir(p.Path))
	switch {
	case name == "":
		return FindingWarning, fmt.Sprintf("Agent Skills spec: add `name: %s` (must match the directory) — `keystone lint --fix` does it", dir)
	case len(name) > maxSkillName || !reSkillName.MatchString(name):
		return FindingError, fmt.Sprintf("skill name %q invalid: lowercase letters, digits, and single hyphens, max %d", name, maxSkillName)
	case name != dir:
		return FindingError, fmt.Sprintf("skill name %q must match its directory %q", name, dir)
	}
	return "", ""
}

// keysOutside returns the keys not allowed by any of the given sets, sorted.
func keysOutside(keys []string, allowed ...map[string]bool) []string {
	var out []string
	for _, k := range keys {
		if !inAny(k, allowed) {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

func inAny(k string, sets []map[string]bool) bool {
	for _, set := range sets {
		if set[k] {
			return true
		}
	}
	return false
}

// NeedsSpecFix reports whether a skill's frontmatter is not yet in spec shape.
func NeedsSpecFix(p Primitive) bool {
	return Kind(p.Kind) == KindSkill && path.Base(p.Path) == "SKILL.md" &&
		(p.Name == "" || len(keysOutside(p.Keys, specSkillKeys)) > 0)
}

// specSkill is the spec frontmatter, in the order the spec documents it.
type specSkill struct {
	Name          string            `yaml:"name"`
	Description   string            `yaml:"description"`
	License       string            `yaml:"license,omitempty"`
	Compatibility string            `yaml:"compatibility,omitempty"`
	AllowedTools  string            `yaml:"allowed-tools,omitempty"`
	Metadata      map[string]string `yaml:"metadata,omitempty"`
}

// SpecSkillFrontmatter renders p's frontmatter in Agent Skills spec shape:
// name from the directory, triggers folded into the description (or into
// metadata when that would pass the 1024 limit), composed tools (so retired
// concerns are inlined) as allowed-tools, model and tags into metadata.
// It refuses a skill carrying keys it cannot carry, rather than drop them.
func SpecSkillFrontmatter(p Primitive) (string, error) {
	if unknown := keysOutside(p.Keys, specSkillKeys, keystoneSkillKeys); len(unknown) > 0 {
		return "", fmt.Errorf("%s: cannot carry %s into spec shape — move by hand", p.Path, strings.Join(unknown, ", "))
	}
	out := specSkill{
		Name:          path.Base(path.Dir(p.Path)),
		Description:   p.Description,
		License:       p.License,
		Compatibility: p.Compatibility,
		AllowedTools:  strings.Join(p.Tools, " "),
		Metadata:      specMetadata(p),
	}
	if len(p.Triggers) > 0 {
		folded := strings.TrimSpace(p.Description) + " Triggers: " + strings.Join(p.Triggers, "; ") + "."
		if len(folded) <= maxSkillDescription {
			out.Description = folded
		} else {
			out.Metadata["triggers"] = strings.Join(p.Triggers, "; ")
		}
	}
	if len(out.Metadata) == 0 {
		out.Metadata = nil
	}
	var b strings.Builder
	enc := yaml.NewEncoder(&b)
	enc.SetIndent(2) // match the repo's frontmatter style
	if err := enc.Encode(out); err != nil {
		return "", err
	}
	return b.String(), enc.Close()
}

func specMetadata(p Primitive) map[string]string {
	md := map[string]string{}
	for k, v := range p.Metadata {
		md[k] = v
	}
	if p.Model != "" {
		md["model"] = p.Model
	}
	if len(p.Tags) > 0 {
		md["tags"] = strings.Join(p.Tags, ",")
	}
	return md
}

// InlineIncludes rewrites raw frontmatter fm so it carries the composed list
// fields directly and drops `includes:` — retiring concerns (ADR 0010)
// without changing what the primitive resolves to. Other keys keep their
// order and values.
func InlineIncludes(fm string, composed Primitive) (string, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(fm), &doc); err != nil {
		return "", err
	}
	if len(doc.Content) == 0 || doc.Content[0].Kind != yaml.MappingNode {
		return "", fmt.Errorf("frontmatter is not a mapping")
	}
	m := doc.Content[0]
	pathsKey := "paths"
	if mapIndex(m, "globs") >= 0 {
		pathsKey = "globs"
	}
	for key, vals := range map[string][]string{
		pathsKey: composed.Paths, "triggers": composed.Triggers, "tools": composed.Tools,
		"tags": composed.Tags, "corpus": composed.Corpus, "deps": composed.Deps,
	} {
		setList(m, key, vals)
	}
	if i := mapIndex(m, "includes"); i >= 0 {
		m.Content = append(m.Content[:i], m.Content[i+2:]...)
	}
	var b strings.Builder
	enc := yaml.NewEncoder(&b)
	enc.SetIndent(2)
	if err := enc.Encode(m); err != nil {
		return "", err
	}
	return b.String(), enc.Close()
}

// mapIndex returns the index of key's key-node in mapping m, or -1.
func mapIndex(m *yaml.Node, key string) int {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return i
		}
	}
	return -1
}

// setList sets key to vals (a block sequence), appending the key when
// absent. Empty vals leave the mapping untouched.
func setList(m *yaml.Node, key string, vals []string) {
	if len(vals) == 0 {
		return
	}
	seq := &yaml.Node{Kind: yaml.SequenceNode}
	for _, v := range vals {
		seq.Content = append(seq.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: v})
	}
	if i := mapIndex(m, key); i >= 0 {
		m.Content[i+1] = seq
		return
	}
	m.Content = append(m.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: key}, seq)
}
