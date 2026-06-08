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

// validate enforces required fields, name format, and tier values.
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
	switch m.Tier {
	case "", TierOrg, TierTeam:
	default:
		return fmt.Errorf("%s: tier %q must be %q or %q", PolicyManifestFile, m.Tier, TierOrg, TierTeam)
	}
	if m.ResolvedTier() == TierOrg {
		if len(m.Strict.Sensors) > 0 {
			return fmt.Errorf("%s: org-tier policies cannot declare strict sensors (sensors cascade is team → project only)", PolicyManifestFile)
		}
		if len(m.Required.Sensors) > 0 {
			return fmt.Errorf("%s: org-tier policies cannot declare required sensors (sensors cascade is team → project only)", PolicyManifestFile)
		}
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
	sensorsPrefix := filepath.Join(allowedPrefix, "sensors")
	isOrg := m.ResolvedTier() == TierOrg
	var files []string
	var stray []string
	var sensorFilesInOrg []string

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
		if isOrg && strings.HasPrefix(rel, sensorsPrefix+string(filepath.Separator)) {
			sensorFilesInOrg = append(sensorFilesInOrg, rel)
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
	if len(sensorFilesInOrg) > 0 {
		return nil, fmt.Errorf(
			"policy %q is tier %q but ships sensor files (sensors cascade is team → project only):\n  %s",
			m.Name, TierOrg, strings.Join(sensorFilesInOrg, "\n  "),
		)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("policy %q contains no files under %s/", m.Name, allowedPrefix)
	}
	return files, nil
}
