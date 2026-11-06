package main

import (
	"fmt"
	"os"
	"path/filepath"
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
//   - Optional positional argument <kind> narrows to that kind, or to a
//     category (constraints / capabilities).
//   - Multiple `--tag X` flags filter AND — a primitive must declare
//     every requested tag to appear.
//   - No filter ⇒ every primitive.
//
// Output is grouped by category (constraints first), then kind, then id.
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

	primitives, warnings, err := primitive.Walk(absDir, config.DefaultAgentsRoot)
	if err != nil {
		return err
	}
	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "keystone list: %s: %s\n", w.Path, w.Message)
	}
	composed, _ := primitive.Compose(primitives)

	matched := filterPrimitives(composed, kindFilter, tags)
	primitive.SortByCategory(matched)

	printListed(matched)
	return nil
}

// printListed prints the matches as a table under a heading per category.
func printListed(matched []primitive.Primitive) {
	if len(matched) == 0 {
		fmt.Fprintln(os.Stdout, "(no matching primitives)")
		return
	}

	kindW, idW := listColumnWidths(matched)
	heading := ""
	for _, p := range matched {
		if h := primitive.CategoryTitle(p.Kind); h != heading {
			if heading != "" {
				fmt.Fprintln(os.Stdout)
			}
			fmt.Fprintf(os.Stdout, "# %s\n", h)
			heading = h
		}
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
}

// listColumnWidths sizes the columns to the longest match, keeping the
// table readable without word-wrapping.
func listColumnWidths(matched []primitive.Primitive) (kindW, idW int) {
	for _, p := range matched {
		kindW = max(kindW, len(p.Kind))
		idW = max(idW, len(p.ID))
	}
	return kindW, idW
}

// filterPrimitives returns the subset matching the kind filter (if
// any) and ALL requested tags. The tag check is AND, not OR — a
// primitive lacking any one of the requested tags is dropped.
func filterPrimitives(primitives []primitive.Primitive, kind string, tags []string) []primitive.Primitive {
	var out []primitive.Primitive
	for _, p := range primitives {
		if kind != "" && !matchesKindOrCategory(p.Kind, kind) {
			continue
		}
		if !hasAllTags(p.Tags, tags) {
			continue
		}
		out = append(out, p)
	}
	return out
}

// categoryFilters lets `keystone list` take a category in place of a kind.
var categoryFilters = map[string]primitive.Category{
	"constraint": primitive.CategoryConstraint, "constraints": primitive.CategoryConstraint,
	"capability": primitive.CategoryCapability, "capabilities": primitive.CategoryCapability,
}

// matchesKindOrCategory: the filter is a kind (or its 4.x spelling) or a
// category.
func matchesKindOrCategory(kind, filter string) bool {
	if c, ok := categoryFilters[filter]; ok {
		return primitive.CategoryOf(primitive.Kind(kind)) == c
	}
	return primitive.IsKind(kind, filter)
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
  keystone list [<kind> | constraints | capabilities] [--tag <tag>]... [--dir <path>]

Walks the charter, applies composition, and prints one line per
matching primitive (kind / id / description, with declared tags in
brackets), constraints first, under a heading per category.

Filters:
  <kind>            Optional positional. A constraint kind (rule, reference,
                    hook, permission, eval), a capability kind (skill,
                    command, subagent, tool, document), or a category.
  --tag <tag>       Repeatable. Multiple --tag arguments AND together —
                    a primitive must declare every requested tag.
  --dir <path>      Project root (defaults to cwd).

Examples:
  keystone list                          # all primitives
  keystone list constraints              # rules, references, hooks, permissions, evals
  keystone list hook                     # all hooks
  keystone list --tag security           # any primitive tagged 'security'
  keystone list rule --tag onboarding --tag security
`)
}
