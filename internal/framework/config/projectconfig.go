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

// SchemaVersion is the current keystone.json schema version. Bumped
// when fields are renamed or removed. 2.0 dropped `harness_root` (the
// layout is now fixed at .keystone/harness/); existing v1 installs are
// upgraded by `keystone migrate`.
const SchemaVersion = "2"

// DefaultPolicyHost is the git host prepended when a policy source uses the
// `<owner>/<repo>` shorthand (e.g. `tacoda/tacoda-org` resolves to
// `https://github.com/tacoda/tacoda-org.git`). Policies hosted elsewhere
// write the host explicitly (`gitlab.com/acme/policies`).
const DefaultPolicyHost = "github.com"

// ProjectConfig is the parsed shape of keystone.json. Records the
// framework version this project is pinned to, the nested policy tree,
// and optional per-port budgets.
//
// 2.0 note: the harness layout is no longer configurable. Older
// installs whose keystone.json carries a `harness_root` field continue
// to parse cleanly — the field is decoded but ignored. The migrator
// strips it on its way through.
type ProjectConfig struct {
	Version          string                `json:"version"`
	FrameworkVersion string                `json:"framework_version,omitempty"`
	HarnessRoot      string                `json:"harness_root,omitempty"` // Deprecated: ignored at 2.0; stripped by `keystone migrate`.
	Policies          []PolicyNode          `json:"policies"`
	Budgets          map[string]BudgetSpec `json:"budgets,omitempty"`
}

// UnmarshalJSON adds backward-compat for pre-2.0 configs whose
// keystone.json carries a `plugins` field instead of `policies`.
// Writes always emit the new field name; reads tolerate the old one
// so unmigrated installs degrade gracefully (per the migrations
// no-breaking-changes invariant).
func (c *ProjectConfig) UnmarshalJSON(data []byte) error {
	type alias ProjectConfig
	aux := &struct {
		*alias
		LegacyPlugins []PolicyNode `json:"plugins,omitempty"`
	}{alias: (*alias)(c)}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if len(c.Policies) == 0 && len(aux.LegacyPlugins) > 0 {
		c.Policies = aux.LegacyPlugins
	}
	return nil
}

// PolicyNode is one node in the nested policy tree declared in keystone.json.
// Among policies, policies nested deeper refine the outer policies they're
// nested in for non-strict items; the project layer always wins by default.
// The strict map declares per-port absolute locks — strict items cannot be
// overridden by anything (project, sibling policy, ancestor, descendant).
//
// Source format: `[<host>/]<owner>/<repo>`. When the host is omitted (one
// slash, no dots), DefaultPolicyHost is prepended. Examples:
//   - "tacoda/tacoda-org"                   → github.com/tacoda/tacoda-org
//   - "github.com/tacoda/tacoda-org"        → github.com/tacoda/tacoda-org
//   - "gitlab.com/acme/policies"            → gitlab.com/acme/policies
//
// Version is a git ref (tag, branch, or SHA). Required at 1.0.
type PolicyNode struct {
	Name     string              `json:"name"`
	Source   string              `json:"source"`
	Version  string              `json:"version"`
	Strict   map[string][]string `json:"strict,omitempty"`
	Children []PolicyNode        `json:"children,omitempty"`
}

// BudgetSpec is an optional cap on context loaded for one port. Wired up in
// Phase 5; included here so the schema is stable.
type BudgetSpec struct {
	MaxTokens        int `json:"max_tokens,omitempty"`
	MaxTokensPerLoad int `json:"max_tokens_per_load,omitempty"`
}

// ResolvedHarnessRoot always returns the fixed harness path at 2.0.
//
// Deprecated: callers should reference config.DefaultHarnessRoot
// directly. Retained so existing call sites compile while they migrate.
func (c *ProjectConfig) ResolvedHarnessRoot() string {
	return DefaultHarnessRoot
}

// DefaultProjectConfig returns the seed config written by `keystone init`
// when no keystone.json exists yet. The harness root is no longer a
// per-project setting at 2.0 — the parameter is retained for back-compat
// with existing call sites but is dropped from the emitted config.
func DefaultProjectConfig(_ string) *ProjectConfig {
	return &ProjectConfig{
		Version: SchemaVersion,
		Policies: []PolicyNode{},
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
	if cfg.Policies == nil {
		cfg.Policies = []PolicyNode{}
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

// ParseShorthand splits a CLI policy spec of the form
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

// DefaultPolicyName derives the policy name from a source string by taking
// the last path segment. `tacoda/tacoda-org` → "tacoda-org";
// `gitlab.com/acme/policies` → "policies". Used by `keystone policy add`
// when the user does not pass --name.
func DefaultPolicyName(source string) string {
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
// Adds DefaultPolicyHost when no host is present (the source has exactly
// one slash and no dots in the leading segment). Always uses https. The
// .git suffix is appended if not already present.
func ExpandSource(source string) string {
	source = strings.TrimSpace(source)
	parts := strings.Split(source, "/")
	if len(parts) >= 2 && !strings.ContainsAny(parts[0], ".") {
		// owner/repo form — prepend host
		source = DefaultPolicyHost + "/" + source
	}
	if !strings.HasSuffix(source, ".git") {
		source += ".git"
	}
	return "https://" + source
}

// validation

// policyNamePattern matches policy names: lowercase kebab-case, ≤64 chars,
// must start with a letter. Same shape as manifest names so a policy
// repo's manifest name lines up with what the consumer references in
// keystone.json.
var policyNamePattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,63}$`)

// sourcePattern matches the source field shape: optional host segment with
// at least one dot, followed by one or more path segments. Two slashes
// minimum when a host is present, one slash minimum for `owner/repo` form.
var sourcePattern = regexp.MustCompile(`^([a-z0-9.-]+\.[a-z]+/)?[a-z0-9_.-]+(/[a-z0-9_.-]+)+$`)

// ValidateSource reports whether a policy source string parses cleanly. The
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
	for i := range c.Policies {
		if err := validatePolicyNode(&c.Policies[i], seen); err != nil {
			return fmt.Errorf("policies[%d]: %w", i, err)
		}
	}
	return nil
}

func validatePolicyNode(n *PolicyNode, seen map[string]bool) error {
	if n.Name == "" {
		return fmt.Errorf("missing required field 'name'")
	}
	if !policyNamePattern.MatchString(n.Name) {
		return fmt.Errorf("name %q must match %s", n.Name, policyNamePattern)
	}
	if seen[n.Name] {
		return fmt.Errorf("duplicate policy name %q in tree", n.Name)
	}
	seen[n.Name] = true
	if err := ValidateSource(n.Source); err != nil {
		return fmt.Errorf("policy %q: %w", n.Name, err)
	}
	if n.Version == "" {
		return fmt.Errorf("policy %q missing required field 'version'", n.Name)
	}
	for i := range n.Children {
		if err := validatePolicyNode(&n.Children[i], seen); err != nil {
			return fmt.Errorf("children[%d]: %w", i, err)
		}
	}
	return nil
}
