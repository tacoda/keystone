package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// runShow handles `keystone show <kind> <id> [--json] [--dir <path>]`.
//
// Walks the charter once, composes the target primitive, and prints
// the descriptor plus the cross-references that are implicit in
// other primitives' frontmatter:
//
//   - The target's own descriptor (kind, id, description, path,
//     severity, model, tags, host_triggers).
//   - Includes: which concerns this primitive composes in.
//   - Included-by (when the target is a concern): which primitives
//     pull this concern in.
//   - Traces: the corpus entries this guide forward-links to (and the
//     reverse — guides that trace into this corpus entry).
//   - Host hooks: each host_triggers entry for sensors.
//   - Tags: orthogonal taxonomy.
//
// Pure read; never writes. Walks the full primitive set once,
// dereferences associations by id, prints.
func runShow(args []string) error {
	dir := "."
	asJSON := false
	var kindArg, idArg string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--help" || a == "-h":
			printShowUsage(os.Stdout)
			return nil
		case a == "--dir":
			if i+1 >= len(args) {
				return fmt.Errorf("flag %s requires a value", a)
			}
			dir = args[i+1]
			i++
		case strings.HasPrefix(a, "--dir="):
			dir = strings.TrimPrefix(a, "--dir=")
		case a == "--json":
			asJSON = true
		case strings.HasPrefix(a, "-"):
			return fmt.Errorf("unknown flag %s", a)
		default:
			if kindArg == "" {
				kindArg = a
			} else if idArg == "" {
				idArg = a
			} else {
				return fmt.Errorf("unexpected extra positional %q (usage: keystone show <kind> <id>)", a)
			}
		}
	}
	if kindArg == "" || idArg == "" {
		return fmt.Errorf("missing required positional arguments — usage: keystone show <kind> <id>")
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}

	primitives, _, err := primitive.Walk(absDir, config.DefaultCharterRoot)
	if err != nil {
		return err
	}
	composed, _ := primitive.Compose(primitives)

	target, found := findPrimitive(composed, kindArg, idArg)
	if !found {
		return fmt.Errorf("no primitive with kind=%s id=%s", kindArg, idArg)
	}

	view := buildShowView(target, composed)

	if asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)
		return enc.Encode(view)
	}
	printShowText(os.Stdout, view)
	return nil
}

// showView is the structured snapshot rendered by `keystone show`.
// Same shape feeds both text and JSON output — single source of truth
// for what `show` knows.
type showView struct {
	Kind         string                  `json:"kind"`
	ID           string                  `json:"id"`
	Description  string                  `json:"description"`
	Path         string                  `json:"path"`
	Provenance   string                  `json:"provenance,omitempty"`
	Severity     string                  `json:"severity,omitempty"`
	Model        string                  `json:"model,omitempty"`
	Phase        string                  `json:"phase,omitempty"`
	Tags         []string                `json:"tags,omitempty"`
	Tools        []string                `json:"tools,omitempty"`
	Globs        []string                `json:"globs,omitempty"`
	Triggers     []string                `json:"triggers,omitempty"`
	Includes     []string                `json:"includes,omitempty"`
	IncludedBy   []string                `json:"included_by,omitempty"`
	Traces       []string                `json:"traces,omitempty"`
	TracedBy     []string                `json:"traced_by,omitempty"`
	HostTriggers []primitive.HostTrigger `json:"host_triggers,omitempty"`
}

func buildShowView(target primitive.Primitive, all []primitive.Primitive) showView {
	v := showView{
		Kind:         target.Kind,
		ID:           target.ID,
		Description:  target.Description,
		Path:         target.Path,
		Provenance:   target.Provenance,
		Severity:     target.Severity,
		Model:        target.Model,
		Phase:        target.Phase,
		Tags:         target.Tags,
		Tools:        target.Tools,
		Globs:        target.Globs,
		Triggers:     target.Triggers,
		Includes:     target.Includes,
		Traces:       target.Corpus,
		HostTriggers: target.HostTriggers,
	}

	// Reverse-lookup: who includes this primitive (only meaningful
	// when target is a concern, but we run the lookup unconditionally
	// because nothing prevents a future kind from being included-able).
	for _, p := range all {
		for _, inc := range p.Includes {
			if inc == target.ID {
				v.IncludedBy = append(v.IncludedBy, p.Kind+"/"+p.ID)
			}
		}
	}
	sort.Strings(v.IncludedBy)

	// Reverse-lookup: who traces to this primitive (typically corpus
	// targets — find every guide whose `traces:` mentions this id).
	for _, p := range all {
		for _, tr := range p.Corpus {
			if tr == target.ID {
				v.TracedBy = append(v.TracedBy, p.Kind+"/"+p.ID)
			}
		}
	}
	sort.Strings(v.TracedBy)

	return v
}

