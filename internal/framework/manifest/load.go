package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Load reads and parses the plugin manifest from pluginRoot. Returns a
// validated *Manifest on success.
func Load(pluginRoot string) (*Manifest, error) {
	path := filepath.Join(pluginRoot, PluginManifestFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", PluginManifestFile, err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse %s: %w", PluginManifestFile, err)
	}
	if err := m.validate(); err != nil {
		return nil, err
	}
	return &m, nil
}
