package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// promoteLadder orders information kinds by standing: corpus (the what)
// → reference (the why) → rule (the how). `keystone charter promote`
// walks up; `demote` walks down. Capability kinds (skill, command,
// agent, hook, document, eval, permission) are lateral — they move, but
// there is no ladder between them.
var promoteLadder = []primitive.Kind{
	primitive.KindCorpus, primitive.KindReference, primitive.KindRule,
}

// runCharterMove handles `keystone charter move <kind>/<id>
// <new-kind>[/<new-path>]`. It relocates the file into the target kind's
// folder (optionally renaming inside it), drops the now-redundant
// `kind:` (the destination folder is the declaration), updates `id:`
// when it changed, rewrites every `corpus:` citation across the charter
// that points at the moved id, and re-indexes. Bodies are untouched;
// vendored policy files are never moved or rewritten.
func runCharterMove(args []string) error {
	return runMove(args, "moved", "")
}

// runCharterPromote handles `keystone charter promote <kind>/<id>` —
// one step up the corpus → reference → rule ladder, keeping the topic
// path: corpus/billing/props promotes to reference/billing/props.
func runCharterPromote(args []string) error {
	return runLadder(args, +1)
}

// runCharterDemote handles `keystone charter demote <kind>/<id>` — one
// step down the ladder: a rule stops gating and becomes the why; a
// reference stops being citable and becomes plain fact.
func runCharterDemote(args []string) error {
	return runLadder(args, -1)
}

func runLadder(args []string, dir int) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: keystone charter %s <kind>/<id>", ladderVerb(dir))
	}
	ref := args[0]
	haveKind, _, ok := splitKindID(ref)
	if !ok {
		return fmt.Errorf("expected <kind>/<id>, got %q", ref)
	}
	next, ok := ladderStep(haveKind, dir)
	if !ok {
		return fmt.Errorf("%s %s is not a ladder kind (ladder: corpus → reference → rule)", ladderVerb(dir), haveKind)
	}
	return runMove([]string{ref, string(next)}, ladderVerb(dir)+"d", "")
}

func ladderVerb(dir int) string {
	if dir > 0 {
		return "promote"
	}
	return "demote"
}

// ladderStep returns the adjacent ladder kind in direction dir, or
// false when k is off the ladder or already at its end.
func ladderStep(k string, dir int) (primitive.Kind, bool) {
	cur := -1
	for i, kk := range promoteLadder {
		if string(kk) == k {
			cur = i
			break
		}
	}
	if cur < 0 {
		return "", false
	}
	next := cur + dir
	if next < 0 || next >= len(promoteLadder) {
		return "", false
	}
	return promoteLadder[next], true
}

func runMove(args []string, verb, _ string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: keystone charter move <kind>/<id> <new-kind>[/<new-path>]")
	}
	ref, target := args[0], args[1]
	haveKind, haveID, ok := splitKindID(ref)
	if !ok {
		return fmt.Errorf("expected <kind>/<id>, got %q", ref)
	}
	wantKind, wantPath, err := parseMoveTarget(target)
	if err != nil {
		return err
	}
	if !primitive.IsKnownKind(wantKind) {
		return fmt.Errorf("unknown kind %q (allowed: %s)", wantKind, primitive.KnownKindList())
	}

	absDir, err := filepath.Abs(".")
	if err != nil {
		return err
	}
	prims, _, err := primitive.Walk(absDir, config.DefaultAgentsRoot)
	if err != nil {
		return err
	}
	var moved *primitive.Primitive
	for i := range prims {
		if prims[i].Kind == haveKind && prims[i].ID == haveID {
			moved = &prims[i]
			break
		}
	}
	if moved == nil {
		return fmt.Errorf("no primitive %s/%s", haveKind, haveID)
	}
	if moved.Provenance != "project" {
		return fmt.Errorf("%s/%s is %s — keystone only moves project-owned primitives", haveKind, haveID, moved.Provenance)
	}

	newID := haveID
	if wantPath != "" {
		newID = strings.TrimSuffix(strings.TrimPrefix(wantPath, "/"), "/")
	}

	rel := filepath.Join(config.DefaultAgentsRoot, kindDirName(wantKind),
		strings.TrimPrefix(filepath.FromSlash(newID+".md"), string(filepath.Separator)))
	if _, err := os.Stat(filepath.Join(absDir, rel)); err == nil {
		return fmt.Errorf("destination exists: %s", rel)
	}
	absSrc := filepath.Join(absDir, filepath.FromSlash(moved.Path))
	if err := os.MkdirAll(filepath.Dir(filepath.Join(absDir, rel)), 0o755); err != nil {
		return err
	}

	data, err := os.ReadFile(absSrc)
	if err != nil {
		return err
	}
	out := string(data)
	if fm, body, hasFM := primitive.SplitFrontmatter(out); hasFM {
		fm2, ferr := primitive.MoveFrontmatter(fm, wantKind, newID)
		if ferr != nil {
			return fmt.Errorf("rewrite frontmatter: %w", ferr)
		}
		out = "---\n" + fm2 + "---\n" + body
	}
	if err := os.WriteFile(filepath.Join(absDir, rel), []byte(out), 0o644); err != nil {
		return err
	}
	if err := os.Remove(absSrc); err != nil {
		return err
	}

	if newID != haveID {
		n, err := rewriteCitations(absDir, haveKind, haveID, wantKind, newID)
		if err != nil {
			return err
		}
		if n > 0 {
			fmt.Fprintf(os.Stdout, "  rewrote %d citation(s)\n", n)
		}
	}

	fmt.Fprintf(os.Stdout, "✓ %s %s/%s → %s/%s (%s)\n", verb, haveKind, haveID, wantKind, newID, rel)
	return reindex(absDir)
}

