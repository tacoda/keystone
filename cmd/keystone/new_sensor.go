package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/tacoda/keystone/internal/framework/config"
)

func runNewSensor(args []string) error {
	dir := "."
	kind := "computational"
	var positional []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--dir":
			if i+1 >= len(args) {
				return fmt.Errorf("flag %s requires a value", a)
			}
			dir = args[i+1]
			i++
		case strings.HasPrefix(a, "--dir="):
			dir = strings.TrimPrefix(a, "--dir=")
		case a == "--kind":
			if i+1 >= len(args) {
				return fmt.Errorf("flag %s requires a value", a)
			}
			kind = args[i+1]
			i++
		case strings.HasPrefix(a, "--kind="):
			kind = strings.TrimPrefix(a, "--kind=")
		default:
			positional = append(positional, a)
		}
	}
	if len(positional) != 1 {
		return fmt.Errorf("`keystone new sensor` requires exactly one name argument")
	}
	name := positional[0]
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}
	harnessRoot := config.DefaultHarnessRoot
	path := filepath.Join(absDir, harnessRoot, "sensors", name+".md")
	body := fmt.Sprintf(`---
kind: %s
---

# Sensor: %s

One-sentence description of what this sensor checks.

## Command

`+"```"+`
<shell invocation; e.g. go vet ./...>
`+"```"+`

## Interpretation

Exit 0 = pass. Non-zero exit = fail; capture stderr.

## Remediation

On fail, hand the captured output to the agent and request a fix before
proceeding.
`, kind, name)
	return writeSkeleton(path, body)
}
