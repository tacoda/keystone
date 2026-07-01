// Package eval runs charter evals — measurable checks of "did this
// charter change actually do what we expected?"
//
// Phase A (this file): static + sensor levels. Static walks the
// charter through the eval's fixture file set and reports which
// primitives would activate; sensor runs declared sensors against
// the fixture and captures exit codes.
//
// Phase B (deferred to 2.1): --baseline diff mode (compare two refs)
// and agent-level evals (LLM-driven, judge-graded).
package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// Spec is one parsed eval (`EVAL.md` + sibling `expected.json`).
type Spec struct {
	primitive.Primitive
	// Level is one of: "static", "sensor", "agent". Phase A
	// implements static + sensor; agent runs are skipped with a
	// "not-implemented" status.
	Level    string   `json:"level"`
	Levels   []string `json:"levels,omitempty"` // multi-level evals: parsed from frontmatter `levels:` array
	Expected Expected `json:"expected"`
	// Dir is the absolute directory holding this eval's files.
	Dir string `json:"dir"`
}

// Expected is the assertions block — typically supplied via a
// sibling `expected.json`, but can also be inlined in EVAL.md
// frontmatter under an `expected:` key.
type Expected struct {
	// Static checks.
	Static StaticExpected `json:"static,omitempty"`
	// Sensor checks.
	Sensors []SensorExpected `json:"sensors,omitempty"`
	// Touched files the static check uses for glob matching. Repo-
	// relative POSIX paths.
	TouchedFiles []string `json:"touched_files,omitempty"`
}

type StaticExpected struct {
	// RulesFired lists primitive ids (kind=guide or kind=rule)
	// whose `globs:` should match at least one TouchedFile.
	RulesFired []string `json:"rules_fired,omitempty"`
	// RulesSilent lists primitive ids whose globs should NOT match.
	RulesSilent []string `json:"rules_silent,omitempty"`
}

type SensorExpected struct {
	// Sensor id (kind=sensor primitive id).
	ID string `json:"id"`
	// ExpectExit is the required exit code. 0 = pass.
	ExpectExit int `json:"expect_exit"`
	// MustContain — substrings the sensor's stdout/stderr must
	// include.
	MustContain []string `json:"must_contain,omitempty"`
}

// Report is the result of one eval run.
type Report struct {
	StartedAt string   `json:"started_at"`
	EndedAt   string   `json:"ended_at"`
	Results   []Result `json:"results"`
	Passed    int      `json:"passed"`
	Failed    int      `json:"failed"`
	Skipped   int      `json:"skipped"`
}

// Result is one eval's outcome.
type Result struct {
	ID       string   `json:"id"`
	Level    string   `json:"level"`
	Status   string   `json:"status"` // "pass" | "fail" | "skip"
	Messages []string `json:"messages,omitempty"`
	Duration string   `json:"duration"`
}

