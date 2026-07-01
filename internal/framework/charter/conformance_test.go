package charter

import "testing"

func TestConformance_CleanProjectIsConformant(t *testing.T) {
	tmp := t.TempDir()
	// A well-formed guide (paired corpus) governing all .go; a corpus.
	write(t, tmp, ".charter/guides/idioms/go/x.md",
		"---\nkind: guide\nid: idioms/go/x\ndescription: d\nglobs:\n  - \"**/*.go\"\ncorpus:\n  - idioms/go/x\n---\nbody\n")
	write(t, tmp, ".charter/corpus/idioms/go/x.md",
		"---\nkind: corpus\nid: idioms/go/x\ndescription: d\n---\nwhy\n")
	write(t, tmp, "main.go", "package main\n")

	rub, err := Conformance(tmp, ".charter")
	if err != nil {
		t.Fatal(err)
	}
	if rub.Verdict != Conformant {
		t.Errorf("verdict = %s, want %s\n%+v", rub.Verdict, Conformant, rub.Criteria)
	}
	if len(rub.Criteria) != 4 {
		t.Errorf("expected 4 criteria, got %d", len(rub.Criteria))
	}
}

func TestConformance_UnpairedGuideDrifts(t *testing.T) {
	tmp := t.TempDir()
	// Inferential guide with NO corpus → pairing fails → verdict FAIL (0%).
	write(t, tmp, ".charter/guides/idioms/go/x.md",
		"---\nkind: guide\nid: idioms/go/x\ndescription: d\nglobs:\n  - \"**/*.go\"\n---\nbody\n")
	write(t, tmp, "main.go", "package main\n")

	rub, err := Conformance(tmp, ".charter")
	if err != nil {
		t.Fatal(err)
	}
	// pairing 0% → FAIL → NON-CONFORMANT.
	if rub.Verdict != NonConformant {
		t.Errorf("verdict = %s, want %s", rub.Verdict, NonConformant)
	}
	var pairing *Criterion
	for i := range rub.Criteria {
		if rub.Criteria[i].Name == "Guide↔corpus pairing" {
			pairing = &rub.Criteria[i]
		}
	}
	if pairing == nil || pairing.Status != statusFail {
		t.Errorf("pairing criterion = %+v, want FAIL", pairing)
	}
}
