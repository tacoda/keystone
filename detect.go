package main

import (
	"os"
	"path/filepath"
)

var supportedAgents = []string{
	"claude-code",
	"codex",
	"pi",
	"cursor",
	"aider",
	"github-copilot-cli",
	"continue",
	"cline",
	"goose",
	"_generic",
}

func isSupportedAgent(name string) bool {
	for _, a := range supportedAgents {
		if a == name {
			return true
		}
	}
	return false
}

// detectAgent inspects dir for agent-specific marker files and returns the
// matching agent name, or "" if none match. Order mirrors install.sh.
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
		return "github-copilot-cli"
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