func splitKindID(ref string) (kind, id string, ok bool) {
	i := strings.IndexByte(ref, '/')
	if i <= 0 || i == len(ref)-1 {
		return "", "", false
	}
	return ref[:i], ref[i+1:], true
}

// parseMoveTarget splits <new-kind>[/<new-path>].
func parseMoveTarget(target string) (kind, path string, err error) {
	i := strings.IndexByte(target, '/')
	if i < 0 {
		return target, "", nil
	}
	kind = target[:i]
	path = target[i+1:]
	if path == "" {
		return "", "", fmt.Errorf("trailing slash in target %q", target)
	}
	return kind, path, nil
}

// kindDirName is the folder a kind lives in under .agents/. Not a
// mechanical plural: the kind named corpus lives in corpus/, reference
// in references/, and the rest pluralize.
func kindDirName(k string) string {
	switch k {
	case "reference":
		return "references"
	case "corpus":
		return "corpus"
	}
	for _, kk := range primitive.KnownKinds {
		if string(kk) == k {
			return k + "s"
		}
	}
	return ""
}

// rewriteCitations updates every project primitive whose `corpus:` list
// cites the moved id (in any accepted form).
func rewriteCitations(absDir, oldKind, oldID, newKind, newID string) (int, error) {
	prims, _, err := primitive.Walk(absDir, config.DefaultAgentsRoot)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, p := range prims {
		if p.Provenance != "project" || len(p.Corpus) == 0 {
			continue
		}
		hit := false
		for _, ref := range p.Corpus {
			bare := strings.TrimPrefix(ref, "corpus/")
			if j := strings.IndexByte(bare, '/'); j > 0 {
				bare = bare[j+1:]
			}
			if bare == oldID {
				hit = true
				break
			}
		}
		if !hit {
			continue
		}
		abs := filepath.Join(absDir, filepath.FromSlash(p.Path))
		data, err := os.ReadFile(abs)
		if err != nil {
			return n, err
		}
		fm, body, hasFM := primitive.SplitFrontmatter(string(data))
		if !hasFM {
			continue
		}
		fm2, ferr := primitive.ReplaceCorpusRef(fm, oldKind, oldID, newKind, newID)
		if ferr != nil {
			return n, ferr
		}
		if err := os.WriteFile(abs, []byte("---\n"+fm2+"---\n"+body), 0o644); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

// reindex regenerates both index files after a move.
func reindex(absDir string) error {
	prims, warnings, err := primitive.Walk(absDir, config.DefaultAgentsRoot)
	if err != nil {
		return err
	}
	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "keystone: %s: %s\n", w.Path, w.Message)
	}
	composed, _ := primitive.Compose(prims)
	idx := primitive.Build(composed, time.Now())
	outPath := filepath.Join(absDir, config.StateRoot, config.IndexName)
	if err := primitive.Write(outPath, idx); err != nil {
		return err
	}
	lite := primitive.BuildLite(idx)
	return primitive.WriteLite(filepath.Join(absDir, config.StateRoot, config.IndexLiteName), lite)
}
