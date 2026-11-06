package projection

import (
	"path"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

// ClaudeCode is the Claude Code profile (docs/hosts.md). Skills are linked
// per entry: Claude Code reads the Agent Skills format and follows
// symlinked skill directories, but does not scan .agents/skills/. Rules,
// commands, and subagents differ in format (Claude rules read only
// `paths:`; subagents need `name:`; commands use `argument-hint:`), so
// they get shims. Hooks and permissions live in .claude/settings.json and
// are projected by the claudecode adapter.
func ClaudeCode() Profile {
	shim := Rule{Strategy: Shim, Dest: primitive.ProjectionRelPath}
	return Profile{
		Host: "claude-code",
		Rules: map[primitive.Kind]Rule{
			primitive.KindSkill:    {Strategy: Symlink, Dest: claudeSkillDest, Target: skillDir},
			primitive.KindRule:     shim,
			primitive.KindCommand:  shim,
			primitive.KindSubagent: shim,
			primitive.KindHook:     shim, // inferential hooks only (Dest is "" otherwise)
		},
		Reads: map[primitive.Kind][]string{
			primitive.KindSkill: {".claude/skills"},
		},
	}
}

// claudeSkillDest links the whole skill directory under its canonical
// name, so the Agent Skills rule "name matches the directory" holds on
// the host too. Flat 4.x skill files (a playbook not yet migrated) have no
// directory to link and are skipped.
func claudeSkillDest(p primitive.Primitive) string {
	if path.Base(p.Path) != "SKILL.md" {
		return ""
	}
	return path.Join(".claude/skills", path.Base(skillDir(p)))
}

// Enabled returns the profiles to project, in plan order. Claude Code is
// always projected (as in 4.x); the other hosts move onto this table in
// Phase 4, selected by keystone.json `adapters`.
func Enabled() []Profile {
	return []Profile{ClaudeCode()}
}
