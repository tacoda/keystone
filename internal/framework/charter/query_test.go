package charter

import (
	"strings"
	"testing"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

func qprims() []primitive.Primitive {
	fm := primitive.Frontmatter{}
	return []primitive.Primitive{
		{Frontmatter: mkFM(fm, "rule", "process/always", "Always applies.")},
		{Frontmatter: mkPathedFM(fm, "corpus", "billing/props", "How props are billed.", []string{"**/*.go"})},
		{Frontmatter: mkCitedFM(fm, "rule", "go/opts", "Go options.", []string{"**/*.go"}, []string{"go/errors"})},
		{Frontmatter: mkFM(fm, "reference", "go/errors", "Why errors wrap.")},
		{Frontmatter: mkFM(fm, "hook", "build", "Builds the project.")},
		{Frontmatter: mkFM(fm, "skill", "deploy-app", "Deploys the app to production.")},
	}
}

func mkFM(_ primitive.Frontmatter, kind, id, desc string) primitive.Frontmatter {
	return primitive.Frontmatter{Kind: kind, ID: id, Description: desc}
}

func mkPathedFM(_ primitive.Frontmatter, kind, id, desc string, paths []string) primitive.Frontmatter {
	return primitive.Frontmatter{Kind: kind, ID: id, Description: desc, Paths: paths}
}

func mkCitedFM(_ primitive.Frontmatter, kind, id, desc string, paths, corpus []string) primitive.Frontmatter {
	return primitive.Frontmatter{Kind: kind, ID: id, Description: desc, Paths: paths, Corpus: corpus}
}

func TestQuery_EverywhereRulesOnly(t *testing.T) {
	res := Query(qprims(), []string{"internal/x.go"}, "")
	// Everywhere = rules without paths only: process/always. Hooks,
	// skills, uncited references, and no-path knowledge stay out.
	counts := map[string]int{}
	for _, e := range res.Entries {
		counts[e.Kind]++
		if e.Why != "everywhere" && !strings.HasPrefix(e.Why, "cited") && !strings.HasPrefix(e.Why, "paths:") {
			t.Errorf("unexpected entry %s/%s why=%q", e.Kind, e.ID, e.Why)
		}
	}
	if counts["rule"] != 2 { // always + go/wrap (path-matched, cited)
		t.Errorf("rules = %d, want 2 (everywhere + path-matched): %+v", counts["rule"], res.Entries)
	}
	if counts["hook"] != 0 || counts["skill"] != 0 {
		t.Errorf("invocable kinds must not flood query: %+v", res.Entries)
	}
	if counts["reference"] != 1 { // cited by the path-matched rule
		t.Errorf("cited reference missing: %+v", res.Entries)
	}
	if counts["corpus"] != 1 { // billing/props matches **/*.go
		t.Errorf("path-matched corpus missing: %+v", res.Entries)
	}
}

func TestQuery_TaskScoresKeywordOverlap(t *testing.T) {
	res := Query(qprims(), nil, "deploy the app")
	if len(res.Entries) == 0 {
		t.Fatal("task query found nothing")
	}
	if res.Entries[0].ID != "deploy-app" {
		t.Errorf("top task hit = %s, want deploy-app", res.Entries[0].ID)
	}
}

func TestQuery_Deduplicates(t *testing.T) {
	// go/opts is both everywhere-adjacent? No: it has paths, matched by
	// file, and cited — the same (kind,id) must appear once.
	res := Query(qprims(), []string{"main.go"}, "options")
	seen := map[string]int{}
	for _, e := range res.Entries {
		seen[e.Kind+"/"+e.ID]++
	}
	for k, n := range seen {
		if n > 1 {
			t.Errorf("%s appeared %d times", k, n)
		}
	}
}
