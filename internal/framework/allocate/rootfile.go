package allocate

import (
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/tacoda/keystone/internal/framework/budget"
	"github.com/tacoda/keystone/internal/framework/primitive"
	"github.com/tacoda/keystone/internal/framework/projection"
)

// RootFiles are the always-loaded instruction files allocate de-bloats.
var RootFiles = []string{"CLAUDE.md", "AGENTS.md", "GEMINI.md", ".goosehints"}

// RootTotal is one root file's always-loaded size, without keystone's
// managed block, and what its movable sections would save.
type RootTotal struct {
	File          string
	Tokens, Saved int
}

// RootTotals sizes each root file present and sums the savings in items.
func RootTotals(projectDir string, items []Item) []RootTotal {
	s := scan{dir: projectDir}
	var out []RootTotal
	for _, rel := range RootFiles {
		if _, ok := s.file(rel); !ok {
			continue
		}
		tokens := budget.Estimate([]byte(withoutBlock(string(s.read(rel)))))
		out = append(out, RootTotal{File: rel, Tokens: tokens, Saved: savedFrom(rel, items)})
	}
	return out
}

// savedFrom sums the tokens of the items moved out of root file rel.
func savedFrom(rel string, items []Item) int {
	saved := 0
	for _, it := range items {
		if strings.HasPrefix(it.Source, rel+" § ") {
			saved += it.Tokens
		}
	}
	return saved
}

// Section is one `## ` section of a root file.
type Section struct {
	Heading, Body string
}

// Text is the section as it appears in the file.
func (sec Section) Text() string { return "## " + sec.Heading + "\n" + sec.Body }

// rootFiles proposes moving each section of an always-loaded root file to
// a kind that loads on demand. What applies everywhere stays, and so do
// iron laws and keystone's managed block. Heuristic: the keystone-allocate
// skill has an agent review the borderline calls.
func rootFiles(s *scan) error {
	for _, rel := range RootFiles {
		if _, ok := s.file(rel); !ok {
			continue
		}
		s.items = append(s.items, sectionItems(rel, string(s.read(rel)))...)
	}
	return nil
}

// sectionItems is the moving items for one root file's sections.
func sectionItems(rel, content string) []Item {
	var out []Item
	_, secs := Sections(content)
	for _, sec := range secs {
		if it, ok := classify(rel, sec); ok {
			out = append(out, it)
		}
	}
	return out
}

// Sections splits a root file into its preamble (everything before the
// first `## `, plus keystone's managed block, which never moves) and its
// sections.
func Sections(content string) (preamble string, secs []Section) {
	content = withoutBlock(content)
	parts := strings.Split("\n"+content, "\n## ")
	preamble = strings.TrimPrefix(parts[0], "\n")
	for _, p := range parts[1:] {
		heading, body, _ := strings.Cut(p, "\n")
		secs = append(secs, Section{Heading: strings.TrimSpace(heading), Body: body})
	}
	return preamble, secs
}

// withoutBlock removes keystone's managed block (it is regenerated, not
// the user's content).
func withoutBlock(content string) string {
	start := strings.Index(content, projection.BlockStart)
	end := strings.Index(content, projection.BlockEnd)
	if start < 0 || end < start {
		return content
	}
	return content[:start] + content[end+len(projection.BlockEnd):]
}

var (
	staysHeading   = regexp.MustCompile(`(?i)iron law|golden rule|non-negotiable|never|security|secret`)
	hookCue        = regexp.MustCompile(`(?i)before (you )?(commit|claim|finish|push)|always run`)
	backtickSpan   = regexp.MustCompile("`([^`\n]+)`")
	numberedStep   = regexp.MustCompile(`(?m)^\s*\d+\.\s`)
	filePattern    = regexp.MustCompile(`^[^\s()]*\*\.[A-Za-z0-9]+$`)
	rationaleTitle = regexp.MustCompile(`(?i)\b(why|background|rationale|history|architecture|context)\b`)
)

// referenceWords is the length past which a section reads as reasoning,
// not a directive. ponytail: fixed threshold; tune if reports misjudge.
const referenceWords = 300

