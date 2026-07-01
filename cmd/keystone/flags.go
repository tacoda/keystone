package main

import (
	"fmt"
	"strings"

	"github.com/tacoda/keystone/internal/framework/config"
)

// initFlags holds the raw, parsed CLI inputs for `keystone init`.
// Empty values mean "unset" — they will either be detected, prompted
// for interactively, or left blank (depending on category and TTY).
type initFlags struct {
	dir         string
	reset       bool // --reset: destructive overwrite of an existing charter
	confirm     bool // --i-understand-this-is-destructive: pairs with --reset
	charterRoot string

	// selections: category ID → list of chosen labels.
	// Single-select categories store at most one entry.
	selections Selections

	// policies: --policy arguments collected verbatim, in CLI order.
	// Free-form refs (URLs), not validated against any catalog. Empty if no
	// policies were requested.
	policies []string
}

// parseInitArgs parses `init`'s args. Behavior:
//   - One optional positional: the destination directory.
//   - --force: boolean.
//   - One flag per Category (named by Category.ID).
//     For multi-select categories, the value is a comma-separated list.
//     For single-select categories, the value is one label.
func parseInitArgs(args []string) (*initFlags, error) {
	flags := &initFlags{
		dir:         ".",
		charterRoot: config.DefaultCharterRoot,
		selections:  Selections{},
	}

	// Build the set of recognized flag names from the catalog.
	valueFlags := map[string]string{} // flag name → category ID
	for _, c := range categories {
		valueFlags["--"+c.ID] = c.ID
	}

	positional := []string{}

	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--reset":
			flags.reset = true

		case a == "--i-understand-this-is-destructive":
			flags.confirm = true

		case a == "--force" || a == "-force":
			return nil, fmt.Errorf("--force was removed in 1.0; the default now leaves existing files in place. " +
				"To destructively rewrite an existing charter, pass --reset --i-understand-this-is-destructive.")

		case a == "--policy":
			// --policy <ref> — repeatable, free-form (not a category).
			if i+1 >= len(args) {
				return nil, fmt.Errorf("flag %s requires a value", a)
			}
			flags.policies = append(flags.policies, args[i+1])
			i++

		case strings.HasPrefix(a, "--policy="):
			flags.policies = append(flags.policies, strings.TrimPrefix(a, "--policy="))

		case a == "--charter-root" || strings.HasPrefix(a, "--charter-root="):
			return nil, fmt.Errorf("--charter-root was removed in 2.0; the charter layout is fixed at .charter/")

		case strings.HasPrefix(a, "--") && strings.Contains(a, "="):
			// --flag=value form
			eq := strings.IndexByte(a, '=')
			name, value := a[:eq], a[eq+1:]
			catID, ok := valueFlags[name]
			if !ok {
				return nil, fmt.Errorf("unknown flag %s", name)
			}
			if err := assignValue(flags.selections, catID, value); err != nil {
				return nil, err
			}

		case valueFlags[a] != "":
			// --flag value form
			if i+1 >= len(args) {
				return nil, fmt.Errorf("flag %s requires a value", a)
			}
			catID := valueFlags[a]
			if err := assignValue(flags.selections, catID, args[i+1]); err != nil {
				return nil, err
			}
			i++

		case strings.HasPrefix(a, "-"):
			return nil, fmt.Errorf("unknown flag %s", a)

		default:
			positional = append(positional, a)
		}
	}

	if len(positional) > 1 {
		return nil, fmt.Errorf("init takes at most one positional argument (got %d)", len(positional))
	}
	if len(positional) == 1 {
		flags.dir = positional[0]
	}

	if err := validateSelections(flags.selections); err != nil {
		return nil, err
	}

	if flags.reset && !flags.confirm {
		return nil, fmt.Errorf("--reset is destructive (it removes and rewrites the existing charter). " +
			"Re-run with --reset --i-understand-this-is-destructive to confirm.")
	}
	if flags.confirm && !flags.reset {
		return nil, fmt.Errorf("--i-understand-this-is-destructive only applies alongside --reset")
	}

	return flags, nil
}

// assignValue parses a flag value (possibly comma-separated for multi-select)
// and stores the resulting labels into sel under catID. Duplicates within a
// single flag invocation are dropped.
func assignValue(sel Selections, catID, raw string) error {
	cat := categoryByID(catID)
	if cat == nil {
		return fmt.Errorf("unknown category %q", catID)
	}

	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v == "" {
			continue
		}
		if seen[v] {
			continue
		}
		seen[v] = true
		values = append(values, v)
	}

	if !cat.MultiSelect && len(values) > 1 {
		return fmt.Errorf("--%s accepts exactly one value (got %d)", cat.ID, len(values))
	}

	sel[catID] = values
	return nil
}

// validateSelections checks every recorded value against its category catalog.
func validateSelections(sel Selections) error {
	for catID, values := range sel {
		cat := categoryByID(catID)
		if cat == nil {
			return fmt.Errorf("unknown category %q", catID)
		}
		for _, v := range values {
			if !cat.isValidValue(v) {
				return fmt.Errorf("invalid value %q for --%s (allowed: %s)",
					v, cat.ID, strings.Join(allowedValueIDs(cat), ", "))
			}
		}
	}
	return nil
}

func allowedValueIDs(cat *Category) []string {
	out := make([]string, 0, len(cat.Values))
	for _, v := range cat.Values {
		out = append(out, v.ID)
	}
	return out
}
