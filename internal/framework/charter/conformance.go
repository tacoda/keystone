package charter

import (
	"fmt"
	"strings"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/loader"
	"github.com/tacoda/keystone/internal/framework/lockfile"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// Conformance verdicts.
const (
	Conformant    = "CONFORMANT"     // every criterion PASSes
	Drifting      = "DRIFTING"       // at least one PARTIAL, no FAIL
	NonConformant = "NON-CONFORMANT" // at least one FAIL
)

// Criterion status values.
const (
	statusPass    = "PASS"
	statusPartial = "PARTIAL"
	statusFail    = "FAIL"
)

// Criterion is one rubric line: an objective, explainable check.
type Criterion struct {
	Name   string
	Rule   string
	Status string // PASS | PARTIAL | FAIL
	Detail string // the measured value / what fell short
}

// Rubric is the charter-conformance scorecard: a set of criteria plus
// the overall verdict derived from them. It is deliberately a rubric,
// not a single number — every line is a real check the reader can act on.
type Rubric struct {
	Criteria []Criterion
	Verdict  string
}

// Conformance evaluates the rubric for a project's charter.
func Conformance(projectDir, charterRoot string) (Rubric, error) {
	prims, _, err := primitive.Walk(projectDir, charterRoot)
	if err != nil {
		return Rubric{}, err
	}
	criteria := []Criterion{
		cascadeCriterion(projectDir, charterRoot),
		validityCriterion(prims),
		pairingCriterion(prims),
		coverageCriterion(projectDir, charterRoot),
	}
	return Rubric{Criteria: criteria, Verdict: verdict(criteria)}, nil
}

// verdict rolls the criteria up: any FAIL → NON-CONFORMANT; else any
// PARTIAL → DRIFTING; else CONFORMANT.
func verdict(cs []Criterion) string {
	v := Conformant
	for _, c := range cs {
		switch c.Status {
		case statusFail:
			return NonConformant
		case statusPartial:
			v = Drifting
		}
	}
	return v
}

// cascadeCriterion: no strict-policy violations (the hard boundary a
// project must honor). PASS when verify is clean or there are no
// policies; FAIL on any strict/depth violation.
func cascadeCriterion(projectDir, charterRoot string) Criterion {
	c := Criterion{Name: "Cascade integrity", Rule: "no strict-policy violations"}
	cfg, err := config.ReadProjectConfig(projectDir)
	if err != nil {
		c.Status, c.Detail = statusPass, "no policies — nothing to violate"
		return c
	}
	if cfg == nil {
		c.Status, c.Detail = statusPass, "no policies — nothing to violate"
		return c
	}
	if len(cfg.Policies) == 0 {
		c.Status, c.Detail = statusPass, "no policies — nothing to violate"
		return c
	}
	expected := map[string]map[string]string{}
	if lf, lerr := lockfile.Read(projectDir, charterRoot); lerr == nil {
		for name, lock := range lf.Policies {
			expected[name] = lock.Files
		}
	}
	res, verr := loader.Verify(projectDir, cfg, expected)
	if verr != nil {
		c.Status, c.Detail = statusFail, "verify error: "+verr.Error()
		return c
	}
	if res.HasErrors() {
		c.Status, c.Detail = statusFail, fmt.Sprintf("%d strict/depth violation(s)", len(res.Violations))
		return c
	}
	c.Status, c.Detail = statusPass, "clean"
	return c
}

// validityCriterion: every primitive lints clean. PASS at 0 errors,
// PARTIAL under 5% erroring, FAIL at/above.
func validityCriterion(prims []primitive.Primitive) Criterion {
	c := Criterion{Name: "Frontmatter validity", Rule: "primitives lint clean (PASS 0 · PARTIAL <5% · FAIL ≥5%)"}
	errs := 0
	for _, f := range primitive.Lint(prims) {
		if f.Severity == primitive.FindingError {
			errs++
		}
	}
	total := len(prims)
	c.Detail = fmt.Sprintf("%d lint error(s) across %d primitive(s)", errs, total)
	switch {
	case errs == 0:
		c.Status = statusPass
	case total > 0 && float64(errs)/float64(total) < 0.05:
		c.Status = statusPartial
	default:
		c.Status = statusFail
	}
	return c
}

// pairingCriterion: idiom guides have a paired corpus (the charter's
// golden rule — `guides/idioms/<X>` pairs with `corpus/<X>`, or an
// explicit `corpus:` link). Scoped to idioms: process/principle/domain
// guides are self-contained directives, and computational guides (LSP)
// and README index files are exempt.
func pairingCriterion(prims []primitive.Primitive) Criterion {
	c := Criterion{Name: "Guide↔corpus pairing", Rule: "idiom guides paired with corpus (PASS ≥90% · PARTIAL ≥60%)"}
	corpusTopics := corpusTopicSet(prims)
	total, paired := 0, 0
	for _, p := range prims {
		if !needsCorpusPairing(p) {
			continue
		}
		total++
		if isPaired(p, corpusTopics) {
			paired++
		}
	}
	c.Status, c.Detail = bandRatio(paired, total, 0.90, 0.60)
	return c
}

// corpusTopicSet indexes corpus primitives by topic (id minus the
// `corpus/` prefix) for pairing lookups.
func corpusTopicSet(prims []primitive.Primitive) map[string]bool {
	set := map[string]bool{}
	for _, p := range prims {
		if primitive.Kind(p.Kind) == primitive.KindCorpus {
			set[strings.TrimPrefix(p.ID, "corpus/")] = true
		}
	}
	return set
}

// isPaired reports whether an idiom guide has a matching corpus — a
// parallel topic or an explicit `corpus:` link.
func isPaired(p primitive.Primitive, corpusTopics map[string]bool) bool {
	if len(p.Corpus) > 0 {
		return true
	}
	return corpusTopics[strings.TrimPrefix(p.ID, "guides/")]
}

// needsCorpusPairing reports whether a primitive is an idiom guide that
// the golden rule expects to pair with a corpus entry.
func needsCorpusPairing(p primitive.Primitive) bool {
	if primitive.Kind(p.Kind) != primitive.KindGuide || p.Mode == "computational" {
		return false
	}
	if strings.HasSuffix(p.ID, "/README") {
		return false
	}
	return strings.Contains(p.ID, "idioms/")
}

// coverageCriterion: source files governed by a guide.
func coverageCriterion(projectDir, charterRoot string) Criterion {
	c := Criterion{Name: "Coverage", Rule: "source files governed (PASS ≥80% · PARTIAL ≥50%)"}
	res, err := Coverage(projectDir, charterRoot)
	if err != nil {
		c.Status, c.Detail = statusFail, "coverage error: "+err.Error()
		return c
	}
	c.Status, c.Detail = bandRatio(res.Governed, res.Total, 0.80, 0.50)
	return c
}

// bandRatio scores num/den against pass/partial thresholds and returns
// the status plus a "N% (num/den)" detail. An empty set is a PASS
// (nothing to fall short of).
func bandRatio(num, den int, passAt, partialAt float64) (status, detail string) {
	if den == 0 {
		return statusPass, "n/a (0)"
	}
	ratio := float64(num) / float64(den)
	detail = fmt.Sprintf("%d%% (%d/%d)", int(ratio*100), num, den)
	switch {
	case ratio >= passAt:
		return statusPass, detail
	case ratio >= partialAt:
		return statusPartial, detail
	default:
		return statusFail, detail
	}
}
