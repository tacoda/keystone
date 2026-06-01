package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"golang.org/x/term"
)

// isTerminal reports whether f is a real TTY. Uses an ioctl rather than the
// character-device check, since /dev/null is a character device but not a TTY.
func isTerminal(f *os.File) bool {
	return term.IsTerminal(int(f.Fd()))
}

// promptMissing fills any unset categories in sel by running a huh form.
// Only categories absent from sel are prompted; others are left as-is.
// If skip[catID] is true, the category is skipped entirely (e.g. agent
// after successful detection).
func promptMissing(sel Selections, skip map[string]bool) error {
	// Collect categories that need a prompt, in catalog order.
	needs := []*Category{}
	for i := range categories {
		c := &categories[i]
		if skip[c.ID] {
			continue
		}
		if _, set := sel[c.ID]; set {
			continue
		}
		needs = append(needs, c)
	}
	if len(needs) == 0 {
		return nil
	}

	// Bound storage for each prompted category. We keep two parallel maps
	// (single-select strings and multi-select slices) because huh's generic
	// fields bind to typed pointers.
	singleVals := map[string]*string{}
	multiVals := map[string]*[]string{}

	fields := make([]huh.Field, 0, len(needs))
	for _, c := range needs {
		options := make([]huh.Option[string], 0, len(c.Values))
		for _, v := range c.Values {
			label := fmt.Sprintf("%s — %s", v.ID, v.Description)
			options = append(options, huh.NewOption(label, v.ID))
		}

		if c.MultiSelect {
			val := []string{}
			multiVals[c.ID] = &val
			fields = append(fields,
				huh.NewMultiSelect[string]().
					Title(c.Label).
					Description(c.Description).
					Options(options...).
					Value(&val),
			)
		} else {
			var val string
			singleVals[c.ID] = &val
			fields = append(fields,
				huh.NewSelect[string]().
					Title(c.Label).
					Description(c.Description).
					Options(options...).
					Value(&val),
			)
		}
	}

	form := huh.NewForm(huh.NewGroup(fields...))
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return fmt.Errorf("cancelled")
		}
		return fmt.Errorf("prompt: %w", err)
	}

	// Persist bound values back into sel.
	for id, v := range singleVals {
		if *v != "" {
			sel[id] = []string{*v}
		}
	}
	for id, v := range multiVals {
		if len(*v) > 0 {
			sel[id] = *v
		}
	}
	return nil
}
