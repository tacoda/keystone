package web

import (
	"net/http"
)

// handleFlywheels renders the flywheel-trigger page. The actions
// themselves (learn / synthesize / audit) are agent-driven — they
// require an agent in session to do the reasoning. The dashboard
// surfaces the canonical invocation phrasing the user types into
// their host of choice.
func (s *server) handleFlywheels(w http.ResponseWriter, r *http.Request) {
	s.renderPage(w, r, "flywheels.html", map[string]any{
		"ProjectDir": s.projectDir,
	})
}
