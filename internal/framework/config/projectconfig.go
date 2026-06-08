package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ProjectConfigFile is the basename of the project-level keystone config,
// located at the repo root (not inside the harness folder).
const ProjectConfigFile = "keystone.json"

// SchemaVersion is the current keystone.json schema version. Bumped when
// fields are renamed or removed.
const SchemaVersion = "1"

// DefaultPluginHost is the git host prepended when a plugin source uses the
// `<owner>/<repo>` shorthand (e.g. `tacoda/tacoda-org` resolves to
// `https://github.com/tacoda/tacoda-org.git`). Plugins hosted elsewhere
// write the host explicitly (`gitlab.com/acme/policies`).
const DefaultPluginHost = "github.com"

// ProjectConfig is the parsed shape of keystone.json. Records the harness
// folder name, the framework version this project is pinned to, the nested
// plugin tree, and optional per-port budgets (Phase 5).
type ProjectConfig struct {
	Version          string                `json:"version"`
	FrameworkVersion string                `json:"framework_version,omitempty"`
	HarnessRoot      string                `json:"harness_root,omitempty"`
	Plugins          []PluginNode          `json:"plugins"`
	Budgets          map[string]BudgetSpec `json:"budgets,omitempty"`
}

// PluginNode is one node in the nested plugin tree declared in keystone.json.
// Among plugins, plugins nested deeper refine the outer plugins they're
// nested in for non-strict items; the project layer always wins by default.
// The strict map declares per-port absolute locks — strict items cannot be
// overridden by anything (project, sibling plugin, ancestor, descendant).
//
// Source format: `[<host>/]<owner>/<repo>`. When the host is omitted (one
// slash, no dots), DefaultPluginHost is prepended. Examples:
//   - "tacoda/tacoda-org"                   → github.com/tacoda/tacoda-org
//   - "github.com/tacoda/tacoda-org"        → github.com/tacoda/tacoda-org
//   - "gitlab.com/acme/policies"            → gitlab.com/acme/policies
//
// Version is a git ref (tag, branch, or SHA). Required at 1.0.
type PluginNode struct {
	Name     string              `json:"name"`
	Source   string              `json:"source"`
	Version  string              `json:"version"`
	Strict   map[string][]string `json:"strict,omitempty"`
	Children []PluginNode        `json:"children,omitempty"`
}

// BudgetSpec is an optional cap on context loaded for one port. Wired up in
// Phase 5; included here so the schema is stable.
type BudgetSpec struct {
	MaxTokens        int `json:"max_tokens,omitempty"`
	MaxTokensPerLoad int `json:"max_tokens_per_load,omitempty"`
}

// ResolvedHarnessRoot returns the configured harness root, falling back to
// DefaultHarnessRoot when the field is unset. Use this everywhere a command
// needs the harness folder name from a loaded ProjectConfig.
func (c *ProjectConfig) ResolvedHarnessRoot() string {
	if c == nil || c.HarnessRoot == "" {
		return DefaultHarnessRoot
	}
	return c.HarnessRoot
}

// DefaultProjectConfig returns the seed config written by `keystone init`
// when no keystone.json exists yet. Empty plugins[] and the configured
// harness root.
func DefaultProjectConfig(harnessRoot string) *ProjectConfig {
	return &ProjectConfig{
		Version:     SchemaVersion,
		HarnessRoot: harnessRoot,
		Plugins:     []PluginNode{},
	}
}

// ReadProjectConfig loads keystone.json from projectDir. Returns the OS-level
// error untouched when the file does not exist, so callers can distinguish
// "no config yet" from "malformed config" via errors.Is(err, os.ErrNotExist).
func ReadProjectConfig(projectDir string) (*ProjectConfig, error) {
	path := filepath.Join(projectDir, ProjectConfigFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg ProjectConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", ProjectConfigFile, err)
	}
	if cfg.Plugins == nil {
		cfg.Plugins = []PluginNode{}
	}
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validate %s: %w", ProjectConfigFile, err)
	}
	return &cfg, nil
}

// WriteProjectConfig serializes the config to <projectDir>/keystone.json
// with indented JSON and a trailing newline.
func WriteProjectConfig(projectDir string, cfg *ProjectConfig) error {
	if err := cfg.validate(); err != nil {
		return fmt.Errorf("validate %s: %w", ProjectConfigFile, err)
	}
	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal %s: %w", ProjectConfigFile, err)
	}
	out = append(out, '\n')
	path := filepath.Join(projectDir, ProjectConfigFile)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, out, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", ProjectConfigFile, err)
	}
	return nil
}

