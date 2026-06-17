package manifest

import (
	"fmt"
	"regexp"
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
