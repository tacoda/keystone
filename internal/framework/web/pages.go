package web

import (
	"net/http"

	"github.com/tacoda/keystone/internal/framework/config"
)

// handlePrimitivesNew renders the new-primitive form. GET only.
// The form posts to /web/actions/primitives/new.
func (s *server) handlePrimitivesNew(w http.ResponseWriter, r *http.Request) {
	s.render(w, "primitive_new.html", map[string]any{
		"ProjectDir": s.projectDir,
		// Framework abstractions are shown first (user is encouraged to
		// reach for these); agent abstractions are the extension surface.
		"FrameworkKinds": []string{"guide", "corpus", "sensor", "action", "playbook"},
		"AgentKinds":     []string{"rule", "skill", "subagent", "command"},
	})
}

// handlePolicies lists every entry in keystone.json's `policies:`
// array and the add/remove forms.
func (s *server) handlePolicies(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.ReadProjectConfig(s.projectDir)
	policies := []config.PolicyNode{}
	if err == nil && cfg != nil {
		policies = cfg.Policies
	}
	s.render(w, "policies.html", map[string]any{
		"ProjectDir": s.projectDir,
		"Policies":   policies,
	})
}

// handleSourcesNew renders the new-source form.
func (s *server) handleSourcesNew(w http.ResponseWriter, r *http.Request) {
	s.render(w, "source_new.html", map[string]any{
		"ProjectDir": s.projectDir,
		// Types the built-in registry knows about.
		"Types": []string{"folder", "url"},
	})
}

// handleVerifyPage renders the verify dashboard. The "run" button
// posts to /web/actions/verify which swaps the result into #result.
func (s *server) handleVerifyPage(w http.ResponseWriter, r *http.Request) {
	s.render(w, "verify.html", map[string]any{
		"ProjectDir": s.projectDir,
	})
}