// classify maps one section to an item that moves, or reports false when
// the section belongs in the always-loaded file.
func classify(file string, sec Section) (Item, bool) {
	if staysHeading.MatchString(sec.Heading) {
		return Item{}, false
	}
	slug := slugify(sec.Heading)
	if slug == "" {
		return Item{}, false
	}
	src := file + " § " + sec.Heading
	tokens := budget.Estimate([]byte(sec.Text()))
	for _, rule := range classifiers {
		if it, ok := rule(sec, slug); ok {
			it.Source, it.Tokens = src, tokens
			return it, true
		}
	}
	return Item{}, false
}

var classifiers = []func(Section, string) (Item, bool){hookSection, pathSection, procedureSection, referenceSection}

// hookSection: "before you commit, run `make lint`" is a check, so it
// becomes a hook on Stop that runs the first backticked command.
func hookSection(sec Section, slug string) (Item, bool) {
	cmd := cuedCommand(sec.Body)
	if cmd == "" {
		return Item{}, false
	}
	fm := frontmatter([2]string{"id", slug}, [2]string{"description", sec.Heading},
		[2]string{"mode", "computational"}, [2]string{"event", "Stop"}, [2]string{"run", cmd})
	return Item{Kind: primitive.KindHook, Target: ".agents/hooks/" + slug + ".md", Content: []byte(fm + sec.Text()),
		Note: "runs `" + cmd + "` on Stop and blocks on failure; check the event and command"}, true
}

// pathSection: a section that names file patterns applies only to them.
func pathSection(sec Section, slug string) (Item, bool) {
	globs := globsIn(sec.Text())
	if len(globs) == 0 {
		return Item{}, false
	}
	seq := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	for _, g := range globs {
		seq.Content = append(seq.Content, scalar(g))
	}
	doc := map[string]*yaml.Node{"id": scalar(slug), "description": scalar(sec.Heading), "paths": seq}
	full := render([]string{"id", "description", "paths"}, doc, sec.Text())
	return Item{Kind: primitive.KindRule, Target: ".agents/rules/" + slug + ".md", Content: full,
		Note: "loads only for " + strings.Join(globs, ", ")}, true
}

// procedureSection: three or more numbered steps read as a procedure the
// user asks for by name.
func procedureSection(sec Section, slug string) (Item, bool) {
	if len(numberedStep.FindAllString(sec.Body, -1)) < 3 {
		return Item{}, false
	}
	fm := frontmatter([2]string{"id", slug}, [2]string{"description", sec.Heading})
	return Item{Kind: primitive.KindCommand, Target: ".agents/commands/" + slug + ".md", Content: []byte(fm + sec.Text()),
		Note: "a procedure; make it a skill instead if the agent should start it unasked"}, true
}

// referenceSection: long or explanatory sections are reasoning, opened on
// demand.
func referenceSection(sec Section, slug string) (Item, bool) {
	if !rationaleTitle.MatchString(sec.Heading) && len(strings.Fields(sec.Body)) < referenceWords {
		return Item{}, false
	}
	fm := frontmatter([2]string{"id", slug}, [2]string{"description", sec.Heading})
	return Item{Kind: primitive.KindReference, Target: ".agents/references/" + slug + ".md", Content: []byte(fm + sec.Text())}, true
}

// globsIn lists backticked file patterns that name an extension (`*.go`,
// `src/**/*.tsx`), in order. A folder pattern such as `.claude/*` is more
// often an output being described than the scope of the section.
func globsIn(text string) []string {
	var out []string
	seen := map[string]bool{}
	for _, m := range backtickSpan.FindAllStringSubmatch(text, -1) {
		g := m[1]
		if filePattern.MatchString(g) && !seen[g] {
			seen[g] = true
			out = append(out, g)
		}
	}
	return out
}

// cuedCommand is the first backticked command on a line that also says
// when to run it ("before you commit, run `make lint`"); "" when none.
func cuedCommand(body string) string {
	for _, line := range strings.Split(body, "\n") {
		if hookCue.MatchString(line) {
			if cmd := commandIn(line); cmd != "" {
				return cmd
			}
		}
	}
	return ""
}

// commandIn is the first backticked span in line that is not a pattern.
func commandIn(line string) string {
	for _, m := range backtickSpan.FindAllStringSubmatch(line, -1) {
		if !strings.Contains(m[1], "*") {
			return m[1]
		}
	}
	return ""
}