// ParseShorthand splits a CLI plugin spec of the form
// `[<host>/]<owner>/<repo>[@<version>]` into its source and version pieces.
// Examples:
//   - "tacoda/tacoda-org@0.2.0"        → ("tacoda/tacoda-org",        "0.2.0")
//   - "tacoda/tacoda-org"              → ("tacoda/tacoda-org",        "")
//   - "github.com/acme/policies@v1.0"  → ("github.com/acme/policies", "v1.0")
//
// Validation of the source itself is delegated to ValidateSource — this
// function only handles the @-split.
func ParseShorthand(spec string) (source, version string) {
	if i := strings.LastIndexByte(spec, '@'); i >= 0 {
		return spec[:i], spec[i+1:]
	}
	return spec, ""
}

// DefaultPluginName derives the plugin name from a source string by taking
// the last path segment. `tacoda/tacoda-org` → "tacoda-org";
// `gitlab.com/acme/policies` → "policies". Used by `keystone plugin add`
// when the user does not pass --name.
func DefaultPluginName(source string) string {
	source = strings.TrimSuffix(source, "/")
	if i := strings.LastIndexByte(source, '/'); i >= 0 {
		return source[i+1:]
	}
	return source
}

// ExpandSource expands a shorthand source string into a full git URL.
//   - "tacoda/tacoda-org"        → "https://github.com/tacoda/tacoda-org.git"
//   - "gitlab.com/acme/policies" → "https://gitlab.com/acme/policies.git"
//
// Adds DefaultPluginHost when no host is present (the source has exactly
// one slash and no dots in the leading segment). Always uses https. The
// .git suffix is appended if not already present.
func ExpandSource(source string) string {
	source = strings.TrimSpace(source)
	parts := strings.Split(source, "/")
	if len(parts) >= 2 && !strings.ContainsAny(parts[0], ".") {
		// owner/repo form — prepend host
		source = DefaultPluginHost + "/" + source
	}
	if !strings.HasSuffix(source, ".git") {
		source += ".git"
	}
	return "https://" + source
}

// validation

// pluginNamePattern matches plugin names: lowercase kebab-case, ≤64 chars,
// must start with a letter. Same shape as manifest names so a plugin
// repo's manifest name lines up with what the consumer references in
// keystone.json.
var pluginNamePattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,63}$`)

// sourcePattern matches the source field shape: optional host segment with
// at least one dot, followed by one or more path segments. Two slashes
// minimum when a host is present, one slash minimum for `owner/repo` form.
var sourcePattern = regexp.MustCompile(`^([a-z0-9.-]+\.[a-z]+/)?[a-z0-9_.-]+(/[a-z0-9_.-]+)+$`)

// ValidateSource reports whether a plugin source string parses cleanly. The
// resolver expands the shorthand to a full git URL; this function only
// gates what's allowed in the file.
func ValidateSource(source string) error {
	if source == "" {
		return fmt.Errorf("source is empty")
	}
	if strings.HasPrefix(source, "git+") || strings.HasPrefix(source, "https://") || strings.HasPrefix(source, "http://") {
		return fmt.Errorf("source %q uses 0.x-style URL; 1.0 expects shorthand form `[<host>/]<owner>/<repo>` (e.g. tacoda/tacoda-org)", source)
	}
	if !sourcePattern.MatchString(source) {
		return fmt.Errorf("source %q does not match `[<host>/]<owner>/<repo>`", source)
	}
	return nil
}

func (c *ProjectConfig) validate() error {
	if c.Version == "" {
		return fmt.Errorf("missing required field 'version'")
	}
	seen := map[string]bool{}
	for i := range c.Plugins {
		if err := validatePluginNode(&c.Plugins[i], seen); err != nil {
			return fmt.Errorf("plugins[%d]: %w", i, err)
		}
	}
	return nil
}

func validatePluginNode(n *PluginNode, seen map[string]bool) error {
	if n.Name == "" {
		return fmt.Errorf("missing required field 'name'")
	}
	if !pluginNamePattern.MatchString(n.Name) {
		return fmt.Errorf("name %q must match %s", n.Name, pluginNamePattern)
	}
	if seen[n.Name] {
		return fmt.Errorf("duplicate plugin name %q in tree", n.Name)
	}
	seen[n.Name] = true
	if err := ValidateSource(n.Source); err != nil {
		return fmt.Errorf("plugin %q: %w", n.Name, err)
	}
	if n.Version == "" {
		return fmt.Errorf("plugin %q missing required field 'version'", n.Name)
	}
	for i := range n.Children {
		if err := validatePluginNode(&n.Children[i], seen); err != nil {
			return fmt.Errorf("children[%d]: %w", i, err)
		}
	}
	return nil
}
