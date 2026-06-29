package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestOpencodeMCPConfig_WritesLocalServer(t *testing.T) {
	dir := t.TempDir()
	path, body, err := opencodeMCPConfig(dir)
	if err != nil {
		t.Fatalf("opencodeMCPConfig: %v", err)
	}
	if got := filepath.Base(path); got != "opencode.json" {
		t.Errorf("path = %s, want opencode.json", got)
	}

	var doc map[string]any
	if err := json.Unmarshal(body, &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if doc["$schema"] != "https://opencode.ai/config.json" {
		t.Errorf("missing $schema, got %v", doc["$schema"])
	}
	srv := doc["mcp"].(map[string]any)["keystone"].(map[string]any)
	if srv["type"] != "local" {
		t.Errorf("type = %v, want local", srv["type"])
	}
	if srv["enabled"] != true {
		t.Errorf("enabled = %v, want true", srv["enabled"])
	}
	cmd := srv["command"].([]any)
	if len(cmd) == 0 || cmd[0] != "keystone" {
		t.Errorf("command = %v, want [keystone mcp serve ...]", cmd)
	}
}

func TestOpencodeMCPConfig_PreservesExistingKeys(t *testing.T) {
	dir := t.TempDir()
	existing := `{"$schema":"x","model":"anthropic/claude","mcp":{"other":{"type":"local"}}}`
	if err := os.WriteFile(filepath.Join(dir, "opencode.json"), []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}
	_, body, err := opencodeMCPConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(body, &doc); err != nil {
		t.Fatal(err)
	}
	if doc["model"] != "anthropic/claude" {
		t.Errorf("model key dropped: %v", doc["model"])
	}
	mcp := doc["mcp"].(map[string]any)
	if _, ok := mcp["other"]; !ok {
		t.Errorf("existing mcp server 'other' dropped")
	}
	if _, ok := mcp["keystone"]; !ok {
		t.Errorf("keystone server not added")
	}
}
