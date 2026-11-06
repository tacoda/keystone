package allocate

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// movedHook is a matcher-free command hook moved out of settings.json.
type movedHook struct{ event, command string }

// Ownership is what --overwrite hands to keystone: host paths it now owns
// (lockfile `projections`, so `keystone project` replaces or prunes them)
// and settings.json entries it now manages (`managed_entries`).
type Ownership struct {
	Owned   []string
	Managed map[string][]string
	Kept    []string // originals left alone, each with the reason
}

// hostOf names the host that reads a source path, "" for none keystone
// projects to (Windsurf, Gemini).
func hostOf(source string) string {
	for _, h := range hostPrefixes {
		if strings.HasPrefix(source, h[0]) {
			return h[1]
		}
	}
	return ""
}

var hostPrefixes = [][2]string{
	{".claude/", "claude-code"}, {"CLAUDE.md", "claude-code"}, {"AGENTS.md", "claude-code"},
	{".cursor/", "cursor"}, {".cursorrules", "cursor"}, {".github/", "github-copilot"},
	{".clinerules", "cline"}, {".continue/", "continue"}, {".opencode/", "opencode"},
}

// Overwrite takes over the originals of the items whose targets were
// written: host files become keystone-owned, moved sections leave their
// root file, moved hooks leave settings.json, and moved permission rules
// become keystone-managed entries. An item whose target already existed
// is left alone, so its original is never lost.
//
// enabled lists the hosts keystone projects to (keystone.json adapters): an
// original is handed over only when keystone will project to its host,
// since an owned path no projection covers is pruned.
func Overwrite(projectDir string, items []Item, wrote, enabled []string) (Ownership, error) {
	h := handover{written: toSet(wrote), hosts: toSet(enabled), sections: map[string][]string{},
		own: Ownership{Managed: map[string][]string{}}}
	for _, it := range items {
		h.claim(it)
	}
	for file, texts := range h.sections {
		if err := removeSections(filepath.Join(projectDir, file), texts); err != nil {
			return h.own, err
		}
	}
	return h.own, removeHooks(filepath.Join(projectDir, claudeSettingsPath), h.hooks)
}

// handover collects what Overwrite hands to keystone.
type handover struct {
	written, hosts map[string]bool
	own            Ownership
	sections       map[string][]string // root file → section texts to remove
	hooks          []movedHook
}

// claim takes one written item's original, or records why it is kept.
func (h *handover) claim(it Item) {
	if !h.written[it.Target] {
		return
	}
	if host := hostOf(it.Source); !h.hosts[host] {
		h.own.Kept = append(h.own.Kept, it.Source+" — "+keptReason(host))
		return
	}
	switch {
	case it.section != "":
		file, _, _ := strings.Cut(it.Source, " § ")
		h.sections[file] = append(h.sections[file], it.section)
	case it.hook != nil:
		h.hooks = append(h.hooks, *it.hook)
	case len(it.entries) > 0:
		h.own.Managed[claudeSettingsPath] = append(h.own.Managed[claudeSettingsPath], it.entries...)
	case it.own != "":
		h.own.Owned = appendNew(h.own.Owned, it.own)
	}
}

func keptReason(host string) string {
	if host == "" {
		return "keystone projects to no host that reads it"
	}
	return "add it with `keystone target add " + host + "`, then rerun"
}

// removeSections deletes each moved section's text from a root file. A
// section that no longer matches the file (edited since) is left in place.
func removeSections(path string, texts []string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)
	for _, t := range texts {
		content = strings.Replace(content, "\n"+t, "", 1)
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

// removeHooks drops the moved command hooks from settings.json, keeping
// every other key, group, and hook.
func removeHooks(path string, moved []movedHook) error {
	if len(moved) == 0 {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		return err
	}
	hooks, _ := doc["hooks"].(map[string]any)
	for _, m := range moved {
		hooks[m.event] = withoutCommand(hooks[m.event], m.command)
		if list, _ := hooks[m.event].([]any); len(list) == 0 {
			delete(hooks, m.event)
		}
	}
	out, err := marshalIndent(doc)
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o644)
}

// withoutCommand removes the command hook from matcher-free groups, and
// any group it leaves empty.
func withoutCommand(groups any, command string) []any {
	list, _ := groups.([]any)
	var out []any
	for _, g := range list {
		group, _ := g.(map[string]any)
		if m, _ := group["matcher"].(string); m == "" || m == "*" {
			group["hooks"] = dropCommand(group["hooks"], command)
		}
		if hs, _ := group["hooks"].([]any); len(hs) > 0 {
			out = append(out, group)
		}
	}
	return out
}

func dropCommand(hooks any, command string) []any {
	list, _ := hooks.([]any)
	var out []any
	for _, h := range list {
		if entry, _ := h.(map[string]any); entry["command"] != command {
			out = append(out, h)
		}
	}
	return out
}

func appendNew(list []string, x string) []string {
	for _, e := range list {
		if e == x {
			return list
		}
	}
	return append(list, x)
}

// marshalIndent writes JSON the way keystone's own merges do (no HTML
// escaping, so `a && b` stays readable).
func marshalIndent(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	err := enc.Encode(v)
	return buf.Bytes(), err
}
