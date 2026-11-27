package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// learningPromoted is where accepted inbox candidates go after
// synthesize promotes them — kept as the record, never deleted.
const learningPromoted = ".keystone/learning/promoted"

func synthesizeCmd() *cobra.Command {
	c := &cobra.Command{
		Use:                "synthesize [--dry-run] [--json]",
		Short:              "Promote accepted learning-inbox candidates to their proposed layer, then re-index (the Learning flywheel's promotion step)",
		DisableFlagParsing: true,
		RunE:               runAndForward(runSynthesize),
	}
	return c
}

// candidateFM is the frontmatter shape `keystone learn` writes, plus the
// curation fields a human or agent sets before synthesize may act:
// `status: accepted`, `proposed-layer`, optional `proposed-id` and
// `proposed-globs` (which become the promoted rule's `paths:`).
type candidateFM struct {
	Status        string   `yaml:"status"`
	Source        string   `yaml:"source"`
	Description   string   `yaml:"description"`
	ProposedLayer string   `yaml:"proposed-layer"`
	ProposedID    string   `yaml:"proposed-id"`
	ProposedGlobs []string `yaml:"proposed-globs"`
}

type synthMove struct {
	Candidate string `json:"candidate"`
	Kind      string `json:"kind"`
	ID        string `json:"id"`
	Path      string `json:"path"`

	// body and curation fields are carried from the candidate but never
	// serialized into the report.
	body   string   `json:"-"`
	globs  []string `json:"-"`
	desc   string   `json:"-"`
	source string   `json:"-"`
}

type synthSkip struct {
	Candidate string `json:"candidate"`
	Reason    string `json:"reason"`
}

type synthReport struct {
	Promoted []synthMove `json:"promoted"`
	Skipped  []synthSkip `json:"skipped"`
}

