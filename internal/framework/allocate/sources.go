package allocate

import (
	"path"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

// sources is every host-specific location allocate reads (docs/hosts.md).
var sources = []func(*scan) error{
	claudeAndCursor, cursorRules, copilot, clineAndWindsurf, continueFiles, opencodeFiles, claudeSettings, rootFiles,
}

func claudeAndCursor(s *scan) error {
	for _, host := range []string{".claude", ".cursor"} {
		s.skills(host + "/skills")
		s.each(host+"/agents", ".md", primitive.KindSubagent, ".agents/agents")
		s.each(host+"/commands", ".md", primitive.KindCommand, ".agents/commands")
	}
	s.each(".claude/rules", ".md", primitive.KindRule, ".agents/rules")
	return nil
}

// cursorRules: a Cursor rule with a description but no globs and no
// alwaysApply is "agent requested", which is what a skill is; the rest
// are rules.
func cursorRules(s *scan) error {
	for _, rel := range s.files(".cursor/rules", ".mdc") {
		src := s.read(rel)
		if desc, ok := agentRequested(src); ok {
			slug := slugify(strings.TrimSuffix(path.Base(rel), ".mdc"))
			body := frontmatter([2]string{"name", slug}, [2]string{"description", desc}) + primitive.Body(string(src))
			s.add(Item{Source: rel, Kind: primitive.KindSkill, Target: ".agents/skills/" + slug + "/SKILL.md",
				Content: []byte(body), Note: "agent-requested rule → skill", own: rel})
			continue
		}
		s.convertTo(rel, primitive.KindRule, ".agents/rules/"+strings.TrimSuffix(strings.TrimPrefix(rel, ".cursor/rules/"), ".mdc")+".md")
	}
	s.single(".cursorrules", primitive.KindRule, ".agents/rules/cursorrules.md")
	return nil
}

func copilot(s *scan) error {
	s.single(".github/copilot-instructions.md", primitive.KindRule, ".agents/rules/copilot-instructions.md")
	s.each(".github/instructions", ".instructions.md", primitive.KindRule, ".agents/rules")
	s.each(".github/prompts", ".prompt.md", primitive.KindCommand, ".agents/commands")
	s.each(".github/agents", ".md", primitive.KindSubagent, ".agents/agents")
	return nil
}

func clineAndWindsurf(s *scan) error {
	s.single(".clinerules", primitive.KindRule, ".agents/rules/clinerules.md") // the file form
	s.each(".clinerules", ".md", primitive.KindRule, ".agents/rules/cline")    // the folder form
	s.single(".windsurfrules", primitive.KindRule, ".agents/rules/windsurfrules.md")
	return nil
}

func continueFiles(s *scan) error {
	s.each(".continue/rules", ".md", primitive.KindRule, ".agents/rules")
	s.each(".continue/prompts", ".prompt", primitive.KindCommand, ".agents/commands")
	return nil
}

func opencodeFiles(s *scan) error {
	for _, dir := range []string{"agent", "agents"} {
		s.each(".opencode/"+dir, ".md", primitive.KindSubagent, ".agents/agents")
	}
	for _, dir := range []string{"command", "commands"} {
		s.each(".opencode/"+dir, ".md", primitive.KindCommand, ".agents/commands")
	}
	return nil
}

// each converts every file under dir ending in suffix to kind, at the same
// relative path under dest. A rule's id keeps its folders (process/spec);
// other kinds use the file name.
func (s *scan) each(dir, suffix string, kind primitive.Kind, dest string) {
	for _, rel := range s.files(dir, suffix) {
		sub := strings.TrimSuffix(strings.TrimPrefix(rel, dir+"/"), suffix)
		sub = strings.TrimSuffix(sub, ".agent") // Copilot's name.agent.md
		s.convertTo(rel, kind, dest+"/"+sub+".md")
	}
}

// single converts one file, when it exists.
func (s *scan) single(rel string, kind primitive.Kind, target string) {
	if _, ok := s.file(rel); ok {
		s.convertTo(rel, kind, target)
	}
}

func (s *scan) convertTo(rel string, kind primitive.Kind, target string) {
	content, note := convert(s.read(rel), idFor(kind, target))
	if kind == primitive.KindSubagent && !strings.Contains(string(content), "\ntools:") {
		note = joinNote(note, "add `tools:` — the host gave this subagent every tool")
	}
	s.add(Item{Source: rel, Kind: kind, Target: target, Content: content, Note: note, own: rel})
}

func joinNote(a, b string) string {
	if a == "" {
		return b
	}
	return a + "; " + b
}

// skills copies each skill folder file by file; SKILL.md is already in
// Agent Skills shape on these hosts.
func (s *scan) skills(dir string) {
	for _, rel := range s.files(dir, "") {
		target := ".agents/skills/" + strings.TrimPrefix(rel, dir+"/")
		name, _, _ := strings.Cut(strings.TrimPrefix(rel, dir+"/"), "/")
		s.add(Item{Source: rel, Kind: primitive.KindSkill, Target: target, Content: s.read(rel), own: dir + "/" + name})
	}
}

// idFor derives an id from the target path: a rule keeps its folders
// under .agents/rules/; anything else is the file name.
func idFor(kind primitive.Kind, target string) string {
	name := strings.TrimSuffix(target, ".md")
	if kind == primitive.KindRule {
		return strings.TrimPrefix(name, ".agents/rules/")
	}
	return path.Base(name)
}

// agentRequested reports a Cursor rule that applies by description alone.
func agentRequested(src []byte) (string, bool) {
	fm, _, ok := primitive.SplitFrontmatter(string(src))
	if !ok {
		return "", false
	}
	var r struct {
		Description string `yaml:"description"`
		Globs       any    `yaml:"globs"`
		AlwaysApply bool   `yaml:"alwaysApply"`
	}
	if yaml.Unmarshal([]byte(fm), &r) != nil {
		return "", false
	}
	return r.Description, r.Description != "" && r.Globs == nil && !r.AlwaysApply
}

// slugify makes an Agent Skills name: lowercase letters, digits, hyphens.
func slugify(s string) string {
	return strings.Trim(nonSlug.ReplaceAllString(strings.ToLower(s), "-"), "-")
}

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)
