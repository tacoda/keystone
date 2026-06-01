package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// copyMode controls behavior when a destination file already exists.
type copyMode int

const (
	overwrite     copyMode = iota // replace existing files
	skipIfExists                  // leave existing files alone; warn
)

// copyTree copies every regular file under srcRoot in the embedded FS to
// destDir on disk. Paths under srcRoot are joined onto destDir as-is
// (the srcRoot prefix itself is not included).
func copyTree(srcFS fs.FS, srcRoot, destDir string, mode copyMode) error {
	return fs.WalkDir(srcFS, srcRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == srcRoot {
			return nil
		}

		rel, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return err
		}
		dest := filepath.Join(destDir, rel)

		if d.IsDir() {
			return os.MkdirAll(dest, 0o755)
		}

		if mode == skipIfExists {
			if _, statErr := os.Stat(dest); statErr == nil {
				fmt.Fprintf(os.Stderr, "  exists: %s (skipped — review and merge manually)\n", dest)
				return nil
			} else if !os.IsNotExist(statErr) {
				return statErr
			}
		}

		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}

		src, err := srcFS.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()

		out, err := os.Create(dest)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, src); err != nil {
			out.Close()
			return err
		}
		if err := out.Close(); err != nil {
			return err
		}

		fmt.Fprintf(os.Stdout, "  wrote: %s\n", dest)
		return nil
	})
}
