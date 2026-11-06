package loader

import (
	"io/fs"
	"os"
	"path/filepath"
)

// portFolders maps a `strict` / `required` key to its 5.0 folder under
// .agents/. Keys may be a 5.0 folder, a kind, or a 4.x port, so a policy
// written for 4.x keeps locking the same items after the 5.0 move.
var portFolders = map[string]string{
	"rules": "rules", "rule": "rules", "guides": "rules", "guide": "rules",
	"references": "references", "reference": "references", "corpus": "references",
	"hooks": "hooks", "hook": "hooks", "sensors": "hooks", "sensor": "hooks",
	"permissions": "permissions", "permission": "permissions", "posture": "permissions",
	"skills": "skills", "skill": "skills", "playbooks": "skills", "playbook": "skills", "patterns": "skills",
	"commands": "commands", "command": "commands", "actions": "commands", "action": "commands",
	"agents": "agents", "subagent": "agents", "subagents": "agents",
}

// legacyFolders are the 4.x folders a vendored policy tree may still use.
var legacyFolders = map[string][]string{
	"rules": {"guides"}, "references": {"corpus"}, "hooks": {"sensors"},
	"permissions": {"posture"}, "skills": {"playbooks", "patterns"}, "commands": {"actions"},
}

// folderFor is the 5.0 folder for a strict/required key; an unknown key is
// taken as a folder name.
func folderFor(key string) string {
	if f, ok := portFolders[key]; ok {
		return f
	}
	return key
}

// policyFolders are the folders to search inside a vendored policy: the
// 5.0 folder, then its 4.x names.
func policyFolders(key string) []string {
	f := folderFor(key)
	return append([]string{f}, legacyFolders[f]...)
}

// isItem reports whether path is item: `<item>.md` anywhere under the
// folder, or a skill folder `<item>/SKILL.md`.
func isItem(path, item string) bool {
	if filepath.Base(path) == item+".md" {
		return true
	}
	return filepath.Base(path) == "SKILL.md" && filepath.Base(filepath.Dir(path)) == item
}

// findItem lists every file under root that is item, as paths relative to
// base. A missing root has none.
func findItem(base, root, item string) ([]string, error) {
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return nil, nil
	}
	var hits []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !isItem(path, item) {
			return nil
		}
		rel, relErr := filepath.Rel(base, path)
		hits = append(hits, filepath.ToSlash(rel))
		return relErr
	})
	return hits, err
}
