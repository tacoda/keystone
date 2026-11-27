package budget

import (
	"strings"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// AlwaysLoaded is the keystone.json `budgets` key that caps each host's
// always-loaded tier.
const AlwaysLoaded = "always_loaded"

// DefaultAlwaysLoadedCap is the always-loaded cap applied when
// keystone.json has no `budgets.always_loaded` entry (6.0 pillar 1:
// nothing loads eagerly but the block and the index, and verify enforces
// the cap). Units are Estimate words. An explicit `always_loaded` entry
// overrides it; `max_tokens: 0` on that entry disables the cap.
const DefaultAlwaysLoadedCap = 8000

// withDefaultCap returns the normalized caps with the default
// always_loaded cap filled in when the project did not configure one.
func withDefaultCap(cs map[string]config.BudgetSpec) map[string]config.BudgetSpec {
	if _, ok := cs[AlwaysLoaded]; !ok {
		cs[AlwaysLoaded] = config.BudgetSpec{MaxTokens: DefaultAlwaysLoadedCap}
	}
	return cs
}

// legacyPorts maps the 4.x `budgets` keys (charter directories) to kinds.
// `adapters` has no 5.0 counterpart and is ignored.
var legacyPorts = map[string]string{
	"guides": "rule", "corpus": "reference", "sensors": "hook",
	"actions": "command", "playbooks": "skill",
}

// caps normalizes keystone.json `budgets`: always_loaded, a 5.0 kind, a
// 4.x kind name, or a 4.x port. Two keys for one kind keep the stricter.
func caps(budgets map[string]config.BudgetSpec) map[string]config.BudgetSpec {
	out := map[string]config.BudgetSpec{}
	for key, spec := range budgets {
		k := capKey(key)
		if k == "" {
			continue
		}
		out[k] = stricter(out[k], spec)
	}
	return out
}

func capKey(key string) string {
	if key == AlwaysLoaded {
		return key
	}
	if kind, ok := legacyPorts[key]; ok {
		return kind
	}
	if k, _ := primitive.NormalizeKind(strings.TrimSuffix(key, "s")); primitive.IsKnownKind(string(k)) {
		return string(k)
	}
	return ""
}

func stricter(a, b config.BudgetSpec) config.BudgetSpec {
	return config.BudgetSpec{MaxTokens: minCap(a.MaxTokens, b.MaxTokens), MaxTokensPerLoad: minCap(a.MaxTokensPerLoad, b.MaxTokensPerLoad)}
}

// minCap is the smaller non-zero cap (0 means none).
func minCap(a, b int) int {
	if a == 0 {
		return b
	}
	if b == 0 {
		return a
	}
	return min(a, b)
}

// checkCaps lists every cap the report exceeds: each host's always-loaded
// tier, each kind's on-demand total, and each file against its kind's
// per-load cap. The default always_loaded cap applies when the project
// configured none.
func checkCaps(r Report, cs map[string]config.BudgetSpec) []Violation {
	cs = withDefaultCap(cs)
	vs := alwaysViolations(r.Hosts, cs[AlwaysLoaded].MaxTokens)
	vs = append(vs, kindViolations(r.ByKind, cs)...)
	return append(vs, perLoadViolations(r.OnDemand, cs)...)
}

func alwaysViolations(hosts []HostLoad, limit int) []Violation {
	var vs []Violation
	for _, h := range hosts {
		if limit > 0 && h.Always > limit {
			// Name the files that make up the over-cap tier.
			var files []string
			for _, e := range h.Entries {
				files = append(files, e.Path)
			}
			vs = append(vs, Violation{
				Scope:  AlwaysLoaded + " (" + h.Host + ")",
				Path:   strings.Join(files, ", "),
				Tokens: h.Always, Cap: limit,
			})
		}
	}
	return vs
}

func kindViolations(byKind map[string]int, cs map[string]config.BudgetSpec) []Violation {
	var vs []Violation
	for _, k := range primitive.KnownKinds {
		kind := string(k)
		if limit := cs[kind].MaxTokens; limit > 0 && byKind[kind] > limit {
			vs = append(vs, Violation{Scope: kind, Tokens: byKind[kind], Cap: limit})
		}
	}
	return vs
}

func perLoadViolations(files []Entry, cs map[string]config.BudgetSpec) []Violation {
	var vs []Violation
	for _, e := range files {
		if limit := cs[e.Kind].MaxTokensPerLoad; limit > 0 && e.Tokens > limit {
			vs = append(vs, Violation{Scope: e.Kind + " per load", Path: e.Path, Tokens: e.Tokens, Cap: limit})
		}
	}
	return vs
}
