package migrations

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// Version 3.0 — community-standard vocabulary. The two-layer
// (framework-wraps-host) taxonomy collapses into one canonical set named
// for what the agent community already calls these things:
//
//   guide    → rule        (dir guides/   → rules/)
//   sensor   → hook | agent (dir sensors/ → hooks/ | agents/)
//   action   → command      (dir actions/   → commands/)
//   playbook → skill         (dir playbooks/ → skills/<id>/SKILL.md)
//   persona  → agent         (dir personas/  → agents/)
//   subagent → agent         (kind rewrite only; already under agents/)
//
// A sensor splits on its nature: a computational sensor (one with
// `host_triggers:`, which fires as a host hook) becomes a `hook`; an
// inferential review sensor (no host_triggers) becomes an `agent`.
//
// The migration also renames the `traces:` association to `corpus:` and
// creates the document workspace (.keystone/work/). It never rewrites a
// primitive that is already on the 3.0 vocabulary, so re-running is a
// no-op.

func init() {
	Register(Migration{
		Version: "3.0",
		Up:      planUp_3_0,
		Down:    planDown_3_0,
	})
}

// kindRewrite is one old→new kind mapping with its directory move.
type kindRewrite struct {
	newKind string
	newDir  string
}

// dirToOldKind maps each retired 2.x directory to the old kind it held,
// so the migration can find primitives to convert regardless of whether
// they declared `kind:` explicitly.
var oldKindDirs = map[string]string{
	"guides":    "guide",
	"sensors":   "sensor",
	"actions":   "action",
	"playbooks": "playbook",
	"personas":  "persona",
}

var (
	reTracesKey = regexp.MustCompile(`(?m)^traces:`)
)

func planUp_3_0(absDir string) (*Plan, error) {
	p := &Plan{}
	newRoot := config.DefaultHarnessRoot // .harness
	if !dirExists(filepath.Join(absDir, legacyHarnessRoot)) && !dirExists(filepath.Join(absDir, newRoot)) {
		// Fresh install — scaffolded directly from 3.0 templates.
		return p, nil
	}

	p.Add("move the harness from .keystone/ to .harness/ (one standardized root)", func(absDir string) error {
		return relocateToHarness(absDir, newRoot)
	})
	p.Add("rename primitives to the 3.0 vocabulary + move to canonical dirs", func(absDir string) error {
		return migratePrimitivesV3(filepath.Join(absDir, newRoot))
	})
	p.Add("create the document workspace (.harness/work/ + roadmaps/)", func(absDir string) error {
		for _, rel := range []string{
			filepath.Join(newRoot, "work", ".gitkeep"),
			filepath.Join(newRoot, "work", "roadmaps", ".gitkeep"),
			filepath.Join(newRoot, "documents", ".gitkeep"),
		} {
			if err := ensureFile(filepath.Join(absDir, rel)); err != nil {
				return err
			}
		}
		return nil
	})
	p.Add("rebuild INDEX.json + INDEX.lite.json on the new vocabulary", func(absDir string) error {
		return rebuildIndex(absDir, newRoot, config.KeystoneDir(newRoot))
	})
	return p, nil
}

// relocateToHarness moves a pre-3.0 `.keystone/` install to the single
// `.harness/` root: the harness subtree becomes the root, and the
// umbrella files (INDEX*, lockfile, state, context.json) move up into it.
// Idempotent — a no-op once `.keystone/harness` is gone.
func relocateToHarness(absDir, newRoot string) error {
	legacyHarnessAbs := filepath.Join(absDir, legacyHarnessRoot)
	newRootAbs := filepath.Join(absDir, newRoot)
	if !dirExists(legacyHarnessAbs) {
		return nil // already migrated
	}
	if dirExists(newRootAbs) {
		return fmt.Errorf("both %s and %s exist — resolve by hand", legacyHarnessRoot, newRoot)
	}
	if err := os.Rename(legacyHarnessAbs, newRootAbs); err != nil {
		return fmt.Errorf("move harness tree: %w", err)
	}
	// Lift the umbrella files from .keystone/ into the new root.
	legacyUmbrella := filepath.Join(absDir, legacyKeystoneDir)
	for _, name := range []string{config.IndexName, config.IndexLiteName, config.LockfileName, "context.json", "state"} {
		src := filepath.Join(legacyUmbrella, name)
		if _, err := os.Stat(src); err != nil {
			continue
		}
		if err := os.Rename(src, filepath.Join(newRootAbs, name)); err != nil {
			return fmt.Errorf("move %s: %w", name, err)
		}
	}
	// Drop the now-empty .keystone/ (ignore if other content remains).
	_ = os.Remove(legacyUmbrella)
	return nil
}

func planDown_3_0(absDir string) (*Plan, error) {
	// Down is best-effort: the kind rename is reversible by directory, but
	// the sensor split (hook vs agent) and document workspace are 3.0-only
	// additions. We reverse the directory/kind moves we can and leave the
	// rest documented.
	p := &Plan{}
	p.Add("3.0 down is not fully reversible (sensor split + documents are 3.0-only); no-op", func(string) error {
		return nil
	})
	return p, nil
}

