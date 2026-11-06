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
// fmShadow mirrors Frontmatter's YAML shape without its UnmarshalYAML
// method, so decoding into it doesn't recurse. Kept at package scope to
// keep UnmarshalYAML itself small.
type fmShadow struct {
	Kind          string            `yaml:"kind"`
	ID            string            `yaml:"id"`
	Description   string            `yaml:"description"`
	Name          string            `yaml:"name"`
	License       string            `yaml:"license"`
	Compatibility string            `yaml:"compatibility"`
	Metadata      map[string]string `yaml:"metadata"`
	AllowedTools  string            `yaml:"allowed-tools"`
	Paths         []string          `yaml:"paths"`
	Globs         []string          `yaml:"globs"` // 4.x alias for `paths:`
	Phase         string            `yaml:"phase"`
	Triggers      []string          `yaml:"triggers"`
	Tools         []string          `yaml:"tools"`
	Model         string            `yaml:"model"`
	Args          yaml.Node         `yaml:"args"`
	Corpus        []string          `yaml:"corpus"`
	Deps          []string          `yaml:"deps"`
	Severity      string            `yaml:"severity"`
	Tier          string            `yaml:"tier"`
	Tags          []string          `yaml:"tags"`
	Includes      []string          `yaml:"includes"`
	HostTriggers  []HostTrigger     `yaml:"host_triggers"`
	Produces      []string          `yaml:"produces"`
	Consumes      []string          `yaml:"consumes"`
	Stop          string            `yaml:"stop"`
	Gates         []string          `yaml:"gates"`
	Gate          string            `yaml:"gate"`
	Type          string            `yaml:"type"`
	ProducedBy    string            `yaml:"produced_by"`
	Supersedes    []string          `yaml:"supersedes"`
	Mode          string            `yaml:"mode"`
	Event         string            `yaml:"event"`
	On            string            `yaml:"on"` // 4.x alias for `event:`
	Run           string            `yaml:"run"`
	Transport     string            `yaml:"transport"`
	Agent         string            `yaml:"agent"`
	Returns       string            `yaml:"returns"`
	Allow         []string          `yaml:"allow"`
	Ask           []string          `yaml:"ask"`
	Deny          []string          `yaml:"deny"`
}

func (f *Frontmatter) UnmarshalYAML(value *yaml.Node) error {
	var s fmShadow
	if err := value.Decode(&s); err != nil {
		return err
	}
	k, _ := NormalizeKind(s.Kind) // 4.x kind names map to 5.0 kinds
	f.Kind = string(k)
	f.ID = firstNonEmpty(s.ID, s.Name) // spec skills carry `name`, not `id`
	f.Description = s.Description
	f.Name = s.Name
	f.License = s.License
	f.Compatibility = s.Compatibility
	f.Metadata = s.Metadata
	f.AllowedTools = s.AllowedTools
	f.Keys = mappingKeys(value)
	f.Paths = s.Paths
	if len(f.Paths) == 0 {
		f.Paths = s.Globs
	}
	f.Phase = s.Phase
	f.Triggers = s.Triggers
	f.Tools = s.Tools
	f.Model = s.Model
	f.Corpus = s.Corpus
	f.Deps = s.Deps
	f.Severity = s.Severity
	f.Tier = s.Tier
	f.Tags = s.Tags
	f.Includes = s.Includes
	f.HostTriggers = s.HostTriggers
	f.Produces = s.Produces
	f.Consumes = s.Consumes
	f.Stop = s.Stop
	f.Gates = s.Gates
	f.Gate = s.Gate
	f.Type = s.Type
	f.ProducedBy = s.ProducedBy
	f.Supersedes = s.Supersedes
	f.Mode = s.Mode
	f.Event = firstNonEmpty(s.Event, s.On) // `event:` canonical; `on:` alias
	f.Run = s.Run
	f.Transport = s.Transport
	f.Agent = s.Agent
	f.Returns = s.Returns
	f.Allow = s.Allow
	f.Ask = s.Ask
	f.Deny = s.Deny
	f.applySpecFallbacks()
	args, err := decodeArgsNode(s.Args)
	if err != nil {
		return err
	}
	f.Args = args
	return nil
}

// applySpecFallbacks fills keystone fields from their Agent Skills spec
// spellings when the keystone spelling is absent.
func (f *Frontmatter) applySpecFallbacks() {
	if len(f.Tools) == 0 && f.AllowedTools != "" {
		f.Tools = strings.Fields(f.AllowedTools)
	}
	if f.Model == "" {
		f.Model = f.Metadata["model"]
	}
	if len(f.Tags) == 0 && f.Metadata["tags"] != "" {
		for _, t := range strings.Split(f.Metadata["tags"], ",") {
			f.Tags = append(f.Tags, strings.TrimSpace(t))
		}
	}
}

// mappingKeys returns the top-level keys of a YAML mapping, in order.
func mappingKeys(node *yaml.Node) []string {
	if node.Kind != yaml.MappingNode {
		return nil
	}
	keys := make([]string, 0, len(node.Content)/2)
	for i := 0; i < len(node.Content); i += 2 {
		keys = append(keys, node.Content[i].Value)
	}
	return keys
}

// firstNonEmpty returns the first non-empty string, else "".
func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// decodeArgsNode lifts the flexible `args:` form into []Arg. Each entry may be
// a scalar (sugar for Arg{Name}) or a full {name,type,required,description}
// mapping. A zero/absent node yields nil; anything but a list is an error.
func decodeArgsNode(node yaml.Node) ([]Arg, error) {
	if node.Kind == 0 {
		return nil, nil
	}
	if node.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("args must be a list")
	}
	var args []Arg
	for _, child := range node.Content {
		switch child.Kind {
		case yaml.ScalarNode:
			args = append(args, Arg{Name: child.Value})
		case yaml.MappingNode:
			var a Arg
			if err := child.Decode(&a); err != nil {
				return nil, fmt.Errorf("args entry: %w", err)
			}
			args = append(args, a)
		default:
			return nil, fmt.Errorf("args entry: unsupported yaml kind %d", child.Kind)
		}
	}
	return args, nil
}
