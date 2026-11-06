package main

import (
	"fmt"
	"strings"
)

// parseFlags reads `--name value`, `--name=value`, and boolean `--name`
// flags into the given targets. `--help`/`-h` calls usage and returns
// help=true. Positional arguments and unknown flags are errors.
func parseFlags(args []string, values map[string]*string, bools map[string]*bool, usage func()) (help bool, err error) {
	for i := 0; i < len(args); i++ {
		if args[i] == "--help" || args[i] == "-h" {
			usage()
			return true, nil
		}
		if i, err = parseOneFlag(args, i, values, bools); err != nil {
			return false, err
		}
	}
	return false, nil
}

// parseOneFlag consumes the flag at args[i] (and its value, if separate)
// and returns the index of the last argument it used.
func parseOneFlag(args []string, i int, values map[string]*string, bools map[string]*bool) (int, error) {
	a := args[i]
	if !strings.HasPrefix(a, "-") {
		return i, fmt.Errorf("unexpected positional argument %q", a)
	}
	name, val, hasVal := strings.Cut(a, "=")
	if b, ok := bools[name]; ok && !hasVal {
		*b = true
		return i, nil
	}
	v, ok := values[name]
	switch {
	case !ok:
		return i, fmt.Errorf("unknown flag %s", a)
	case hasVal:
		*v = val
		return i, nil
	case i+1 >= len(args):
		return i, fmt.Errorf("flag %s requires a value", a)
	}
	*v = args[i+1]
	return i + 1, nil
}