// findPrimitive matches by (kind, id) exact-equal. Returns the
// composed primitive (post-includes merge) so the displayed surface
// reflects what the agent actually sees.
func findPrimitive(primitives []primitive.Primitive, kind, id string) (primitive.Primitive, bool) {
	for _, p := range primitives {
		if p.Kind == kind && p.ID == id {
			return p, true
		}
	}
	return primitive.Primitive{}, false
}

// printShowText renders the view as a human-readable card. Keeps
// sections compact and only prints headings that have content.
func printShowText(w *os.File, v showView) {
	fmt.Fprintf(w, "%s/%s\n", v.Kind, v.ID)
	if v.Description != "" {
		fmt.Fprintf(w, "  %s\n", v.Description)
	}
	fmt.Fprintln(w)

	if v.Path != "" {
		fmt.Fprintf(w, "  path:       %s\n", v.Path)
	}
	if v.Provenance != "" {
		fmt.Fprintf(w, "  provenance: %s\n", v.Provenance)
	}
	if v.Severity != "" {
		fmt.Fprintf(w, "  severity:   %s\n", v.Severity)
	}
	if v.Model != "" {
		fmt.Fprintf(w, "  model:      %s\n", v.Model)
	}
	if v.Phase != "" {
		fmt.Fprintf(w, "  phase:      %s\n", v.Phase)
	}
	if len(v.Tags) > 0 {
		fmt.Fprintf(w, "  tags:       %s\n", strings.Join(v.Tags, ", "))
	}

	if len(v.Tools) > 0 {
		fmt.Fprintln(w, "\n  tools:")
		for _, t := range v.Tools {
			fmt.Fprintf(w, "    - %s\n", t)
		}
	}
	if len(v.Globs) > 0 {
		fmt.Fprintln(w, "\n  globs:")
		for _, g := range v.Globs {
			fmt.Fprintf(w, "    - %s\n", g)
		}
	}
	if len(v.Triggers) > 0 {
		fmt.Fprintln(w, "\n  triggers:")
		for _, t := range v.Triggers {
			fmt.Fprintf(w, "    - %s\n", t)
		}
	}
	if len(v.HostTriggers) > 0 {
		fmt.Fprintln(w, "\n  host_triggers:")
		for _, h := range v.HostTriggers {
			matcher := h.Matcher
			if matcher == "" {
				matcher = "(any)"
			}
			fmt.Fprintf(w, "    - %s / %s → %s\n", h.Phase, matcher, h.Command)
		}
	}
	if len(v.Includes) > 0 {
		fmt.Fprintln(w, "\n  includes (composed in):")
		for _, id := range v.Includes {
			fmt.Fprintf(w, "    - %s\n", id)
		}
	}
	if len(v.IncludedBy) > 0 {
		fmt.Fprintln(w, "\n  included by:")
		for _, ref := range v.IncludedBy {
			fmt.Fprintf(w, "    - %s\n", ref)
		}
	}
	if len(v.Traces) > 0 {
		fmt.Fprintln(w, "\n  traces (forward):")
		for _, t := range v.Traces {
			fmt.Fprintf(w, "    - %s\n", t)
		}
	}
	if len(v.TracedBy) > 0 {
		fmt.Fprintln(w, "\n  traced by:")
		for _, ref := range v.TracedBy {
			fmt.Fprintf(w, "    - %s\n", ref)
		}
	}
}

func printShowUsage(w *os.File) {
	fmt.Fprint(w, `keystone show — show one primitive's descriptor + associations

Usage:
  keystone show <kind> <id> [--json] [--dir <path>]

Walks the charter once, composes the target primitive, and prints:

  - The target's own descriptor (kind, id, description, path,
    severity, model, tags, host_triggers).
  - includes: concerns the target composes in.
  - included by: primitives that include this one (when target is
    a concern).
  - traces: forward links from a guide to corpus entries.
  - traced by: reverse — guides whose traces:[ ] mention this id.
  - Host hooks (when target is a sensor with host_triggers).

Flags:
  --json            Emit a structured JSON object instead of text.
  --dir <path>      Project root (defaults to cwd).

Examples:
  keystone show sensor secret-scan
  keystone show concern reads-diff
  keystone show guide guides/idioms/go/stdlib-first --json
`)
}