var inboxName = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}-\d{2}-\d{2}-(.*)\.md$`)

// runSynthesize handles `keystone synthesize [--dry-run] [--json] [--dir D]`.
// The deterministic half of the promotion step: every inbox candidate
// marked `status: accepted` with a mappable `proposed-layer` is written
// into that layer as a primitive and moved to learning/promoted/. Layer
// choice and acceptance stay with the human/agent curating the
// candidate — synthesize only executes what was accepted.
func runSynthesize(args []string) error {
	dryRun, asJSON := false, false
	for i := 0; i < len(args); i++ {
		switch a := args[i]; {
		case a == "--dry-run":
			dryRun = true
		case a == "--json":
			asJSON = true
		case a == "--dir":
			i++
			if i >= len(args) {
				return fmt.Errorf("--dir requires a value")
			}
		case strings.HasPrefix(a, "--dir="):
		case strings.HasPrefix(a, "-"):
			return fmt.Errorf("unknown flag %q", a)
		default:
			return fmt.Errorf("synthesize takes no positional args; got %q", a)
		}
	}
	absDir, err := filepath.Abs(dirArg(args))
	if err != nil {
		return err
	}
	rep, err := synthesizeProject(absDir, dryRun)
	if err != nil {
		return err
	}
	if asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(rep)
	}
	for _, s := range rep.Skipped {
		fmt.Fprintf(os.Stdout, "  … %s — %s\n", s.Candidate, s.Reason)
	}
	for _, m := range rep.Promoted {
		verb := "promoted"
		if dryRun {
			verb = "would promote"
		}
		fmt.Fprintf(os.Stdout, "✓ %s %s → %s/%s (%s)\n", verb, m.Candidate, m.Kind, m.ID, m.Path)
	}
	if len(rep.Promoted) == 0 && len(rep.Skipped) == 0 {
		fmt.Fprintln(os.Stdout, "(learning inbox empty)")
	}
	return nil
}

// synthesizeProject walks the learning inbox and promotes every accepted
// candidate. With dryRun it reports what it would do and writes nothing.
func synthesizeProject(absDir string, dryRun bool) (synthReport, error) {
	var rep synthReport
	inbox := filepath.Join(absDir, learningInbox)
	entries, err := os.ReadDir(inbox)
	if os.IsNotExist(err) {
		return rep, nil
	}
	if err != nil {
		return rep, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	promoted := 0
	for _, name := range names {
		rel := filepath.Join(learningInbox, name)
		move, skip, err := planPromotion(absDir, rel, name)
		if err != nil {
			return rep, err
		}
		if skip != nil {
			rep.Skipped = append(rep.Skipped, *skip)
			continue
		}
		if !dryRun {
			if err := applyPromotion(absDir, rel, *move); err != nil {
				return rep, err
			}
			promoted++
		}
		rep.Promoted = append(rep.Promoted, *move)
	}
	if promoted > 0 {
		if err := reindex(absDir); err != nil {
			return rep, err
		}
	}
	return rep, nil
}

// planPromotion maps one inbox candidate to its destination primitive,
// or a skip with the reason it was left alone.
func planPromotion(absDir, rel, name string) (*synthMove, *synthSkip, error) {
	skip := func(reason string) (*synthMove, *synthSkip, error) {
		return nil, &synthSkip{Candidate: rel, Reason: reason}, nil
	}
	m := inboxName.FindStringSubmatch(name)
	if m == nil {
		return skip("filename is not <captured>-<slug>.md")
	}
	slug := m[1]
	data, err := os.ReadFile(filepath.Join(absDir, filepath.FromSlash(rel)))
	if err != nil {
		return nil, nil, err
	}
	fm, body, ok := primitive.SplitFrontmatter(string(data))
	if !ok {
		return skip("no frontmatter")
	}
	var c candidateFM
	if err := yaml.Unmarshal([]byte(fm), &c); err != nil {
		return nil, nil, fmt.Errorf("%s: parse frontmatter: %w", rel, err)
	}
	if c.Status != "accepted" {
		return skip("status is not accepted — curate the candidate first")
	}
	if inv := invalidLayer(c.ProposedLayer); inv != "" {
		return skip("proposed-layer " + inv)
	}
	kind, id := "", c.ProposedID
	switch {
	case c.ProposedLayer == "hook":
		kind = "hook"
		if id == "" {
			id = slug
		}
	case strings.HasPrefix(c.ProposedLayer, "rules/"):
		kind = "rule"
		if id == "" {
			id = strings.TrimPrefix(c.ProposedLayer, "rules/")
		}
		if id == "" {
			id = slug
		}
	case c.ProposedLayer == "references/principles":
		kind = "reference"
		if id == "" {
			id = "principles/" + slug
		}
	default:
		return skip("no proposed-layer — the layer choice is the curator's judgment")
	}
	dest := filepath.Join(config.DefaultAgentsRoot, kindDirName(kind), filepath.FromSlash(id+".md"))
	if _, err := os.Stat(filepath.Join(absDir, dest)); err == nil {
		return skip("destination exists: " + dest + " (no silent overwrites — merge by hand)")
	}
	return &synthMove{
		Candidate: rel, Kind: kind, ID: id, Path: filepath.ToSlash(dest),
		body: body, globs: c.ProposedGlobs, desc: c.Description, source: c.Source,
	}, nil, nil
}

// applyPromotion writes the destination primitive from the candidate's
// body, then files the candidate under learning/promoted/ as the record.
func applyPromotion(absDir, rel string, m synthMove) error {
	dest := filepath.Join(absDir, filepath.FromSlash(m.Path))
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	fm := map[string]any{"kind": m.Kind, "id": m.ID}
	desc := m.desc
	if desc == "" {
		desc = "Promoted learning candidate " + filepath.Base(rel)
		if m.source != "" {
			desc += " (source: " + m.source + ")"
		}
	}
	fm["description"] = desc
	if len(m.globs) > 0 {
		fm["paths"] = m.globs
	}
	fmBytes, err := yaml.Marshal(fm)
	if err != nil {
		return err
	}
	content := "---\n" + string(fmBytes) + "---\n" + m.body
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	if err := os.WriteFile(dest, []byte(content), 0o644); err != nil {
		return err
	}
	promoted := filepath.Join(absDir, learningPromoted, filepath.Base(rel))
	if err := os.MkdirAll(filepath.Dir(promoted), 0o755); err != nil {
		return err
	}
	return os.Rename(filepath.Join(absDir, filepath.FromSlash(rel)), promoted)
}
