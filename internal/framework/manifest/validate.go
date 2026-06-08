package manifest

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var namePattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,63}$`)

// validate enforces required fields and name format.
func (m *Manifest) validate() error {
	if m.Name == "" {
		return fmt.Errorf("%s: missing required field 'name'", PolicyManifestFile)
	}
	if !namePattern.MatchString(m.Name) {
		return fmt.Errorf("%s: name %q must match %s", PolicyManifestFile, m.Name, namePattern)
	}
	if m.Version == "" {
		return fmt.Errorf("%s: missing required field 'version'", PolicyManifestFile)
	}
	return nil
}

// ValidateContent walks policyRoot/policy/ and ensures every regular file
// lives under policy/harness/policies/<namespace>/. Returns the list of
// relative file paths (within policyRoot) that will be copied, or an error
// if any file is outside the allowed namespace.
func ValidateContent(policyRoot string, m *Manifest) ([]string, error) {
	contentRoot := filepath.Join(policyRoot, PolicyContentRoot)
	info, err := os.Stat(contentRoot)
	if err != nil {
		return nil, fmt.Errorf("policy must contain a %s/ directory: %w", PolicyContentRoot, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", PolicyContentRoot)
	}

	allowedPrefix := filepath.Join(PolicyContentRoot, "harness", "policies", m.Namespace())
	var files []string
	var stray []string

	err = filepath.WalkDir(contentRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(policyRoot, path)
		if relErr != nil {
			return relErr
		}
		if !strings.HasPrefix(rel, allowedPrefix+string(filepath.Separator)) {
			stray = append(stray, rel)
			return nil
		}
		files = append(files, rel)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(stray) > 0 {
		return nil, fmt.Errorf(
			"policy %q writes files outside its allowed namespace %s/:\n  %s",
			m.Name, allowedPrefix, strings.Join(stray, "\n  "),
		)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("policy %q contains no files under %s/", m.Name, allowedPrefix)
	}
	return files, nil
}
