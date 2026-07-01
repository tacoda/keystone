package budget

import (
	"testing"

	"github.com/tacoda/keystone/internal/framework/config"
)

func TestEstimate(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int
	}{
		{"empty", "", 0},
		{"whitespace only", "   \n\t  \n", 0},
		{"one word", "hello", 1},
		{"three words", "alpha beta gamma", 3},
		{"newlines + tabs", "alpha\nbeta\tgamma\nrho", 4},
		{"markdown link", "see [reasoning](corpus/X.md) for context", 4}, // markdown link is one whitespace-token
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Estimate([]byte(tt.in)); got != tt.want {
				t.Errorf("Estimate(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestPortForPath(t *testing.T) {
	tests := []struct {
		rel      string
		charter  string
		wantPort string
	}{
		{"charter/guides/process/spec.md", "charter", "guides"},
		{"charter/corpus/principles/tdd.md", "charter", "corpus"},
		{"charter/sensors/build.md", "charter", "sensors"},
		{"charter/actions/spec.md", "charter", "actions"},
		{"charter/playbooks/task.md", "charter", "playbooks"},
		{"charter/adapters/claude-code/lifecycle.md", "charter", "adapters"},
		{"charter/learning/inbox/X.md", "charter", ""},
		{"charter/archive/X.md", "charter", ""},
		{"charter/README.md", "charter", ""},
		{"charter/policies/X/guides/Y.md", "charter", ""}, // policies/ has its own logic; not a port
		{"custom/guides/X.md", "custom", "guides"},
		{"other/guides/X.md", "charter", ""}, // wrong charter root
	}
	for _, tt := range tests {
		t.Run(tt.rel, func(t *testing.T) {
			if got := PortForPath(tt.rel, tt.charter); got != tt.wantPort {
				t.Errorf("PortForPath(%q, %q) = %q, want %q", tt.rel, tt.charter, got, tt.wantPort)
			}
		})
	}
}

func TestAllocator_Report(t *testing.T) {
	a := NewAllocator()
	a.Add("guides", "charter/guides/a.md", 100)
	a.Add("guides", "charter/guides/b.md", 50)
	a.Add("guides", "charter/guides/c.md", 200)
	a.Add("corpus", "charter/corpus/x.md", 1000)

	cfg := &config.ProjectConfig{
		Budgets: map[string]config.BudgetSpec{
			"guides": {MaxTokens: 300},
			"corpus": {MaxTokens: 800},
		},
	}

	rep := a.Report(cfg, 2)
	if len(rep) != 2 {
		t.Fatalf("expected 2 port reports, got %d", len(rep))
	}

	// Sorted by port name.
	if rep[0].Port != "corpus" || rep[1].Port != "guides" {
		t.Errorf("port order: got %s,%s", rep[0].Port, rep[1].Port)
	}

	// corpus: 1000 tokens, max 800, over by 200
	c := rep[0]
	if c.Tokens != 1000 || c.MaxTokens != 800 || c.OverBy != 200 {
		t.Errorf("corpus report: %+v", c)
	}
	if !c.IsOverBudget() {
		t.Errorf("corpus should be over budget")
	}

	// guides: 350 tokens total (100+50+200), max 300, over by 50
	g := rep[1]
	if g.Tokens != 350 || g.MaxTokens != 300 || g.OverBy != 50 {
		t.Errorf("guides report: %+v", g)
	}
	if !g.IsOverBudget() {
		t.Errorf("guides should be over budget")
	}

	// TopFiles truncated to 2, sorted desc.
	if len(g.TopFiles) != 2 {
		t.Errorf("expected 2 top files, got %d", len(g.TopFiles))
	}
	if g.TopFiles[0].Path != "charter/guides/c.md" || g.TopFiles[0].Tokens != 200 {
		t.Errorf("top file [0]: %+v", g.TopFiles[0])
	}
}

func TestAllocator_NoBudget(t *testing.T) {
	a := NewAllocator()
	a.Add("sensors", "charter/sensors/x.md", 500)
	rep := a.Report(nil, 0)
	if len(rep) != 1 {
		t.Fatalf("got %d reports", len(rep))
	}
	if rep[0].MaxTokens != 0 {
		t.Errorf("MaxTokens should be 0 when budgets unset, got %d", rep[0].MaxTokens)
	}
	if rep[0].IsOverBudget() {
		t.Errorf("should not be over budget when no cap set")
	}
}
