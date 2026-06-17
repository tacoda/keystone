package web

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tacoda/keystone/internal/framework/config"
)

// InboxItem is one learning-candidate file under
// .keystone/harness/learning/inbox/. The body is the raw markdown
// the agent wrote when it captured the candidate.
type InboxItem struct {
	Name   string `json:"name"`     // filename only
	Path   string `json:"path"`     // repo-relative
	Body   string `json:"body"`
	Status string `json:"status"`   // pulled from frontmatter when present
}

// listInbox walks the learning inbox and returns each candidate in
// chronological order (newest first by mtime).
func listInbox(projectDir string) ([]InboxItem, error) {
	dir := filepath.Join(projectDir, config.DefaultHarnessRoot, "learning", "inbox")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	out := make([]InboxItem, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(e.Name(), ".md") || e.Name() == "README.md" {
			continue
		}
		abs := filepath.Join(dir, e.Name())
		body, err := os.ReadFile(abs)
		if err != nil {
			continue
		}
		out = append(out, InboxItem{
			Name:   e.Name(),
			Path:   filepath.ToSlash(filepath.Join(config.DefaultHarnessRoot, "learning", "inbox", e.Name())),
			Body:   string(body),
			Status: extractStatus(string(body)),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name > out[j].Name })
	return out, nil
}

// extractStatus pulls `status:` from frontmatter (the inbox uses
// `status: new | reviewed | accepted | rejected`).
func extractStatus(body string) string {
	if !strings.HasPrefix(body, "---") {
		return ""
	}
	end := strings.Index(body[3:], "---")
	if end < 0 {
		return ""
	}
	fm := body[3 : 3+end]
	for _, line := range strings.Split(fm, "\n") {
		s := strings.TrimSpace(line)
		if strings.HasPrefix(s, "status:") {
			return strings.TrimSpace(strings.TrimPrefix(s, "status:"))
		}
	}
	return ""
}

func (s *server) handleInbox(w http.ResponseWriter, r *http.Request) {
	items, err := listInbox(s.projectDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "inbox.html", map[string]any{
		"ProjectDir": s.projectDir,
		"Items":      items,
	})
}

// handleActionInboxReject deletes a learning-inbox candidate
// outright. Rejection in the canonical flywheel moves the file into
// archive/ — but at the dashboard layer "delete" is the simplest
// expression of "I don't want this." Users who want archive history
// can rename instead of clicking reject.
func (s *server) handleActionInboxReject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		htmxFragment(w, errorFragment(err.Error()))
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" || strings.ContainsAny(name, "/\\") {
		htmxFragment(w, errorFragment("invalid name"))
		return
	}
	path := filepath.Join(s.projectDir, config.DefaultHarnessRoot, "learning", "inbox", name)
	if err := os.Remove(path); err != nil {
		htmxFragment(w, errorFragment(err.Error()))
		return
	}
	htmxRedirect(w, r, "/inbox")
}

// handleActionInboxAccept renames a candidate to mark it accepted.
// The actual promotion (move into guides/ + paired corpus/) is
// agent-driven via the synthesize action — the dashboard just
// stamps the file's status so the agent knows to promote it next
// time `synthesize` runs.
func (s *server) handleActionInboxAccept(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		htmxFragment(w, errorFragment(err.Error()))
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" || strings.ContainsAny(name, "/\\") {
		htmxFragment(w, errorFragment("invalid name"))
		return
	}
	path := filepath.Join(s.projectDir, config.DefaultHarnessRoot, "learning", "inbox", name)
	body, err := os.ReadFile(path)
	if err != nil {
		htmxFragment(w, errorFragment(err.Error()))
		return
	}
	updated := setOrAddStatus(string(body), "accepted")
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		htmxFragment(w, errorFragment(err.Error()))
		return
	}
	htmxRedirect(w, r, "/inbox")
}

// setOrAddStatus updates the `status:` frontmatter key, or inserts
// one if missing. Preserves the rest of the file byte-for-byte.
func setOrAddStatus(body, status string) string {
	if !strings.HasPrefix(body, "---") {
		// No frontmatter — prepend a fresh block.
		return "---\nstatus: " + status + "\n---\n\n" + body
	}
	end := strings.Index(body[3:], "---")
	if end < 0 {
		return body
	}
	fm := body[3 : 3+end]
	rest := body[3+end:]
	lines := strings.Split(fm, "\n")
	updated := false
	for i, line := range lines {
		s := strings.TrimSpace(line)
		if strings.HasPrefix(s, "status:") {
			lines[i] = "status: " + status
			updated = true
			break
		}
	}
	if !updated {
		// Prepend status to frontmatter.
		fm = "\nstatus: " + status + fm
		return "---" + fm + rest
	}
	return "---" + strings.Join(lines, "\n") + rest
}
