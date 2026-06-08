package main

import (
	"fmt"
	"path/filepath"
)

func runNewAction(args []string) error {
	projectDir, harnessRoot, remaining, err := parseDirAndHarnessRoot(args)
	if err != nil {
		return err
	}
	if len(remaining) != 1 {
		return fmt.Errorf("`keystone new action` requires exactly one name argument")
	}
	name := remaining[0]
	path := filepath.Join(projectDir, harnessRoot, "actions", name+".md")
	body := fmt.Sprintf(`# Action: %s

One-sentence description.

## Entry condition

What must be true before this action runs.

## Activities

1. <verb> + <artifact>.
2. <verb> + <artifact>.
3. <verb> + <artifact>.

## Exit condition

What must be true when this action completes.
`, name)
	return writeSkeleton(path, body)
}

func runNewPlaybook(args []string) error {
	projectDir, harnessRoot, remaining, err := parseDirAndHarnessRoot(args)
	if err != nil {
		return err
	}
	if len(remaining) != 1 {
		return fmt.Errorf("`keystone new playbook` requires exactly one name argument")
	}
	name := remaining[0]
	path := filepath.Join(projectDir, harnessRoot, "playbooks", name+".md")
	body := fmt.Sprintf(`# Playbook: %s

One-sentence description.

## Sequence

1. **<action-name>** — why this step.
2. **<action-name>** — why this step.
3. **<action-name>** — why this step.

## Halt conditions

- When <sensor or action> returns fail.
- When the user provides ambiguous input.
`, name)
	return writeSkeleton(path, body)
}
