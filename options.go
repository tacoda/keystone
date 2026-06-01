package main

// Category describes one user-selectable dimension of the install.
// Every selection is keyed by ID (kebab-case); Label is shown in prompts.
type Category struct {
	ID          string        // kebab-case identifier; also used as flag name
	Label       string        // human-readable prompt title
	Description string        // one-line subtitle shown in the prompt
	MultiSelect bool          // true → user may pick zero or more; false → exactly one
	Required    bool          // true → install fails if unset and no TTY for prompting
	Values      []OptionValue // allowed labels, in display order
}

// OptionValue is one selectable label inside a Category.
type OptionValue struct {
	ID          string // kebab-case identifier — written to profile and used for conditional installs
	Description string // one-line hint shown next to the label in the prompt
}

// categories lists every option the installer can record, in prompt order.
// Order matters: agent comes first because it gates downstream behavior;
// language/architecture/etc. come later because they describe the project.
var categories = []Category{
	{
		ID:          "agent",
		Label:       "Agent",
		Description: "Which AI coding agent will read this harness.",
		Required:    true,
		Values: []OptionValue{
			{"claude-code", "Anthropic Claude Code (CLI / IDE extensions)"},
			{"codex", "OpenAI Codex / Codex CLI"},
			{"cursor", "Cursor editor"},
			{"aider", "Aider"},
			{"github-copilot", "GitHub Copilot (VS Code + CLI)"},
			{"continue", "Continue.dev"},
			{"cline", "Cline (VS Code)"},
			{"goose", "Block Goose"},
			{"pi", "Pi"},
			{"generic", "No specific agent — install the corpus only"},
		},
	},
	{
		ID:          "app-type",
		Label:       "Application type",
		Description: "What kind of thing is being built.",
		Values: []OptionValue{
			{"web-application", "Server-rendered or SPA web app"},
			{"web-api", "HTTP/JSON or gRPC service, no UI"},
			{"cli-tool", "Command-line tool"},
			{"library", "Library / SDK consumed by other code"},
			{"mobile-app", "iOS / Android / cross-platform mobile"},
			{"desktop-app", "Native or Electron desktop app"},
			{"data-pipeline", "Batch / streaming data processing"},
			{"embedded", "Firmware / embedded device"},
			{"other", "Something else (write it into the profile by hand)"},
		},
	},
	{
		ID:          "architecture",
		Label:       "Architecture preferences",
		Description: "Opinionated patterns to surface in the harness. Multiple allowed.",
		MultiSelect: true,
		Values: []OptionValue{
			{"hexagonal", "Hexagonal / ports and adapters (Cockburn)"},
			{"clean-architecture", "Clean architecture (Martin)"},
			{"onion-architecture", "Onion architecture (Palermo)"},
			{"layered", "Layered (presentation / application / domain / persistence)"},
			{"mvc", "Model-View-Controller"},
			{"mvvm", "Model-View-ViewModel"},
			{"event-driven", "Event-driven / event sourcing"},
			{"microservices", "Microservices"},
			{"monolith", "Modular monolith"},
			{"serverless", "Serverless / FaaS"},
			{"spa", "Single-page application (frontend)"},
			{"continuous-delivery", "Continuous Delivery (Humble & Farley)"},
			{"none-yet", "No declared architecture (yet)"},
		},
	},
	{
		ID:          "testing",
		Label:       "Testing approach",
		Description: "Disciplines the team practices. Multiple allowed.",
		MultiSelect: true,
		Values: []OptionValue{
			{"tdd", "Test-driven development"},
			{"bdd", "Behavior-driven development"},
			{"atdd", "Acceptance test-driven development"},
			{"property-based", "Property-based testing"},
			{"snapshot", "Snapshot testing"},
			{"manual-qa", "Manual QA"},
			{"none-yet", "No declared testing discipline (yet)"},
		},
	},
	{
		ID:          "compliance",
		Label:       "Compliance scope",
		Description: "Regimes the system must satisfy. Multiple allowed.",
		MultiSelect: true,
		Values: []OptionValue{
			{"gdpr", "GDPR"},
			{"hipaa", "HIPAA"},
			{"pci-dss", "PCI DSS"},
			{"soc2", "SOC 2"},
			{"fedramp", "FedRAMP"},
			{"none", "None applicable"},
		},
	},
}

// categoryByID looks up a category. Returns nil if not found.
func categoryByID(id string) *Category {
	for i := range categories {
		if categories[i].ID == id {
			return &categories[i]
		}
	}
	return nil
}

// isValidValue reports whether v is a known label in cat.
func (c *Category) isValidValue(v string) bool {
	for _, ov := range c.Values {
		if ov.ID == v {
			return true
		}
	}
	return false
}

// Selections is the resolved set of choices for an install, keyed by category ID.
// Single-select categories store a one-element slice for uniformity.
type Selections map[string][]string
