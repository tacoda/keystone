package projection

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHookBridges_PerHost(t *testing.T) {
	root := seedCharter(t) // one computational hook on Stop
	write(t, root, map[string]string{".cursor/hooks.json": `{"version":1,"hooks":{"stop":[{"command":"./mine.sh"}]}}`})
	pr := Plan(walk(t, root), Enabled(allHosts))
	if _, err := Apply(root, pr.Outputs, nil, Options{Mode: LinkSymlink}); err != nil {
		t.Fatal(err)
	}
	assertContains(t, root, ".claude/settings.json", `"Stop"`, "keystone hook fire Stop --host claude-code")
	assertContains(t, root, ".cursor/hooks.json", `"stop"`, "./mine.sh", "keystone hook fire Stop --host cursor", `"failClosed": true`)
	assertContains(t, root, ".github/hooks/keystone.json", `"agentStop"`, `"bash": "keystone hook fire Stop --host github-copilot || exit 2"`)
	assertContains(t, root, ".opencode/plugins/keystone.js", `"session.idle"`, "keystone hook fire Stop --host opencode < ${new Response(JSON.stringify(input))}")
	reported := reportedHosts(pr.Outputs, "hook")
	for _, h := range []string{"continue", "codex", "cline", "goose", "pi", "aider"} {
		if !reported[h] {
			t.Errorf("%s has no bridge and must be reported", h)
		}
	}
	if d := Check(root, pr.Outputs, LinkSymlink); len(d) != 0 {
		t.Errorf("fresh bridges drifted: %v", d)
	}
}

func assertContains(t *testing.T, root, rel string, wants ...string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Errorf("%s: %v", rel, err)
		return
	}
	for _, w := range wants {
		if !strings.Contains(string(data), w) {
			t.Errorf("%s missing %q:\n%s", rel, w, data)
		}
	}
}

func reportedHosts(outs []Output, kind string) map[string]bool {
	got := map[string]bool{}
	for _, o := range outs {
		if o.Strategy == Report && o.Kind == kind {
			got[o.Host] = true
		}
	}
	return got
}

func TestHookBridges_Warnings(t *testing.T) {
	root := seedCharter(t)
	write(t, root, map[string]string{".agents/hooks/notify.md": "---\nid: notify\ndescription: x\nmode: computational\nevent: Notification\nrun: \"true\"\n---\n"})
	var msgs []string
	for _, w := range Plan(walk(t, root), Enabled([]string{"cursor", "github-copilot"})).Warnings {
		msgs = append(msgs, w.Host+": "+w.Message)
	}
	all := strings.Join(msgs, "\n")
	for _, want := range []string{"cursor: no host event for Notification", "github-copilot: also reads .claude/settings.json hooks"} {
		if !strings.Contains(all, want) {
			t.Errorf("warnings missing %q:\n%s", want, all)
		}
	}
}

func TestMergeVersionedHooks_Idempotent(t *testing.T) {
	entry := func(p string) map[string]any { return map[string]any{"command": bridgeCmd(p, "cursor")} }
	once, err := mergeVersionedHooks([]byte(`{"version":1,"hooks":{"stop":[{"command":"./mine.sh"}]}}`), []string{"Stop"}, cursorEvents, entry, "command")
	if err != nil {
		t.Fatal(err)
	}
	twice, _ := mergeVersionedHooks(once, []string{"Stop"}, cursorEvents, entry, "command")
	if string(once) != string(twice) {
		t.Errorf("not idempotent:\n%s\n---\n%s", once, twice)
	}
	none, _ := mergeVersionedHooks(once, nil, cursorEvents, entry, "command")
	if strings.Contains(string(none), "keystone hook fire") || !strings.Contains(string(none), "./mine.sh") {
		t.Errorf("removing all hooks should drop only keystone's entries:\n%s", none)
	}
}

// TestHookBridges_RemovedHooksLeaveSharedFiles — deleting the last host-event
// hook strips keystone's entries from shared config files and prunes owned
// bridge files; user entries stay.
func TestHookBridges_RemovedHooksLeaveSharedFiles(t *testing.T) {
	root := seedCharter(t)
	write(t, root, map[string]string{".cursor/hooks.json": `{"version":1,"hooks":{"stop":[{"command":"./mine.sh"}]}}`})
	hosts := Enabled([]string{"cursor", "github-copilot"})
	first, err := Apply(root, Plan(walk(t, root), hosts).Outputs, nil, Options{Mode: LinkSymlink})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, ".agents/hooks/build.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := Apply(root, Plan(walk(t, root), hosts).Outputs, first.Owned, Options{Mode: LinkSymlink}); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{".claude/settings.json", ".cursor/hooks.json"} {
		data, _ := os.ReadFile(filepath.Join(root, f))
		if strings.Contains(string(data), "keystone hook fire") {
			t.Errorf("%s still bridges:\n%s", f, data)
		}
	}
	if data, _ := os.ReadFile(filepath.Join(root, ".cursor/hooks.json")); !strings.Contains(string(data), "./mine.sh") {
		t.Error("user Cursor hook must stay")
	}
	if _, err := os.Stat(filepath.Join(root, ".github/hooks/keystone.json")); !os.IsNotExist(err) {
		t.Error("owned Copilot bridge file should be pruned")
	}
}