// v3move is one planned file conversion: write newText at newPath, then
// remove oldPath if it moved.
type v3move struct {
	oldPath, newPath, newText string
}

// migratePrimitivesV3 walks the harness tree and converts every retired
// primitive in place: rewrites its `kind:`, renames `traces:` → `corpus:`,
// and moves the file to its 3.0 canonical directory.
func migratePrimitivesV3(harnessAbs string) error {
	var moves []v3move
	err := filepath.WalkDir(harnessAbs, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		m, planErr := planV3Move(harnessAbs, path)
		if planErr != nil {
			return planErr
		}
		if m != nil {
			moves = append(moves, *m)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return applyV3Moves(moves)
}

// planV3Move resolves the conversion for one file, or nil when the file
// is not a retired primitive (already migrated / unrecognized dir).
func planV3Move(harnessAbs, path string) (*v3move, error) {
	rel := filepath.ToSlash(mustRel(harnessAbs, path))
	oldKind, isOldDir := oldKindDirs[firstSegment(rel)]
	if !isOldDir {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	fm, _, _ := primitive.Parse(string(data))
	rw := rewriteForV3(oldKind, fm)
	if rw == nil {
		return nil, nil
	}
	newRel := newRelPath(rel, firstSegment(rel), rw.newDir)
	return &v3move{
		oldPath: path,
		newPath: filepath.Join(harnessAbs, filepath.FromSlash(newRel)),
		newText: applyV3Rewrites(string(data), oldKind, rw.newKind, fm.ID),
	}, nil
}

// applyV3Moves writes each planned conversion and removes the source when
// the file moved directories.
func applyV3Moves(moves []v3move) error {
	for _, m := range moves {
		if err := os.MkdirAll(filepath.Dir(m.newPath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(m.newPath, []byte(m.newText), 0o644); err != nil {
			return err
		}
		if m.newPath == m.oldPath {
			continue
		}
		if err := os.Remove(m.oldPath); err != nil {
			return err
		}
	}
	return nil
}

func mustRel(base, path string) string {
	rel, _ := filepath.Rel(base, path)
	return rel
}

// rewriteForV3 resolves the new kind + directory for a primitive, or nil
// when there is nothing to do. The sensor split keys on host_triggers:
// computational (fires as a host hook) → hook; inferential → agent.
func rewriteForV3(oldKind string, fm primitive.Frontmatter) *kindRewrite {
	switch oldKind {
	case "guide":
		return &kindRewrite{"rule", "rules"}
	case "action":
		return &kindRewrite{"command", "commands"}
	case "playbook":
		return &kindRewrite{"skill", "skills"}
	case "persona":
		return &kindRewrite{"agent", "agents"}
	case "sensor":
		if len(fm.HostTriggers) > 0 {
			return &kindRewrite{"hook", "hooks"}
		}
		return &kindRewrite{"agent", "agents"}
	}
	return nil
}

var (
	reToolsKey    = regexp.MustCompile(`(?m)^tools:`)
	reTriggersKey = regexp.MustCompile(`(?m)^triggers:`)
)

// applyV3Rewrites rewrites the frontmatter kind line, renames traces: →
// corpus:, and supplies the fields the new kind requires but the old one
// lacked: an agent needs tools:, a skill needs triggers:. Body untouched.
func applyV3Rewrites(text, oldKind, newKind, id string) string {
	reKind := regexp.MustCompile(`(?m)^kind:[ \t]*` + regexp.QuoteMeta(oldKind) + `[ \t]*$`)
	text = reKind.ReplaceAllString(text, "kind: "+newKind)
	text = reTracesKey.ReplaceAllString(text, "corpus:")

	if newKind == "agent" && !reToolsKey.MatchString(text) {
		text = appendFrontmatterLines(text, "tools:\n  - Read\n  - Grep")
	}
	if newKind == "skill" && !reTriggersKey.MatchString(text) {
		text = appendFrontmatterLines(text, "triggers:\n  - "+id)
	}
	return text
}

// appendFrontmatterLines inserts lines at the end of the frontmatter
// block (before the closing fence). No-op when there is no frontmatter.
func appendFrontmatterLines(text, lines string) string {
	fm, body, ok := primitive.SplitFrontmatter(text)
	if !ok {
		return text
	}
	return "---\n" + fm + lines + "\n---\n" + body
}

// newRelPath rewrites a primitive's harness-relative path for its new
// directory. Playbooks become skills/<id>/SKILL.md (skills are
// directories); every other kind just swaps its first path segment.
func newRelPath(rel, oldSeg, newDir string) string {
	rest := strings.TrimPrefix(rel, oldSeg+"/")
	if newDir == "skills" {
		base := strings.TrimSuffix(filepath.Base(rest), ".md")
		return "skills/" + base + "/SKILL.md"
	}
	return newDir + "/" + rest
}

func firstSegment(rel string) string {
	if i := strings.IndexByte(rel, '/'); i >= 0 {
		return rel[:i]
	}
	return rel
}

// ensureFile creates an empty file (and parents) if it does not exist.
func ensureFile(absPath string) error {
	if _, err := os.Stat(absPath); err == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(absPath, nil, 0o644)
}
