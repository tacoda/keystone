package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// runNewPlugin scaffolds a NEW PLUGIN REPO — a separate directory the
// author will publish to git. Different from the in-project generators,
// which write inside an existing harness.
func runNewPlugin(args []string) error {
	dir := "."
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
		case strings.HasPrefix(a, "-"):
			return fmt.Errorf("unknown flag %s", a)
		default:
			positional = append(positional, a)
		}
	}
	if len(positional) != 1 {
		return fmt.Errorf("`keystone new plugin` requires exactly one name argument")
	}
	name := positional[0]

	parent, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}
	pluginRoot := filepath.Join(parent, name)
	if _, err := os.Stat(pluginRoot); err == nil {
		return fmt.Errorf("%s already exists", pluginRoot)
	} else if !os.IsNotExist(err) {
		return err
	}

	manifest := map[string]any{
		"name":        name,
		"version":     "0.1.0",
		"description": "TODO: describe what this plugin ships.",
	}
	manifestBytes, _ := json.MarshalIndent(manifest, "", "  ")
	manifestBytes = append(manifestBytes, '\n')

	files := map[string]string{
		"keystone-plugin.json": string(manifestBytes),
		"README.md": fmt.Sprintf(`# %s

A keystone plugin. Pin from a consumer's keystone.json:

`+"```json"+`
{
  "plugins": [
    { "name": "%s", "source": "<owner>/%s", "version": "0.1.0" }
  ]
}
`+"```"+`

## Layout

`+"```"+`
keystone-plugin.json   # name, version, optional strict map
guides/<topic>/        # rules; loaded ambient
corpus/<topic>/        # paired reasoning; on-demand
sensors/               # automated checks
actions/               # lifecycle units
playbooks/             # ordered action sequences
adapters/<agent>/      # per-agent bindings (optional)
`+"```"+`

Edit this README to describe what the plugin ships and how consumers
should use it. Then commit, tag a version, and publish.
`, name, name, name),
		"guides/.gitkeep":   "",
		"corpus/.gitkeep":   "",
		"sensors/.gitkeep":  "",
		"actions/.gitkeep":  "",
		"playbooks/.gitkeep": "",
	}

	for rel, body := range files {
		path := filepath.Join(pluginRoot, rel)
		if err := writeSkeleton(path, body); err != nil {
			return err
		}
	}

	fmt.Fprintf(os.Stdout, "\n▶ Next: `cd %s`, edit the README + keystone-plugin.json, drop content into guides/corpus/sensors/..., then `git init && git tag v0.1.0` and publish.\n", name)
	return nil
}
