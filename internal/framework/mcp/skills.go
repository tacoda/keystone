package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	kconfig "github.com/tacoda/keystone/internal/framework/config"
)

// registerSkillResources auto-discovers every SKILL.md under
// .charter/skills/ and exposes them as MCP resources.
// Hosts like Claude Code that auto-load `skill://` URIs pick these up
// without further configuration.
//
// The URI form is:
//
//	skill://<disk-name>/SKILL.md
//
// where <disk-name> is the directory name on disk (colons in the
// canonical id are normalized to hyphens — see new_skill.go). The
// frontmatter id inside the file preserves the original colon form.
func registerSkillResources(s *server.MCPServer, projectDir string) {
	// List resource: every skill's URI + descriptor.
	s.AddResource(
		mcp.NewResource("skill://list",
			"Project-local skills",
			mcp.WithResourceDescription("Every SKILL.md under .charter/skills/. URI form: skill://<name>/SKILL.md."),
			mcp.WithMIMEType("application/json"),
		),
		func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			skills, err := walkSkills(projectDir)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]any, 0, len(skills))
			for _, sk := range skills {
				out = append(out, map[string]any{
					"name":        sk.diskName,
					"uri":         "skill://" + sk.diskName + "/SKILL.md",
					"path":        sk.relPath,
					"description": sk.description,
				})
			}
			body, _ := json.MarshalIndent(map[string]any{
				"count":  len(out),
				"skills": out,
			}, "", "  ")
			return []mcp.ResourceContents{
				mcp.TextResourceContents{
					URI:      req.Params.URI,
					MIMEType: "application/json",
					Text:     string(body),
				},
			}, nil
		},
	)

	// Per-skill body resource. URI template matches the canonical form
	// host agents expect.
	s.AddResourceTemplate(
		mcp.NewResourceTemplate("skill://{name}/SKILL.md",
			"Project-local skill body",
			mcp.WithTemplateDescription("Returns the full SKILL.md body for one project-local skill. Host agents (Claude Code, Cursor) typically auto-load these."),
			mcp.WithTemplateMIMEType("text/markdown"),
		),
		func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			name, err := parseSkillURI(req.Params.URI)
			if err != nil {
				return nil, err
			}
			skills, err := walkSkills(projectDir)
			if err != nil {
				return nil, err
			}
			for _, sk := range skills {
				if sk.diskName == name {
					body, err := os.ReadFile(filepath.Join(projectDir, sk.relPath))
					if err != nil {
						return nil, fmt.Errorf("read %s: %w", sk.relPath, err)
					}
					return []mcp.ResourceContents{
						mcp.TextResourceContents{
							URI:      req.Params.URI,
							MIMEType: "text/markdown",
							Text:     string(body),
						},
					}, nil
				}
			}
			return nil, fmt.Errorf("no skill named %q", name)
		},
	)
}

// skillEntry is the indexer's view of one SKILL.md.
type skillEntry struct {
	diskName    string // dir name on disk (e.g. "keystone-index")
	relPath     string // path relative to projectDir
	description string // pulled from frontmatter if present
}

// walkSkills scans .charter/skills/ for SKILL.md files.
// Returns an empty slice if the directory does not exist (a fresh
// install without authored skills still serves the resource API).
func walkSkills(projectDir string) ([]skillEntry, error) {
	root := filepath.Join(projectDir, kconfig.DefaultCharterRoot, "skills")
	info, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("stat %s: %w", root, err)
	}
	if !info.IsDir() {
		return nil, nil
	}

	var out []skillEntry
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Base(path) != "SKILL.md" {
			return nil
		}
		parent := filepath.Base(filepath.Dir(path))
		rel, _ := filepath.Rel(projectDir, path)
		entry := skillEntry{
			diskName: parent,
			relPath:  filepath.ToSlash(rel),
		}
		// Parse minimal frontmatter for the description; ignore parse
		// errors (the resource works even without a valid descriptor).
		if raw, err := os.ReadFile(path); err == nil {
			if desc := skillDescriptionFromFrontmatter(string(raw)); desc != "" {
				entry.description = desc
			}
		}
		out = append(out, entry)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// skillDescriptionFromFrontmatter pulls the `description:` value out of
// a SKILL.md's frontmatter without a full YAML parse. Returns "" when
// the field is absent.
func skillDescriptionFromFrontmatter(body string) string {
	if !strings.HasPrefix(body, "---\n") && !strings.HasPrefix(body, "---\r\n") {
		return ""
	}
	end := strings.Index(body[4:], "\n---")
	if end < 0 {
		return ""
	}
	fm := body[4 : 4+end]
	for _, line := range strings.Split(fm, "\n") {
		s := strings.TrimSpace(line)
		if !strings.HasPrefix(s, "description:") {
			continue
		}
		v := strings.TrimSpace(strings.TrimPrefix(s, "description:"))
		v = strings.Trim(v, `'"`)
		return v
	}
	return ""
}

func parseSkillURI(uri string) (string, error) {
	const prefix = "skill://"
	if !strings.HasPrefix(uri, prefix) {
		return "", fmt.Errorf("URI must start with %s", prefix)
	}
	rest := strings.TrimPrefix(uri, prefix)
	rest = strings.TrimSuffix(rest, "/SKILL.md")
	if rest == "" {
		return "", fmt.Errorf("missing skill name in URI: %s", uri)
	}
	return rest, nil
}
