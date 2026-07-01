package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// runList handles `keystone list [<kind>] [--tag <tag>]... [--dir <path>]`.
//
// Walks the charter, applies composition, and prints one line per
// matching primitive: `<kind>  <id>  <description>` with optional
// `[tag1, tag2]` suffix when tags are declared.
//
// Filtering:
//   - Optional positional argument <kind> narrows to that kind only.
//   - Multiple `--tag X` flags filter AND — a primitive must declare
//     every requested tag to appear.
//   - No filter ⇒ every primitive.
//
// Output is sorted (kind, then id) for deterministic eyeballing and
// diff-friendliness.
func runList(args []string) error {
	dir := "."
	var kindFilter string
	var tags []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--help" || a == "-h":
			printListUsage(os.Stdout)
			return nil
		case a == "--dir":
			if i+1 >= len(args) {
				return fmt.Errorf("flag %s requires a value", a)
			}
			dir = args[i+1]
			i++
		case strings.HasPrefix(a, "--dir="):
			dir = strings.TrimPrefix(a, "--dir=")
		case a == "--tag":
			if i+1 >= len(args) {
				return fmt.Errorf("flag %s requires a value", a)
			}
			tags = append(tags, args[i+1])
			i++
		case strings.HasPrefix(a, "--tag="):
			tags = append(tags, strings.TrimPrefix(a, "--tag="))
		case strings.HasPrefix(a, "-"):
			return fmt.Errorf("unknown flag %s", a)
		default:
			if kindFilter != "" {
				return fmt.Errorf("multiple kind arguments (got %q and %q) — only one is supported", kindFilter, a)
			}
			kindFilter = a
		}
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}

	primitives, warnings, err := primitive.Walk(absDir, config.DefaultCharterRoot)
	if err != nil {
		return err
	}
	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "keystone list: %s: %s\n", w.Path, w.Message)
	}
	composed, _ := primitive.Compose(primitives)

	matched := filterPrimitives(composed, kindFilter, tags)
	sort.SliceStable(matched, func(i, j int) bool {
		if matched[i].Kind != matched[j].Kind {
			return matched[i].Kind < matched[j].Kind
		}
		return matched[i].ID < matched[j].ID
	})

	if len(matched) == 0 {
		fmt.Fprintln(os.Stdout, "(no matching primitives)")
		return nil
	}

	// Column widths sized to the longest match — keeps the table
	// readable without word-wrapping.
	kindW, idW := 0, 0
	for _, p := range matched {
		if l := len(p.Kind); l > kindW {
			kindW = l
		}
		if l := len(p.ID); l > idW {
			idW = l
		}
	}

	for _, p := range matched {
		desc := p.Description
		if desc == "" {
			desc = "—"
		}
		line := fmt.Sprintf("%-*s  %-*s  %s", kindW, p.Kind, idW, p.ID, desc)
		if len(p.Tags) > 0 {
			line += fmt.Sprintf("  [%s]", strings.Join(p.Tags, ", "))
		}
		fmt.Fprintln(os.Stdout, line)
	}
	return nil
}

// filterPrimitives returns the subset matching the kind filter (if
// any) and ALL requested tags. The tag check is AND, not OR — a
// primitive lacking any one of the requested tags is dropped.
func filterPrimitives(primitives []primitive.Primitive, kind string, tags []string) []primitive.Primitive {
	var out []primitive.Primitive
	for _, p := range primitives {
		if kind != "" && p.Kind != kind {
			continue
		}
		if !hasAllTags(p.Tags, tags) {
			continue
		}
		out = append(out, p)
	}
	return out
}

func hasAllTags(have, want []string) bool {
	if len(want) == 0 {
		return true
	}
	set := make(map[string]struct{}, len(have))
	for _, t := range have {
		set[t] = struct{}{}
	}
	for _, w := range want {
		if _, ok := set[w]; !ok {
			return false
		}
	}
	return true
}

func printListUsage(w *os.File) {
	fmt.Fprint(w, `keystone list — list primitives by kind and / or tag

Usage:
  keystone list [<kind>] [--tag <tag>]... [--dir <path>]

Walks the charter, applies composition, and prints one line per
matching primitive (kind / id / description, with declared tags in
brackets).

Filters:
  <kind>            Optional positional. One of guide, sensor, action,
                    playbook, persona, corpus, eval, source, concern,
                    rule, skill, subagent, command.
  --tag <tag>       Repeatable. Multiple --tag arguments AND together —
                    a primitive must declare every requested tag.
  --dir <path>      Project root (defaults to cwd).

Examples:
  keystone list                          # all primitives
  keystone list sensor                   # all sensors
  keystone list --tag security           # any primitive tagged 'security'
  keystone list persona --tag review     # personas tagged 'review'
  keystone list guide --tag onboarding --tag security
`)
}