// LoadAll walks the charter, parses every EVAL.md, and pairs each
// with its sibling expected.json (if present).
func LoadAll(projectDir string) ([]Spec, error) {
	primitives, _, err := primitive.Walk(projectDir, config.DefaultCharterRoot)
	if err != nil {
		return nil, err
	}
	var out []Spec
	for _, p := range primitives {
		if p.Kind != string(primitive.KindEval) {
			continue
		}
		spec, err := buildSpec(projectDir, p)
		if err != nil {
			return nil, fmt.Errorf("load %s: %w", p.ID, err)
		}
		out = append(out, spec)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func buildSpec(projectDir string, p primitive.Primitive) (Spec, error) {
	abs := filepath.Join(projectDir, p.Path)
	dir := filepath.Dir(abs)
	s := Spec{
		Primitive: p,
		Dir:       dir,
	}
	// Frontmatter may carry `level:` (single) — already in
	// p.Frontmatter under no canonical typed field. Re-read it
	// loosely from the raw markdown to pull both `level` and any
	// `levels:` list.
	if level, levels, err := scanLevels(abs); err == nil {
		s.Level = level
		s.Levels = levels
	}
	if s.Level == "" && len(s.Levels) == 0 {
		s.Level = "static"
	}
	// expected.json sibling.
	if data, err := os.ReadFile(filepath.Join(dir, "expected.json")); err == nil {
		_ = json.Unmarshal(data, &s.Expected)
	}
	return s, nil
}

// scanLevels pulls `level:` and `levels:` out of the EVAL.md
// frontmatter without going through the full primitive parser
// (which doesn't expose either field).
func scanLevels(path string) (level string, levels []string, err error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return "", nil, err
	}
	src := string(body)
	if !strings.HasPrefix(src, "---") {
		return "", nil, nil
	}
	end := strings.Index(src[3:], "---")
	if end < 0 {
		return "", nil, nil
	}
	fm := src[3 : 3+end]
	inList := false
	for _, line := range strings.Split(fm, "\n") {
		trim := strings.TrimSpace(line)
		if inList {
			if strings.HasPrefix(trim, "- ") {
				levels = append(levels, strings.TrimSpace(strings.TrimPrefix(trim, "- ")))
				continue
			}
			inList = false
		}
		switch {
		case strings.HasPrefix(trim, "level:"):
			level = strings.TrimSpace(strings.TrimPrefix(trim, "level:"))
		case strings.HasPrefix(trim, "levels:"):
			rest := strings.TrimSpace(strings.TrimPrefix(trim, "levels:"))
			if strings.HasPrefix(rest, "[") && strings.HasSuffix(rest, "]") {
				items := strings.Split(strings.Trim(rest, "[]"), ",")
				for _, it := range items {
					levels = append(levels, strings.TrimSpace(it))
				}
			} else if rest == "" {
				inList = true
			}
		}
	}
	return level, levels, nil
}

// Run executes every spec in `specs` against the charter at
// `projectDir`. Filter (substring of id) narrows the set; empty
// runs all.
func Run(ctx context.Context, projectDir string, specs []Spec, filter string) Report {
	rep := Report{StartedAt: time.Now().UTC().Format(time.RFC3339)}
	for _, s := range specs {
		if filter != "" && !strings.Contains(s.ID, filter) {
			continue
		}
		t0 := time.Now()
		levels := s.Levels
		if len(levels) == 0 {
			levels = []string{s.Level}
		}
		for _, lvl := range levels {
			r := runOne(ctx, projectDir, s, lvl)
			r.Duration = time.Since(t0).Round(time.Millisecond).String()
			switch r.Status {
			case "pass":
				rep.Passed++
			case "fail":
				rep.Failed++
			case "skip":
				rep.Skipped++
			}
			rep.Results = append(rep.Results, r)
		}
	}
	rep.EndedAt = time.Now().UTC().Format(time.RFC3339)
	return rep
}

func runOne(ctx context.Context, projectDir string, s Spec, level string) Result {
	r := Result{ID: s.ID, Level: level}
	switch level {
	case "static":
		runStatic(projectDir, s, &r)
	case "sensor":
		runSensors(ctx, projectDir, s, &r)
	case "agent":
		r.Status = "skip"
		r.Messages = append(r.Messages, "agent-level evals not yet implemented (Phase B)")
	default:
		r.Status = "skip"
		r.Messages = append(r.Messages, fmt.Sprintf("unknown level %q", level))
	}
	return r
}

// runStatic walks the charter, finds rule-class primitives, and
// checks each touched file against their `globs:`. Pass when every
// id in RulesFired matches at least one touched file AND every id
// in RulesSilent matches none.
func runStatic(projectDir string, s Spec, r *Result) {
	primitives, _, err := primitive.Walk(projectDir, config.DefaultCharterRoot)
	if err != nil {
		r.Status = "fail"
		r.Messages = append(r.Messages, "walk: "+err.Error())
		return
	}
	byID := map[string]primitive.Primitive{}
	for _, p := range primitives {
		if p.Kind == "guide" || p.Kind == "rule" {
			byID[p.Kind+"/"+p.ID] = p
			byID[p.ID] = p
		}
	}

	failed := false
	for _, id := range s.Expected.Static.RulesFired {
		p, ok := byID[id]
		if !ok {
			r.Messages = append(r.Messages, fmt.Sprintf("✗ rules_fired: no primitive %q", id))
			failed = true
			continue
		}
		if !globsHit(p.Globs, s.Expected.TouchedFiles) {
			r.Messages = append(r.Messages, fmt.Sprintf("✗ rules_fired: %q did NOT match any touched file", id))
			failed = true
			continue
		}
		r.Messages = append(r.Messages, fmt.Sprintf("✓ rules_fired: %q activated", id))
	}
	for _, id := range s.Expected.Static.RulesSilent {
		p, ok := byID[id]
		if !ok {
			continue // missing means it can't fire — silent by default
		}
		if globsHit(p.Globs, s.Expected.TouchedFiles) {
			r.Messages = append(r.Messages, fmt.Sprintf("✗ rules_silent: %q SHOULD NOT have activated", id))
			failed = true
			continue
		}
		r.Messages = append(r.Messages, fmt.Sprintf("✓ rules_silent: %q stayed silent", id))
	}
	if failed {
		r.Status = "fail"
		return
	}
	if len(r.Messages) == 0 {
		r.Messages = append(r.Messages, "no static assertions declared")
	}
	r.Status = "pass"
}

// globsHit returns true if any of the patterns matches at least one
// file path. Uses Go's filepath.Match (glob; no doublestar). For
// `**` patterns, this is best-effort — falls back to substring
// match when the pattern contains `**`.
func globsHit(patterns, touched []string) bool {
	for _, pat := range patterns {
		negate := strings.HasPrefix(pat, "!")
		p := strings.TrimPrefix(pat, "!")
		for _, f := range touched {
			matched := matchGlob(p, f)
			if matched && !negate {
				return true
			}
		}
	}
	return false
}

func matchGlob(pat, path string) bool {
	// `**` → substring fallback (rough but acceptable for static evals).
	if strings.Contains(pat, "**") {
		prefix := pat
		if idx := strings.Index(pat, "**"); idx >= 0 {
			prefix = strings.TrimSuffix(pat[:idx], "/")
		}
		return strings.HasPrefix(path, prefix)
	}
	ok, _ := filepath.Match(pat, path)
	return ok
}

// runSensors runs each declared sensor as a child process and
// captures exit codes + output. Sensor primitives store their
// shell command in the markdown body; we extract it best-effort
// from the first fenced code block.
func runSensors(ctx context.Context, projectDir string, s Spec, r *Result) {
	primitives, _, err := primitive.Walk(projectDir, config.DefaultCharterRoot)
	if err != nil {
		r.Status = "fail"
		r.Messages = append(r.Messages, "walk: "+err.Error())
		return
	}
	byID := map[string]primitive.Primitive{}
	for _, p := range primitives {
		if p.Kind == "sensor" {
			byID[p.ID] = p
		}
	}
	failed := false
	for _, want := range s.Expected.Sensors {
		p, ok := byID[want.ID]
		if !ok {
			r.Messages = append(r.Messages, fmt.Sprintf("✗ sensor %q not found", want.ID))
			failed = true
			continue
		}
		cmd := extractSensorCommand(filepath.Join(projectDir, p.Path))
		if cmd == "" {
			r.Messages = append(r.Messages, fmt.Sprintf("✗ sensor %q has no shell command block", want.ID))
			failed = true
			continue
		}
		out, exit := runShell(ctx, projectDir, cmd)
		if exit != want.ExpectExit {
			r.Messages = append(r.Messages, fmt.Sprintf("✗ sensor %q: exit=%d want=%d", want.ID, exit, want.ExpectExit))
			failed = true
		} else {
			r.Messages = append(r.Messages, fmt.Sprintf("✓ sensor %q exit=%d", want.ID, exit))
		}
		for _, sub := range want.MustContain {
			if !strings.Contains(out, sub) {
				r.Messages = append(r.Messages, fmt.Sprintf("✗ sensor %q output missing %q", want.ID, sub))
				failed = true
			}
		}
	}
	if failed {
		r.Status = "fail"
	} else if len(r.Messages) == 0 {
		r.Status = "skip"
		r.Messages = append(r.Messages, "no sensor assertions declared")
	} else {
		r.Status = "pass"
	}
}

// extractSensorCommand pulls the first fenced ``` block out of a
// sensor body. Sensors author the command inline; this is the
// extraction contract we agree on.
func extractSensorCommand(path string) string {
	body, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	src := string(body)
	open := strings.Index(src, "```")
	if open < 0 {
		return ""
	}
	rest := src[open+3:]
	// Skip language tag if present.
	if nl := strings.Index(rest, "\n"); nl >= 0 {
		rest = rest[nl+1:]
	}
	close := strings.Index(rest, "```")
	if close < 0 {
		return ""
	}
	return strings.TrimSpace(rest[:close])
}

func runShell(ctx context.Context, dir, cmdline string) (string, int) {
	cmd := exec.CommandContext(ctx, "sh", "-c", cmdline)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	exit := 0
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			exit = exiterr.ExitCode()
		} else {
			exit = -1
		}
	}
	return string(out), exit
}

