package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerSourceTools wires the external-source MCP tools — stage 3
// of the runtime resolution flow. Sources are configured in
// .keystone/context.json; adapters live in adapter.go.
func registerSourceTools(s *server.MCPServer, projectDir string) {
	s.AddTool(
		mcp.NewTool("keystone_source_list",
			mcp.WithDescription("List every external source configured in .keystone/context.json plus its current health. Empty list means no sources configured — stage 3 of the resolution flow is unavailable until one is added."),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			adapters, err := loadAdapters(projectDir)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out := make([]map[string]any, 0, len(adapters))
			for _, a := range adapters {
				h := a.Health(ctx)
				out = append(out, map[string]any{
					"name":   a.Name(),
					"type":   a.Type(),
					"health": h,
				})
			}
			body, _ := json.MarshalIndent(map[string]any{
				"count":   len(out),
				"sources": out,
			}, "", "  ")
			return mcp.NewToolResultText(string(body)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("keystone_source_query",
			mcp.WithDescription("Stage 3 of the runtime resolution flow: query a configured external source when in-harness rules + corpus are insufficient. NEVER apply results silently — surface the finding to the user and ask where to record it."),
			mcp.WithString("source",
				mcp.Required(),
				mcp.Description("Source name from .keystone/context.json."),
			),
			mcp.WithString("query",
				mcp.Required(),
				mcp.Description("Free-form query string. Adapter-specific shape."),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name, err := req.RequireString("source")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			query, err := req.RequireString("query")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			adapters, err := loadAdapters(projectDir)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			a, err := findAdapter(adapters, name)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			res, err := a.Query(ctx, query)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			body, _ := json.MarshalIndent(map[string]any{
				"source":   a.Name(),
				"type":     a.Type(),
				"query":    query,
				"result":   res,
				"reminder": "stage 3 result — surface to the user; ask 'apply? at what level (project, team, org)?' before recording.",
			}, "", "  ")
			return mcp.NewToolResultText(string(body)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("keystone_source_health",
			mcp.WithDescription("Probe one source's reachability + auth state without running a query."),
			mcp.WithString("source",
				mcp.Required(),
				mcp.Description("Source name from .keystone/context.json."),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name, err := req.RequireString("source")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			adapters, err := loadAdapters(projectDir)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			a, err := findAdapter(adapters, name)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			h := a.Health(ctx)
			body, _ := json.MarshalIndent(map[string]any{
				"source": a.Name(),
				"type":   a.Type(),
				"health": h,
			}, "", "  ")
			return mcp.NewToolResultText(string(body)), nil
		},
	)
}

// registerSourceResources exposes the same source data as MCP
// resources for hosts that prefer resource URIs over tool calls.
func registerSourceResources(s *server.MCPServer, projectDir string) {
	s.AddResource(
		mcp.NewResource("keystone://source/list",
			"External sources configured",
			mcp.WithResourceDescription("Names + healths of every external source declared in .keystone/context.json."),
			mcp.WithMIMEType("application/json"),
		),
		func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			adapters, err := loadAdapters(projectDir)
			if err != nil {
				return nil, err
			}
			entries := make([]map[string]any, 0, len(adapters))
			for _, a := range adapters {
				h := a.Health(ctx)
				entries = append(entries, map[string]any{
					"name":   a.Name(),
					"type":   a.Type(),
					"health": h,
				})
			}
			body, _ := json.MarshalIndent(map[string]any{"sources": entries}, "", "  ")
			return []mcp.ResourceContents{
				mcp.TextResourceContents{
					URI:      req.Params.URI,
					MIMEType: "application/json",
					Text:     string(body),
				},
			}, nil
		},
	)

	s.AddResourceTemplate(
		mcp.NewResourceTemplate("keystone://source/{name}/health",
			"External source health",
			mcp.WithTemplateDescription("Reachability + auth state for one external source."),
			mcp.WithTemplateMIMEType("application/json"),
		),
		func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			name, err := parseSourceHealthURI(req.Params.URI)
			if err != nil {
				return nil, err
			}
			adapters, err := loadAdapters(projectDir)
			if err != nil {
				return nil, err
			}
			a, err := findAdapter(adapters, name)
			if err != nil {
				return nil, err
			}
			h := a.Health(ctx)
			body, _ := json.MarshalIndent(map[string]any{
				"source": a.Name(),
				"type":   a.Type(),
				"health": h,
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
}

func parseSourceHealthURI(uri string) (string, error) {
	const prefix = "keystone://source/"
	if !strings.HasPrefix(uri, prefix) {
		return "", fmt.Errorf("URI must start with %s", prefix)
	}
	rest := strings.TrimPrefix(uri, prefix)
	rest = strings.TrimSuffix(rest, "/health")
	if rest == "" {
		return "", fmt.Errorf("missing source name in URI: %s", uri)
	}
	return rest, nil
}
