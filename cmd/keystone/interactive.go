package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
)

// isTerminal reports whether f is a real TTY. Uses an ioctl rather
// than the character-device check, since /dev/null is a character
// device but not a TTY.
func isTerminal(f *os.File) bool {
	return term.IsTerminal(int(f.Fd()))
}

// promptAgent is the ONE prompt `keystone init` runs at 2.0. Lists
// every supported agent target with a number; user types the
// number (or the agent name, or just hits Enter for the default).
// Returns the chosen target name, or "" if the user picked default
// — caller falls back to `_generic`.
//
// Deliberately no huh dependency. The charter should bootstrap with
// the smallest possible interaction surface; later questions live
// in the bootstrap action, where the agent can ask them
// conversationally against the actual codebase.
func promptAgent(in io.Reader, out io.Writer) (string, error) {
	options := promptAgentOptions()
	fmt.Fprintln(out, "Pick an agent target — type the number or the name, or press Enter for the default.")
	for i, opt := range options {
		fmt.Fprintf(out, "  %d) %-15s %s\n", i+1, opt.id, opt.desc)
	}
	fmt.Fprintf(out, "\n  [Enter = %s] > ", defaultAgent)

	scanner := bufio.NewScanner(in)
	if !scanner.Scan() {
		return "", nil
	}
	pick := strings.TrimSpace(scanner.Text())
	if pick == "" {
		return defaultAgent, nil
	}
	// Numeric pick.
	if n, err := strconv.Atoi(pick); err == nil {
		if n < 1 || n > len(options) {
			return "", fmt.Errorf("pick out of range: %d (expected 1..%d)", n, len(options))
		}
		return options[n-1].id, nil
	}
	// Name pick — must match an entry.
	for _, opt := range options {
		if opt.id == pick {
			return pick, nil
		}
	}
	return "", fmt.Errorf("unknown agent %q (supported: %v)", pick, supportedAgents())
}

const defaultAgent = "generic"

type agentOption struct {
	id, desc string
}

// promptAgentOptions returns the picker entries shown at init time.
// Order matches the catalog; `_generic` lives at the bottom as the
// "no specific host" escape hatch.
func promptAgentOptions() []agentOption {
	descs := map[string]string{
		"claude-code":    "Anthropic Claude Code (CLI)",
		"codex":          "OpenAI Codex / Codex CLI",
		"cursor":         "Cursor editor",
		"aider":          "aider CLI",
		"cline":          "Cline (VS Code agent)",
		"continue":       "Continue (VS Code / JetBrains)",
		"goose":          "Block's Goose",
		"github-copilot": "GitHub Copilot (chat + workspace)",
		"opencode":       "opencode (open source CLI)",
		"pi":             "pi.dev",
		"generic":        "no specific host — generic AGENTS.md",
	}
	supported := supportedAgents()
	out := make([]agentOption, 0, len(supported))
	for _, a := range supported {
		desc := descs[a]
		if desc == "" {
			desc = a
		}
		out = append(out, agentOption{id: a, desc: desc})
	}
	return out
}
