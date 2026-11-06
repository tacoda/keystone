package budget

import (
	"errors"
	"os"
	"path/filepath"
	"sort"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
	"github.com/tacoda/keystone/internal/framework/projection"
)

// Entry is one file's (or one listing's) token cost.
type Entry struct {
	Path   string `json:"path"`
	Kind   string `json:"kind"`
	Tokens int    `json:"tokens"`
}

// HostLoad is what one host loads into every session: its root files,
// the rules it applies everywhere, and the name + description of every
// skill, command, and subagent it lists up front.
type HostLoad struct {
	Host    string  `json:"host"`
	Always  int     `json:"always_loaded"`
	Entries []Entry `json:"entries"` // largest first
}

// Report is the context budget: the always-loaded tier per host, and the
// on-demand tier (bodies an agent opens when they apply), which is the
// same for every host because every host can reach .agents/.
type Report struct {
	Hosts      []HostLoad     `json:"hosts"`
	OnDemand   []Entry        `json:"on_demand"` // largest first
	ByKind     map[string]int `json:"by_kind"`   // on-demand tokens per kind
	Violations []Violation    `json:"violations,omitempty"`
}

// Violation is one cap exceeded.
type Violation struct {
	Scope  string `json:"scope"` // "always_loaded (<host>)", "<kind>", or "<kind> per load"
	Path   string `json:"path,omitempty"`
	Tokens int    `json:"tokens"`
	Cap    int    `json:"cap"`
}

// OverBy is how far past its cap the violation is.
func (v Violation) OverBy() int { return v.Tokens - v.Cap }

// onDemandKinds are the kinds whose bodies an agent opens on demand.
var onDemandKinds = map[string]bool{
	"rule": true, "reference": true, "skill": true, "command": true,
	"subagent": true, "tool": true, "document": true,
}

// listedKinds are the kinds a host lists up front, by name and description.
var listedKinds = map[string]bool{"skill": true, "command": true, "subagent": true}

// MeasureProject walks the charter at projectDir and measures it for the
// hosts and caps in keystone.json (claude-code only, no caps, without one).
func MeasureProject(projectDir string) (Report, error) {
	prims, _, err := primitive.Walk(projectDir, config.DefaultAgentsRoot)
	if err != nil {
		return Report{}, err
	}
	composed, _ := primitive.Compose(prims)
	cfg, err := config.ReadProjectConfig(projectDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return Report{}, err
	}
	var adapters []string
	var budgets map[string]config.BudgetSpec
	if cfg != nil {
		adapters, budgets = cfg.Adapters, cfg.Budgets
	}
	hosts, err := projection.HostsFor(projectDir, adapters)
	if err != nil {
		return Report{}, err
	}
	return Measure(projectDir, composed, hosts, budgets), nil
}

// AlwaysCap is the always_loaded cap in keystone.json, 0 when none.
func AlwaysCap(projectDir string) int {
	cfg, err := config.ReadProjectConfig(projectDir)
	if err != nil || cfg == nil {
		return 0
	}
	return caps(cfg.Budgets)[AlwaysLoaded].MaxTokens
}

// Measure computes the budget from each host's projection profile, so a
// host's always-loaded tier reflects what that host reads, and applies the
// caps in budgets (keystone.json).
func Measure(projectDir string, prims []primitive.Primitive, hosts []projection.Profile, budgets map[string]config.BudgetSpec) Report {
	r := Report{ByKind: map[string]int{}}
	for _, h := range hosts {
		r.Hosts = append(r.Hosts, hostLoad(projectDir, h, prims))
	}
	for _, p := range prims {
		if onDemandKinds[p.Kind] {
			e := Entry{Path: p.Path, Kind: p.Kind, Tokens: fileTokens(projectDir, p.Path)}
			r.OnDemand = append(r.OnDemand, e)
			r.ByKind[p.Kind] += e.Tokens
		}
	}
	sortEntries(r.OnDemand)
	r.Violations = checkCaps(r, caps(budgets))
	return r
}

// hostLoad sums one host's always-loaded tier.
func hostLoad(projectDir string, h projection.Profile, prims []primitive.Primitive) HostLoad {
	l := HostLoad{Host: h.Host}
	for _, e := range rootEntries(projectDir, h.Host) {
		l.add(e)
	}
	for _, p := range prims {
		if e, ok := alwaysEntry(projectDir, h, p); ok {
			l.add(e)
		}
	}
	sortEntries(l.Entries)
	return l
}

// alwaysEntry: a skill, command, or subagent the host loads itself costs
// its listing (name and description); a rule with no `paths:` the host
// loads itself costs its whole body. Anything else loads on demand.
func alwaysEntry(projectDir string, h projection.Profile, p primitive.Primitive) (Entry, bool) {
	if !hostLoads(h, p) {
		return Entry{}, false
	}
	switch {
	case listedKinds[p.Kind]:
		return Entry{Path: p.Path, Kind: p.Kind + " listing", Tokens: Estimate([]byte(p.ID + " " + p.Description))}, true
	case p.Kind == "rule" && len(p.Paths) == 0:
		return Entry{Path: p.Path, Kind: "rule", Tokens: fileTokens(projectDir, p.Path)}, true
	}
	return Entry{}, false
}

// hostLoads: the host reads p by itself — natively, or through a
// projection it writes. An index or report entry means the agent opens p
// from .agents/ when it applies.
func hostLoads(h projection.Profile, p primitive.Primitive) bool {
	r, ok := h.Rules[primitive.Kind(p.Kind)]
	switch {
	case !ok || r.Strategy == projection.Index || r.Strategy == projection.Report:
		return false
	case r.Strategy == projection.Native || r.Dest == nil:
		return true
	}
	return r.Dest(p) != ""
}

// rootFiles are the files a host loads at session start. Every host reads
// AGENTS.md (Claude Code through CLAUDE.md's import); Aider also reads the
// lite index through .aider.conf.yml.
func rootFiles(host string) []string {
	switch host {
	case "claude-code":
		return []string{"CLAUDE.md", "AGENTS.md"}
	case "aider":
		return []string{"AGENTS.md", filepath.ToSlash(filepath.Join(config.StateRoot, config.IndexLiteName))}
	}
	return []string{"AGENTS.md"}
}

// rootEntries is the host's root files that exist.
func rootEntries(projectDir, host string) []Entry {
	var out []Entry
	for _, path := range rootFiles(host) {
		if t := fileTokens(projectDir, path); t > 0 {
			out = append(out, Entry{Path: path, Kind: "root", Tokens: t})
		}
	}
	return out
}

func (l *HostLoad) add(e Entry) {
	l.Entries = append(l.Entries, e)
	l.Always += e.Tokens
}

func fileTokens(projectDir, rel string) int {
	data, err := os.ReadFile(filepath.Join(projectDir, rel))
	if err != nil {
		return 0
	}
	return Estimate(data)
}

func sortEntries(es []Entry) {
	sort.SliceStable(es, func(i, j int) bool {
		if es[i].Tokens != es[j].Tokens {
			return es[i].Tokens > es[j].Tokens
		}
		return es[i].Path < es[j].Path
	})
}
