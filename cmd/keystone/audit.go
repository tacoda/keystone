package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

func auditCmd() *cobra.Command {
	c := &cobra.Command{
		Use:                "audit [--json]",
		Short:              "Dual-flywheel report: learning inbox + pruning candidates (uncited, unscoped, untriggered primitives)",
		DisableFlagParsing: true,
		RunE:               runAndForward(runAudit),
	}
	return c
}

type auditCandidate struct {
	Path   string `json:"path"`
	Layer  string `json:"layer,omitempty"`
	Source string `json:"source,omitempty"`
	Status string `json:"status,omitempty"`
}

type pruneCandidate struct {
	Kind    string   `json:"kind"`
	ID      string   `json:"id"`
	Path    string   `json:"path"`
	Reasons []string `json:"reasons"`
}

type auditReport struct {
	Learning struct {
		Pending  []auditCandidate `json:"pending"`
		Accepted []auditCandidate `json:"accepted"`
	} `json:"learning"`
	Pruning struct {
		Total      int              `json:"total_primitives"`
		Candidates []pruneCandidate `json:"candidates"`
	} `json:"pruning"`
}

func printAuditUsage(w *os.File) {
	fmt.Fprint(w, `keystone audit [--json] [--dir D]

Dual-flywheel report: the learning inbox (pending/accepted) and pruning
candidates (project primitives that are uncited, unscoped, and untriggered).
Data only — acting on a candidate is charter move/demote or an edit.
`)
}

// runAudit handles `keystone audit [--json] [--dir D]`. The report is
// data, not verdicts (decision 16): the Learning flywheel side lists the
// inbox, the Pruning side lists primitives no citation, path scope, or
// event reaches. Acting on a candidate is `charter move` / `demote` /
// an edit — never automatic.
func runAudit(args []string) error {
	dir, asJSON := ".", false
	pos, help, err := parseFlagsArgs(args,
		map[string]*string{"--dir": &dir},
		map[string]*bool{"--json": &asJSON},
		func() { printAuditUsage(os.Stdout) })
	if help || err != nil {
		return err
	}
	if len(pos) > 0 {
		return fmt.Errorf("audit takes no positional args; got %q", pos[0])
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	rep, err := auditProject(absDir)
	if err != nil {
		return err
	}
	if asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(rep)
	}
	fmt.Fprintf(os.Stdout, "Learning flywheel — %d pending, %d accepted\n", len(rep.Learning.Pending), len(rep.Learning.Accepted))
	for _, c := range rep.Learning.Pending {
		fmt.Fprintf(os.Stdout, "  %s%s%s\n", c.Path, layerNote(c.Layer), sourceNote(c.Source))
	}
	for _, c := range rep.Learning.Accepted {
		fmt.Fprintf(os.Stdout, "  %s (accepted)%s%s\n", c.Path, layerNote(c.Layer), sourceNote(c.Source))
	}
	fmt.Fprintf(os.Stdout, "Pruning flywheel — %d candidate(s) of %d primitives\n", len(rep.Pruning.Candidates), rep.Pruning.Total)
	for _, c := range rep.Pruning.Candidates {
		fmt.Fprintf(os.Stdout, "  %s/%s  %s — %s\n", c.Kind, c.ID, c.Path, strings.Join(c.Reasons, ", "))
	}
	if len(rep.Learning.Pending) == 0 && len(rep.Pruning.Candidates) == 0 {
		fmt.Fprintln(os.Stdout, "  (nothing to act on)")
	}
	return nil
}

func layerNote(layer string) string {
	if layer == "" {
		return ""
	}
	return "  → " + layer
}

func sourceNote(source string) string {
	if source == "" {
		return ""
	}
	return "  (" + source + ")"
}

// auditProject computes the dual-flywheel report for the charter at absDir.
func auditProject(absDir string) (auditReport, error) {
	var rep auditReport
	prims, _, err := primitive.Walk(absDir, config.DefaultAgentsRoot)
	if err != nil {
		return rep, err
	}
	rep.Pruning.Total = len(prims)

	// Inbound citations: every corpus: entry any primitive carries,
	// normalized to a bare id (kind prefix stripped), is the set of
	// things something points at.
	cited := map[string]bool{}
	for _, p := range prims {
		for _, ref := range p.Corpus {
			cited[normalizeCitation(ref)] = true
		}
	}
	for _, p := range prims {
		if p.Provenance != "project" {
			continue // adopted repo docs and vendored policy files are not pruning candidates
		}
		switch p.Kind {
		case string(primitive.KindRule), string(primitive.KindReference), string(primitive.KindCorpus):
		default:
			continue // hooks fire on events, skills on description match — different liveness signal
		}
		var reasons []string
		if !cited[p.ID] {
			reasons = append(reasons, "uncited")
		}
		if len(p.Paths) == 0 {
			reasons = append(reasons, "unscoped (no paths:)")
		}
		if p.Event == "" && len(p.Triggers) == 0 {
			reasons = append(reasons, "untriggered")
		}
		if len(reasons) == 3 {
			rep.Pruning.Candidates = append(rep.Pruning.Candidates, pruneCandidate{
				Kind: p.Kind, ID: p.ID, Path: p.Path, Reasons: reasons,
			})
		}
	}
	sort.Slice(rep.Pruning.Candidates, func(i, j int) bool {
		return rep.Pruning.Candidates[i].ID < rep.Pruning.Candidates[j].ID
	})

	inbox := filepath.Join(absDir, learningInbox)
	entries, err := os.ReadDir(inbox)
	if err != nil && !os.IsNotExist(err) {
		return rep, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	for _, name := range names {
		rel := filepath.Join(learningInbox, name)
		c, ok := readAuditCandidate(filepath.Join(absDir, filepath.FromSlash(rel)))
		if !ok {
			continue
		}
		c.Path = rel
		if c.Status == "accepted" {
			rep.Learning.Accepted = append(rep.Learning.Accepted, c)
		} else {
			rep.Learning.Pending = append(rep.Learning.Pending, c)
		}
	}
	return rep, nil
}

// readAuditCandidate plucks the curation fields from an inbox file.
// Frontmatter only — the body is never read.
func readAuditCandidate(abs string) (auditCandidate, bool) {
	var c auditCandidate
	data, err := os.ReadFile(abs)
	if err != nil {
		return c, false
	}
	fm, _, ok := primitive.SplitFrontmatter(string(data))
	if !ok {
		return c, false
	}
	var parsed candidateFM
	if err := yaml.Unmarshal([]byte(fm), &parsed); err != nil {
		return c, false
	}
	c.Layer, c.Source, c.Status = parsed.ProposedLayer, parsed.Source, parsed.Status
	return c, true
}

// normalizeCitation strips an optional kind prefix (corpus/, rule/…)
// from a corpus: entry, leaving the bare id.
func normalizeCitation(ref string) string {
	bare := strings.TrimPrefix(ref, "corpus/")
	if j := strings.IndexByte(bare, '/'); j > 0 {
		bare = bare[j+1:]
	}
	return bare
}
