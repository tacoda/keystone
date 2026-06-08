package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Load reads and parses the plugin manifest from policyRoot. Returns a
// validated *Manifest on success.
func Load(policyRoot string) (*Manifest, error) {
	path := filepath.Join(policyRoot, PolicyManifestFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", PolicyManifestFile, err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse %s: %w", PolicyManifestFile, err)
	}
	if err := m.validate(); err != nil {
		return nil, err
	}
	return &m, nil
}
