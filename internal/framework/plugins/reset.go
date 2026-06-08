package plugins

import (
	"fmt"
	"os"
	"runtime"
	"path/filepath"
)

// Reset removes the installed plugin directory at
// <projectDir>/<harnessRoot>/plugins/<name>/. Called by the drift-reset
// path after Verify reports drift, and by `keystone plugin remove`.
//
// On POSIX the directory's contents are chmodded back to writable before
// removal so a previous Install's read-only marks don't block rm.
func Reset(name, projectDir, harnessRoot string) error {
	if name == "" {
		return fmt.Errorf("plugins.Reset: empty name")
	}
	target := pluginDir(projectDir, harnessRoot, name)

	if runtime.GOOS != "windows" {
		_ = makeWritable(target)
	}
	if err := os.RemoveAll(target); err != nil {
		return fmt.Errorf("remove %s: %w", target, err)
	}
	return nil
}

// makeWritable walks target and chmods every file 0o644 + every directory
// 0o755 so RemoveAll can succeed after Install's 0444 read-only pass.
// Best-effort: errors are returned but the caller will still attempt
// RemoveAll which has its own error surface.
func makeWritable(target string) error {
	return filepath.Walk(target, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if info.IsDir() {
			return os.Chmod(path, 0o755)
		}
		return os.Chmod(path, 0o644)
	})
}
