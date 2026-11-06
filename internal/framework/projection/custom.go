package projection

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

// CustomHostsDir holds the hosts a project defines itself, one JSON file
// per host (`keystone new adapter <name>`).
const CustomHostsDir = ".agents/hosts"

// customHost is a host defined in .agents/hosts/<name>.json: for each
// projectable kind, how the host gets it. It has no format conversion, no
// hook bridge, and no permission translation; `keystone project` reports
// those as gaps.
type customHost struct {
	Name  string                `json:"name"`
	Kinds map[string]customKind `json:"kinds"`
}

type customKind struct {
	Strategy string `json:"strategy"`      // native | link | copy | index
	Dir      string `json:"dir,omitempty"` // host folder, for link and copy
}

// customKinds are the kinds a custom host can receive; any it leaves out
// are indexed in AGENTS.md.
var customKinds = []primitive.Kind{skillK, ruleK, commandK, subagentK}

var customStrategies = map[string]Strategy{"native": Native, "link": Symlink, "copy": Copy, "index": Index}

var hostNameRE = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

// HostsFor is every host keystone projects to: the built-in hosts in
// adapters (claude-code always) plus each custom host in adapters.
func HostsFor(projectDir string, adapters []string) ([]Profile, error) {
	hosts := Enabled(adapters)
	custom, err := LoadCustomHosts(projectDir)
	if err != nil {
		return hosts, err
	}
	want := map[string]bool{}
	for _, a := range adapters {
		want[a] = true
	}
	for _, c := range custom {
		if want[c.Host] {
			hosts = append(hosts, c)
		}
	}
	return hosts, nil
}

// LoadCustomHosts reads every .agents/hosts/<name>.json, in name order. A
// missing folder is no custom hosts; an invalid file is an error naming it.
func LoadCustomHosts(projectDir string) ([]Profile, error) {
	files, err := filepath.Glob(filepath.Join(projectDir, CustomHostsDir, "*.json"))
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	var out []Profile
	for _, f := range files {
		p, err := loadCustomHost(f)
		if err != nil {
			return nil, fmt.Errorf("%s/%s: %w", CustomHostsDir, filepath.Base(f), err)
		}
		out = append(out, p)
	}
	return out, nil
}

func loadCustomHost(file string) (Profile, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return Profile{}, err
	}
	var c customHost
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&c); err != nil {
		return Profile{}, err
	}
	if want := strings.TrimSuffix(filepath.Base(file), ".json"); c.Name != want {
		return Profile{}, fmt.Errorf("name %q must match the file name %q", c.Name, want)
	}
	return customProfile(c)
}

// customProfile validates c and builds its Profile.
func customProfile(c customHost) (Profile, error) {
	if err := checkHostName(c.Name); err != nil {
		return Profile{}, err
	}
	if err := checkKinds(c.Kinds); err != nil {
		return Profile{}, err
	}
	p := Profile{Host: c.Name, Rules: map[primitive.Kind]Rule{}}
	for _, k := range customKinds {
		r, err := customRule(k, c.Kinds[string(k)])
		if err != nil {
			return Profile{}, fmt.Errorf("kind %s: %w", k, err)
		}
		p.Rules[k] = r
	}
	return p, nil
}

// CustomHostSkeleton is a new host file for name: every kind indexed in
// AGENTS.md until the author picks a strategy.
func CustomHostSkeleton(name string) ([]byte, error) {
	if err := checkHostName(name); err != nil {
		return nil, err
	}
	c := customHost{Name: name, Kinds: map[string]customKind{}}
	for _, k := range customKinds {
		c.Kinds[string(k)] = customKind{Strategy: "index"}
	}
	data, err := json.MarshalIndent(c, "", "  ")
	return append(data, '\n'), err
}

// IsCustomHost reports whether projectDir defines host name.
func IsCustomHost(projectDir, name string) bool {
	_, err := os.Stat(filepath.Join(projectDir, CustomHostsDir, name+".json"))
	return err == nil
}

// checkHostName: a slug that no built-in host already uses.
func checkHostName(name string) error {
	if !hostNameRE.MatchString(name) {
		return fmt.Errorf("name %q: use lowercase letters, digits, and hyphens", name)
	}
	for _, mk := range profiles {
		if mk().Host == name {
			return fmt.Errorf("name %q is a built-in host", name)
		}
	}
	return nil
}

func checkKinds(kinds map[string]customKind) error {
	for kind := range kinds {
		if !isCustomKind(kind) {
			return fmt.Errorf("kind %q: a custom host takes skill, rule, command, or subagent", kind)
		}
	}
	return nil
}

func isCustomKind(kind string) bool {
	for _, k := range customKinds {
		if string(k) == kind {
			return true
		}
	}
	return false
}

// customRule is the projection rule for one kind. Left out, a kind is
// indexed.
func customRule(k primitive.Kind, ck customKind) (Rule, error) {
	if ck.Strategy == "" {
		ck.Strategy = "index"
	}
	s, ok := customStrategies[ck.Strategy]
	if !ok {
		return Rule{}, fmt.Errorf("strategy %q: use native, link, copy, or index", ck.Strategy)
	}
	if s != Symlink && s != Copy {
		return Rule{Strategy: s}, nil
	}
	if err := checkHostDir(ck.Dir); err != nil {
		return Rule{}, err
	}
	return mirrorRule(k, s, ck.Dir), nil
}

// checkHostDir: a folder inside the repo, outside keystone's own trees
// (keystone prunes what it writes there).
func checkHostDir(dir string) error {
	clean := path.Clean(filepath.ToSlash(dir))
	switch {
	case dir == "":
		return fmt.Errorf("link and copy need a dir")
	case !filepath.IsLocal(dir):
		return fmt.Errorf("dir %q must be a relative path inside the repo", dir)
	case clean == ".agents" || clean == ".keystone" || strings.HasPrefix(clean, ".agents/") || strings.HasPrefix(clean, ".keystone/"):
		return fmt.Errorf("dir %q must be outside .agents/ and .keystone/", dir)
	}
	return nil
}

// mirrorRule links or copies each primitive into dir under its host name;
// a skill mirrors its whole folder, an always-on or computational rule
// stays in AGENTS.md.
func mirrorRule(k primitive.Kind, s Strategy, dir string) Rule {
	switch k {
	case skillK:
		return Rule{Strategy: s, Dest: skillDirDest(dir), Target: skillDir}
	case ruleK:
		return Rule{Strategy: s, Dest: pathRuleDest(dir, ".md")}
	}
	return Rule{Strategy: s, Dest: nameDest(dir, ".md")}
}
