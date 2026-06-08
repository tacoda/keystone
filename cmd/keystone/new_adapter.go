package main

import (
	"fmt"
	"path/filepath"
)

func runNewAdapter(args []string) error {
	projectDir, harnessRoot, remaining, err := parseDirAndHarnessRoot(args)
	if err != nil {
		return err
	}
	if len(remaining) != 1 {
		return fmt.Errorf("`keystone new adapter` requires exactly one agent argument")
	}
	agent := remaining[0]
	base := filepath.Join(projectDir, harnessRoot, "adapters", agent)

	files := map[string]string{
		"activation.md": fmt.Sprintf(`# Adapter — %s — activation

What this agent reads at session start (the menu file). Describe how the
agent is told to load the harness — e.g. the file path it picks up
automatically, any required configuration to point it at the harness.
`, agent),
		"lifecycle.md": fmt.Sprintf(`# Adapter — %s — lifecycle

How this agent invokes playbooks and actions. List the invocation
incantation per action: command line, slash command, chat directive, or
file edit pattern.
`, agent),
		"sensors.md": fmt.Sprintf(`# Adapter — %s — sensors

How this agent invokes sensors and consumes their output. Note any
capability gaps and document the fallback for each (configure the agent
or maintain a small workaround file under harness/adapters/%s/).
`, agent, agent),
	}

	for relname, body := range files {
		path := filepath.Join(base, relname)
		if err := writeSkeleton(path, body); err != nil {
			return err
		}
	}
	return nil
}
