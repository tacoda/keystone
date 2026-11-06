package allocate

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/tacoda/keystone/internal/framework/adapters/claudecode"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

const claudeSettingsPath = ".claude/settings.json"

// claudeSettings moves the user's permission rules and matcher-free command
// hooks out of .claude/settings.json; keystone's own entries are skipped.
func claudeSettings(s *scan) error {
	rel, ok := s.file(claudeSettingsPath)
	if !ok {
		return nil
	}
	var doc struct {
		Permissions map[string][]string    `json:"permissions"`
		Hooks       map[string][]hookGroup `json:"hooks"`
	}
	if err := json.Unmarshal(s.read(rel), &doc); err != nil {
		s.add(Item{Source: rel, Note: "not valid JSON; left as is"})
		return nil
	}
	s.permissions(doc.Permissions)
	for _, event := range sortedKeys(doc.Hooks) {
		s.hooks(event, doc.Hooks[event])
	}
	return nil
}

type hookGroup struct {
	Matcher string `json:"matcher"`
	Hooks   []struct {
		Type          string `json:"type"`
		Command       string `json:"command"`
		StatusMessage string `json:"statusMessage"`
	} `json:"hooks"`
}

// permissions writes the user's allow/ask/deny rules as one permission
// primitive. Entries keystone added (managed_entries) are left out.
func (s *scan) permissions(perm map[string][]string) {
	mine := toSet(s.managed[claudeSettingsPath])
	var b strings.Builder
	var entries []string
	n := 0
	for _, bucket := range []string{"allow", "ask", "deny"} {
		var kept []string
		for _, rule := range perm[bucket] {
			if !mine[bucket+":"+rule] {
				kept = append(kept, rule)
			}
		}
		if len(kept) == 0 {
			continue
		}
		b.WriteString(bucket + ":\n")
		for _, rule := range kept {
			fmt.Fprintf(&b, "  - %q\n", rule)
			entries = append(entries, bucket+":"+rule)
		}
		n += len(kept)
	}
	if n == 0 {
		return
	}
	content := "---\nid: claude-settings\ndescription: Permission rules moved from .claude/settings.json.\n" + b.String() +
		"---\n\n# Claude Code settings permissions\n\nMoved by `keystone allocate`. Keystone translates these to every host that can enforce them.\n"
	s.add(Item{Source: claudeSettingsPath + " permissions", Kind: primitive.KindPermission,
		Target: ".agents/permissions/claude-settings.md", Content: []byte(content), Note: fmt.Sprintf("%d rule(s)", n), entries: entries})
}

// hooks turns each user command hook on event into a hook primitive. A
// matcher (a tool filter) has no keystone equivalent, so that hook stays.
func (s *scan) hooks(event string, groups []hookGroup) {
	i := 0
	for _, g := range groups {
		for _, h := range g.Hooks {
			if h.Type != "command" || strings.HasPrefix(h.StatusMessage, claudecode.HookEntryStatusPrefix) {
				continue
			}
			src := fmt.Sprintf("%s hooks.%s", claudeSettingsPath, event)
			if g.Matcher != "" && g.Matcher != "*" {
				s.add(Item{Source: src + "[" + g.Matcher + "]", Note: "matcher has no keystone equivalent; stays in settings.json"})
				continue
			}
			i++
			id := fmt.Sprintf("claude-%s-%d", slugify(event), i)
			content := fmt.Sprintf("---\nid: %s\ndescription: %s hook moved from .claude/settings.json.\nmode: computational\nevent: %s\nrun: %q\n---\n", id, event, event, h.Command)
			s.add(Item{Source: src, Kind: primitive.KindHook, Target: ".agents/hooks/" + id + ".md", Content: []byte(content),
				Note: "any non-zero exit now blocks (Claude Code blocked only on 2)", hook: &movedHook{event, h.Command}})
		}
	}
}

func toSet(xs []string) map[string]bool {
	set := map[string]bool{}
	for _, x := range xs {
		set[x] = true
	}
	return set
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
