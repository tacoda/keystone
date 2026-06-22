package primitive

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// SplitFrontmatter returns the YAML between the leading `---` fences and
// the body that follows. ok=false means there is no frontmatter; the
// caller decides whether that's an error for the path in question.
//
// Mirrors the helper in internal/framework/patch — duplicated rather
// than shared to keep the patch package's surface frozen.
func SplitFrontmatter(s string) (fm, body string, ok bool) {
	if !strings.HasPrefix(s, "---\n") && !strings.HasPrefix(s, "---\r\n") {
		return "", s, false
	}
	rest := strings.TrimPrefix(strings.TrimPrefix(s, "---\r\n"), "---\n")
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return "", s, false
	}
	fm = rest[:end+1]
	tail := rest[end+len("\n---"):]
	tail = strings.TrimPrefix(strings.TrimPrefix(tail, "\r\n"), "\n")
	return fm, tail, true
}

// Parse extracts a Frontmatter from the YAML block between the `---`
// fences of a primitive file. Returns ok=false when the file has no
// frontmatter at all (legal during migration; the indexer skips it).
//
// Args may appear as either a list of strings (sugar) or a list of
// {name,type,required,description} mappings. The flexible decoding is
// handled by Frontmatter's UnmarshalYAML.
func Parse(fileContents string) (Frontmatter, bool, error) {
	fm, _, ok := SplitFrontmatter(fileContents)
	if !ok {
		return Frontmatter{}, false, nil
	}
	var out Frontmatter
	if err := yaml.Unmarshal([]byte(fm), &out); err != nil {
		return Frontmatter{}, true, fmt.Errorf("parse frontmatter: %w", err)
	}
	return out, true, nil
}

// UnmarshalYAML lets args: be authored as either a plain string list
// ("- foo") or full Arg mappings. The string form lifts to
// Arg{Name: <s>} so authors don't have to spell out every field for the
// common case.
func (f *Frontmatter) UnmarshalYAML(value *yaml.Node) error {
	// Use a shadow type to avoid recursion through this method.
	type shadow struct {
		Kind         string        `yaml:"kind"`
		ID           string        `yaml:"id"`
		Description  string        `yaml:"description"`
		Globs        []string      `yaml:"globs"`
		Phase        string        `yaml:"phase"`
		Triggers     []string      `yaml:"triggers"`
		Tools        []string      `yaml:"tools"`
		Model        string        `yaml:"model"`
		Args         yaml.Node     `yaml:"args"`
		Traces       []string      `yaml:"traces"`
		Deps         []string      `yaml:"deps"`
		Severity     string        `yaml:"severity"`
		Tier         string        `yaml:"tier"`
		HostTriggers []HostTrigger `yaml:"host_triggers"`
	}
	var s shadow
	if err := value.Decode(&s); err != nil {
		return err
	}
	f.Kind = s.Kind
	f.ID = s.ID
	f.Description = s.Description
	f.Globs = s.Globs
	f.Phase = s.Phase
	f.Triggers = s.Triggers
	f.Tools = s.Tools
	f.Model = s.Model
	f.Traces = s.Traces
	f.Deps = s.Deps
	f.Severity = s.Severity
	f.Tier = s.Tier
	f.HostTriggers = s.HostTriggers

	if s.Args.Kind == 0 {
		return nil
	}
	// args: [a, b, c]  or  args: ["a"]
	if s.Args.Kind == yaml.SequenceNode {
		for _, child := range s.Args.Content {
			switch child.Kind {
			case yaml.ScalarNode:
				f.Args = append(f.Args, Arg{Name: child.Value})
			case yaml.MappingNode:
				var a Arg
				if err := child.Decode(&a); err != nil {
					return fmt.Errorf("args entry: %w", err)
				}
				f.Args = append(f.Args, a)
			default:
				return fmt.Errorf("args entry: unsupported yaml kind %d", child.Kind)
			}
		}
		return nil
	}
	return fmt.Errorf("args must be a list")
}
