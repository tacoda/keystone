package main

import (
	"os"
	"path/filepath"
)

// supportedAgents is derived from the catalog so the two stay in sync.
// Exposed as a slice for error messages.
func supportedAgents() []string {
	cat := categoryByID("agent")
	out := make([]string, 0, len(cat.Values))
	for _, v := range cat.Values {
		out = append(out, v.ID)
	}
	return out
}

func isSupportedAgent(name string) bool {
	cat := categoryByID("agent")
	return cat.isValidValue(name)
}

// detectAgent inspects dir for agent-specific marker files and returns the
// matching agent name, or "" if none match.
func detectAgent(dir string) string {
	exists := func(rel string) bool {
		_, err := os.Stat(filepath.Join(dir, rel))
		return err == nil
	}

	switch {
	case exists("CLAUDE.md") || exists(".claude"):
		return "claude-code"
	case exists("AGENTS.md") && exists(".pi"):
		return "pi"
	case exists(".github/copilot-instructions.md"):
		return "github-copilot"
	case exists(".opencode") || exists("opencode.json"):
		return "opencode"
	case exists(".cursor"):
		return "cursor"
	case exists(".aider.conf.yml") || exists("CONVENTIONS.md"):
		return "aider"
	case exists("AGENTS.md"):
		return "codex"
	case exists(".continuerules"):
		return "continue"
	case exists(".goosehints"):
		return "goose"
	}
	return ""
}
