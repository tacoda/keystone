package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

// learningInbox is the additive step of the Learning flywheel: unsorted
// candidates awaiting synthesis.
const learningInbox = ".keystone/learning/inbox"

// runLearn handles `keystone learn <slug> [--source <text>] [--layer <l>]
// [--globs <g>...] [--body TEXT|-]`. The Learning flywheel's additive
// step: writes the smallest record of what surprised you to
// learning/inbox/<captured>-<slug>.md, so a later synthesize can promote
// it. The frontmatter matches the canonical shape keystone-learn's skill
// defines — one shape, every writer.
func runLearn(args []string) error {
	slug := ""
	var source, layer, body, dir string
	var globs []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--dir":
			i++
			if i >= len(args) {
				return fmt.Errorf("--dir requires a value")
			}
			dir = args[i]
		case a == "--source":
			i++
			if i >= len(args) {
				return fmt.Errorf("--source requires a value")
			}
			source = args[i]
		case a == "--layer":
			i++
			if i >= len(args) {
				return fmt.Errorf("--layer requires a value")
			}
			layer = args[i]
		case a == "--globs":
			i++
			for i < len(args) && !strings.HasPrefix(args[i], "--") {
				globs = append(globs, args[i])
				i++
			}
			i--
		case a == "--body":
			i++
			if i >= len(args) {
				return fmt.Errorf("--body requires TEXT or -")
			}
			if args[i] == "-" {
				b, err := readStdin()
				if err != nil {
					return err
				}
				body = b
			} else {
				body = args[i]
			}
		case strings.HasPrefix(a, "--dir="):
			dir = strings.TrimPrefix(a, "--dir=")
		case strings.HasPrefix(a, "--source="):
			source = strings.TrimPrefix(a, "--source=")
		case strings.HasPrefix(a, "--layer="):
			layer = strings.TrimPrefix(a, "--layer=")
		case strings.HasPrefix(a, "--body="):
			body = strings.TrimPrefix(a, "--body=")
		default:
			if strings.HasPrefix(a, "-") {
				return fmt.Errorf("unknown flag %q", a)
			}
			if slug != "" {
				return fmt.Errorf("learn takes one <slug>; also got %q", a)
			}
			slug = a
		}
	}
	if slug == "" {
		return fmt.Errorf("`keystone learn` requires a <slug>")
	}
	if body == "" {
		return fmt.Errorf("--body TEXT (or - for stdin) is required — a candidate without what happened is not a record")
	}
	if source == "" {
		source = "keystone learn"
	}
	if inv := invalidLayer(layer); inv != "" {
		return fmt.Errorf("--layer %q %s", layer, inv)
	}

	if dir == "" {
		dir = "."
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	stamp := time.Now().UTC().Format("2006-01-02T15-04-05")
	rel := filepath.Join(learningInbox, stamp+"-"+slug+".md")

	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "captured: %s\n", time.Now().UTC().Format("2006-01-02"))
	fmt.Fprintf(&b, "source: %s\n", source)
	if layer != "" {
		fmt.Fprintf(&b, "proposed-layer: %s\n", layer)
	}
	if len(globs) > 0 {
		b.WriteString("proposed-globs:\n")
		for _, g := range globs {
			fmt.Fprintf(&b, "  - %q\n", g)
		}
	}
	b.WriteString("---\n\n")
	b.WriteString("## What happened\n\n" + body + "\n\n")
	b.WriteString("## Why it matters\n\n<the principle, idiom, or rule this implies>\n\n")
	b.WriteString("## Proposed change\n\n<the smallest charter edit that would prevent the next incident>\n")

	if err := os.MkdirAll(filepath.Dir(filepath.Join(absDir, rel)), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(absDir, rel), []byte(b.String()), 0o644); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "  wrote: %s\n", rel)
	return nil
}

// runLearningList handles `keystone learning list` — the inbox, newest
// first, one line per candidate. Promotion is synthesize's judgment, not
// a command.
func runLearningList(args []string) error {
	absDir, err := filepath.Abs(dirArg(args))
	if err != nil {
		return err
	}
	inbox := filepath.Join(absDir, learningInbox)
	entries, err := os.ReadDir(inbox)
	if os.IsNotExist(err) {
		fmt.Fprintln(os.Stdout, "(learning inbox empty)")
		return nil
	}
	if err != nil {
		return err
	}
	type item struct {
		path string
		mt   time.Time
	}
	var items []item
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		items = append(items, item{filepath.Join(learningInbox, e.Name()), info.ModTime()})
	}
	if len(items) == 0 {
		fmt.Fprintln(os.Stdout, "(learning inbox empty)")
		return nil
	}
	sort.Slice(items, func(i, j int) bool { return items[i].mt.After(items[j].mt) })
	for _, it := range items {
		layer, source := candidateFrontmatter(filepath.Join(absDir, it.path))
		line := "  " + it.path
		if layer != "" {
			line += "  → " + layer
		}
		if source != "" {
			line += "  (" + source + ")"
		}
		fmt.Fprintln(os.Stdout, line)
	}
	return nil
}

// candidateFrontmatter plucks proposed-layer and source from a
// candidate's frontmatter lines. Frontmatter seek only — the body is
// never read.
func candidateFrontmatter(abs string) (layer, source string) {
	data, err := os.ReadFile(abs)
	if err != nil {
		return "", ""
	}
	fm, _, ok := primitive.SplitFrontmatter(string(data))
	if !ok {
		return "", ""
	}
	for _, line := range strings.Split(fm, "\n") {
		line = strings.TrimSpace(line)
		if v, ok := strings.CutPrefix(line, "proposed-layer:"); ok {
			layer = strings.TrimSpace(v)
		}
		if v, ok := strings.CutPrefix(line, "source:"); ok {
			source = strings.TrimSpace(v)
		}
	}
	return layer, source
}

// invalidLayer checks --layer against the names keystone promotes into.
// A hook is allowed: findings can propose hook hardening. An empty value
// leaves the choice to synthesize's judgment.
func invalidLayer(layer string) string {
	switch layer {
	case "", "references/principles", "hook":
		return ""
	}
	if strings.HasPrefix(layer, "rules/") {
		return ""
	}
	return "invalid (references/principles | rules/... | hook | empty)"
}

// readStdin reads standard input when piped; an empty string when it is
// a terminal.
func readStdin() (string, error) {
	info, err := os.Stdin.Stat()
	if err != nil || info.Mode()&os.ModeCharDevice != 0 {
		return "", nil
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
