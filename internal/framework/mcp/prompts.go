package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerPrompts wires the MCP prompt surface — pre-defined message
// templates the host UI exposes to the user. Each prompt corresponds
// to a framework action playbook and primes the agent with the
// canonical invocation phrasing + a pointer at the playbook path.
//
// Prompts are user-triggered. The host typically surfaces them in a
// "/" menu or palette. The agent receives the rendered messages and
// follows the playbook body.
func registerPrompts(s *server.MCPServer, projectDir string) {
	s.AddPrompt(
		mcp.NewPrompt("keystone_bootstrap",
			mcp.WithPromptDescription("One-time codebase analysis. Detects stack, fills the state ledger, classifies sensors. Run once per project, before the first `task`."),
		),
		func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			return promptResult(
				"keystone_bootstrap",
				"Run the bootstrap action.",
				`Open .charter/actions/bootstrap.md and execute its Activities in order:

- detect language / framework / CI from the repo
- seed corpus/state/CODEBASE_STATE.md and INSTALL_PROFILE.md
- classify every sensor as computational or inferential
- seed idiom globs from the region map

Surface any conversational gaps (aspirational patterns, methodology, migration plans) and ask the user to fill them in. Do NOT invent answers.`,
			), nil
		},
	)

	s.AddPrompt(
		mcp.NewPrompt("keystone_task",
			mcp.WithPromptDescription("End-to-end task workflow. Orchestrates spec → orient → implementation → check-drift → verify → review."),
			mcp.WithArgument("description",
				mcp.ArgumentDescription("Brief description of the unit of work (e.g., 'add rate-limit middleware', 'fix bug PROJ-123')."),
				mcp.RequiredArgument(),
			),
		),
		func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			desc := req.Params.Arguments["description"]
			if desc == "" {
				return nil, fmt.Errorf("missing argument: description")
			}
			return promptResult(
				"keystone_task",
				fmt.Sprintf("Run task on: %s", desc),
				fmt.Sprintf(`Run task on: %s

Open .charter/playbooks/task.md and execute the sequence:
  spec → orient → (implementation) → check-drift → verify → review

Resolve each stage per the runtime-resolution flow (rules → corpus → external → ask). At any iron-law violation, halt and surface to the user.`, desc),
			), nil
		},
	)

	s.AddPrompt(
		mcp.NewPrompt("keystone_audit",
			mcp.WithPromptDescription("Periodic dual-flywheel: learning + pruning. Walks the inbox, promotes accepted candidates, archives stale guides."),
		),
		func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			return promptResult(
				"keystone_audit",
				"Run the audit action.",
				`Open .charter/actions/audit.md and execute its Activities. Drive the dual flywheel:
  - Learning: walk learning/inbox/, propose promotions to corpus/guides
  - Pruning: identify stale guides/corpus, propose archive moves

Surface every proposed move to the user before applying. Never delete; archive with a one-line reason.`,
			), nil
		},
	)

	s.AddPrompt(
		mcp.NewPrompt("keystone_learn",
			mcp.WithPromptDescription("Capture a learning candidate from a surprise, incident, or review finding. Writes to learning/inbox/."),
			mcp.WithArgument("finding",
				mcp.ArgumentDescription("The observation, surprise, or insight to record. Be concrete — what happened, why it matters, what rule it implies."),
				mcp.RequiredArgument(),
			),
		),
		func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			finding := req.Params.Arguments["finding"]
			if finding == "" {
				return nil, fmt.Errorf("missing argument: finding")
			}
			return promptResult(
				"keystone_learn",
				fmt.Sprintf("Capture learning finding: %s", truncate(finding, 60)),
				fmt.Sprintf(`Run the learn action with this finding:

%s

Open .charter/actions/learn.md and write a candidate to learning/inbox/<timestamp>-<slug>.md following the canonical shape (captured/source/proposed-layer/proposed-globs frontmatter + What happened / Why it matters / Proposed change body).

Promotion to corpus/guides happens in synthesize, not here.`, finding),
			), nil
		},
	)
}

// promptResult is the standard single-message envelope. Every shipped
// prompt returns one user-role text message; the agent reads it and
// follows the embedded instructions.
func promptResult(name, title, body string) *mcp.GetPromptResult {
	return &mcp.GetPromptResult{
		Description: title,
		Messages: []mcp.PromptMessage{
			mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(body)),
		},
	}
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
