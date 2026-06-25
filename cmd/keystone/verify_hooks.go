package main

import "strings"

// verifyWithHooks wraps `keystone verify` with the pre/post-verify framework
// gate. A failing pre-verify hook blocks; post-verify is a reaction. `--sensor`
// mode (a single keystone-owned check, itself used as a host hook) does not
// fire the lifecycle hooks. Kept out of verify.go so that command stays
// untouched. The flag scan is inlined (one pass) to keep the surface tiny.
func verifyWithHooks(args []string) error {
	sensorMode, dir := verifyFlags(args)
	if sensorMode {
		return runVerify(args)
	}
	if err := runHookFire([]string{"pre-verify", "--dir", dir}); err != nil {
		return err
	}
	if err := runVerify(args); err != nil {
		return err
	}
	return runHookFire([]string{"post-verify", "--dir", dir})
}

// verifyFlags scans the verify args for the two flags the gate cares about:
// whether --sensor mode is on, and the --dir value (default ".").
func verifyFlags(args []string) (sensorMode bool, dir string) {
	dir = "."
	for i, a := range args {
		switch {
		case a == "--sensor", strings.HasPrefix(a, "--sensor="):
			sensorMode = true
		case a == "--dir":
			if i+1 < len(args) {
				dir = args[i+1]
			}
		case strings.HasPrefix(a, "--dir="):
			dir = strings.TrimPrefix(a, "--dir=")
		}
	}
	return sensorMode, dir
}