// RenderMarkdown formats a report as a human-readable markdown table.
func RenderMarkdown(r Report) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# keystone eval report\n\n")
	fmt.Fprintf(&b, "- started: %s\n- ended:   %s\n- passed:  %d\n- failed:  %d\n- skipped: %d\n\n", r.StartedAt, r.EndedAt, r.Passed, r.Failed, r.Skipped)
	fmt.Fprintln(&b, "| id | level | status | duration |")
	fmt.Fprintln(&b, "|---|---|---|---|")
	for _, res := range r.Results {
		fmt.Fprintf(&b, "| %s | %s | %s | %s |\n", res.ID, res.Level, res.Status, res.Duration)
	}
	fmt.Fprintln(&b)
	for _, res := range r.Results {
		fmt.Fprintf(&b, "## %s (%s — %s)\n\n", res.ID, res.Level, res.Status)
		for _, m := range res.Messages {
			fmt.Fprintf(&b, "- %s\n", m)
		}
		fmt.Fprintln(&b)
	}
	return b.String()
}

// EvalsRoot returns the absolute dir holding the evals tree.
func EvalsRoot(projectDir string) string {
	return filepath.Join(projectDir, config.DefaultCharterRoot, "evals")
}

// EnsureRoot makes sure the evals directory exists. Returns the
// path. No-op if already present.
func EnsureRoot(projectDir string) (string, error) {
	root := EvalsRoot(projectDir)
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", err
	}
	return root, nil
}

// ListIDs walks the evals dir and returns just the ids (dir names
// holding an EVAL.md). Cheap alternative to LoadAll when the
// caller only needs identifiers.
func ListIDs(projectDir string) ([]string, error) {
	root := EvalsRoot(projectDir)
	var out []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if d.IsDir() || filepath.Base(path) != "EVAL.md" {
			return nil
		}
		out = append(out, filepath.Base(filepath.Dir(path)))
		return nil
	})
	sort.Strings(out)
	return out, err
}
