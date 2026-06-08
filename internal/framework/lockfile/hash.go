package lockfile

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
)

// HashFile returns "sha256:<hex>" for the file at path.
func HashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}

// HashFilesUnder walks dir (relative to installDir) and returns a map of
// path-relative-to-installDir → hash for every regular file found. Used after
// a policy install to seed the lockfile, and on update to detect dirty files.
func HashFilesUnder(installDir, dir string) (map[string]string, error) {
	result := map[string]string{}
	walkRoot := filepath.Join(installDir, dir)
	err := filepath.Walk(walkRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(installDir, path)
		if relErr != nil {
			return relErr
		}
		h, herr := HashFile(path)
		if herr != nil {
			return herr
		}
		result[filepath.ToSlash(rel)] = h
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
