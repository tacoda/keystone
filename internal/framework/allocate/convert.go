package allocate

import (
	"bytes"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

// convert rewrites a host file's frontmatter into charter frontmatter and
// keeps its body. Host keys keystone does not know are kept. It returns
// the new file and a note on what changed ("" when nothing did).
func convert(src []byte, id string) ([]byte, string) {
	fm, body, ok := primitive.SplitFrontmatter(string(src))
	if !ok {
		fm, body = "", string(src)
	}
	keys, doc := parseMapping(fm)
	var notes []string
	keys, notes = normalizePaths(keys, doc, notes)
	notes = splitList(doc, "tools", notes)
	keys, notes = ensureIdentity(keys, doc, id, body, notes)
	return render(keys, doc, body), strings.Join(notes, "; ")
}

// parseMapping reads frontmatter as ordered keys and their nodes; anything
// that is not a YAML mapping reads as empty.
func parseMapping(fm string) ([]string, map[string]*yaml.Node) {
	var n yaml.Node
	if yaml.Unmarshal([]byte(fm), &n) != nil || len(n.Content) != 1 {
		return nil, map[string]*yaml.Node{}
	}
	if n.Content[0].Kind != yaml.MappingNode {
		return nil, map[string]*yaml.Node{}
	}
	return mappingOf(n.Content[0])
}

// ensureIdentity adds the `description:` and `id:` lint requires when the
// host file has neither (a spec `name:` serves as the id).
func ensureIdentity(keys []string, doc map[string]*yaml.Node, id, body string, notes []string) ([]string, []string) {
	if doc["description"] == nil {
		keys = append([]string{"description"}, keys...)
		doc["description"] = scalar(describe(body))
		notes = append(notes, "description taken from the body; check it")
	}
	if doc["id"] == nil && doc["name"] == nil {
		keys = append([]string{"id"}, keys...)
		doc["id"] = scalar(id)
	}
	return keys, notes
}

// describe is a one-line description from a body: its first non-empty
// line, without heading marks, cut to 120 characters.
func describe(body string) string {
	for _, line := range strings.Split(body, "\n") {
		if l := strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(line), "#")); l != "" {
			if len(l) > 120 {
				l = l[:117] + "..."
			}
			return l
		}
	}
	return "TODO — describe what this governs."
}

// normalizePaths maps the host's path-scoping keys to `paths:`: Cursor
// `globs:` and Copilot `applyTo:` (comma strings become lists), and drops
// Cursor `alwaysApply:` (true means no paths at all).
func normalizePaths(keys []string, doc map[string]*yaml.Node, notes []string) ([]string, []string) {
	for _, k := range []string{"globs", "applyTo"} {
		if n := doc[k]; n != nil {
			doc["paths"] = listNode(n)
			keys = replaceKey(keys, k, "paths")
			delete(doc, k)
			notes = append(notes, k+" → paths")
		}
	}
	if n := doc["alwaysApply"]; n != nil {
		if n.Value == "true" && doc["paths"] != nil {
			keys = removeKey(keys, "paths")
			delete(doc, "paths")
		}
		keys = removeKey(keys, "alwaysApply")
		delete(doc, "alwaysApply")
		notes = append(notes, "alwaysApply dropped")
	}
	if p := doc["paths"]; p != nil && isMatchAll(p) {
		keys = removeKey(keys, "paths")
		delete(doc, "paths")
	}
	return keys, notes
}

// splitList turns a comma-separated scalar (Claude Code `tools: Read, Grep`)
// into a YAML list.
func splitList(doc map[string]*yaml.Node, key string, notes []string) []string {
	if n := doc[key]; n != nil && n.Kind == yaml.ScalarNode {
		doc[key] = listNode(n)
		notes = append(notes, key+" → list")
	}
	return notes
}

// listNode is n as a YAML sequence; a scalar is split on commas.
func listNode(n *yaml.Node) *yaml.Node {
	if n.Kind == yaml.SequenceNode {
		return n
	}
	seq := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	for _, part := range strings.Split(n.Value, ",") {
		if p := strings.TrimSpace(part); p != "" {
			seq.Content = append(seq.Content, scalar(p))
		}
	}
	return seq
}

// isMatchAll: a paths list that matches every file is the same as none.
func isMatchAll(n *yaml.Node) bool {
	for _, c := range n.Content {
		if matchAll[c.Value] {
			return true
		}
	}
	return false
}

var matchAll = map[string]bool{"**": true, "**/*": true, "*": true}

func scalar(v string) *yaml.Node { return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: v} }

func mappingOf(m *yaml.Node) ([]string, map[string]*yaml.Node) {
	doc := map[string]*yaml.Node{}
	var keys []string
	for i := 0; i+1 < len(m.Content); i += 2 {
		keys = append(keys, m.Content[i].Value)
		doc[m.Content[i].Value] = m.Content[i+1]
	}
	return keys, doc
}

func replaceKey(keys []string, from, to string) []string {
	for i, k := range keys {
		if k == from {
			keys[i] = to
		}
	}
	return keys
}

func removeKey(keys []string, key string) []string {
	out := keys[:0]
	for _, k := range keys {
		if k != key {
			out = append(out, k)
		}
	}
	return out
}

// render writes keys (in order) from doc as frontmatter, then the body.
func render(keys []string, doc map[string]*yaml.Node, body string) []byte {
	m := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	for _, k := range keys {
		m.Content = append(m.Content, scalar(k), doc[k])
	}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	_ = enc.Encode(m)
	return []byte("---\n" + buf.String() + "---\n" + strings.TrimLeft(body, "\n"))
}

// frontmatter renders a fresh frontmatter block from ordered pairs.
func frontmatter(pairs ...[2]string) string {
	var keys []string
	doc := map[string]*yaml.Node{}
	for _, p := range pairs {
		keys = append(keys, p[0])
		doc[p[0]] = scalar(p[1])
	}
	return string(render(keys, doc, ""))
}
