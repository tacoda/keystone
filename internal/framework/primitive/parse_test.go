package primitive

import (
	"testing"
)

func TestSplitFrontmatter(t *testing.T) {
	cases := []struct {
		name     string
		in       string
		wantFM   string
		wantBody string
		wantOK   bool
	}{
		{
			name: "valid",
			in: "---\nkind: rule\n---\nbody\n",
			wantFM:   "kind: rule\n",
			wantBody: "body\n",
			wantOK:   true,
		},
		{
			name:     "no frontmatter",
			in:       "# heading\nbody\n",
			wantFM:   "",
			wantBody: "# heading\nbody\n",
			wantOK:   false,
		},
		{
			name:     "unterminated fence",
			in:       "---\nkind: rule\nbody\n",
			wantFM:   "",
			wantBody: "---\nkind: rule\nbody\n",
			wantOK:   false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fm, body, ok := SplitFrontmatter(c.in)
			if ok != c.wantOK {
				t.Fatalf("ok = %v, want %v", ok, c.wantOK)
			}
			if fm != c.wantFM {
				t.Errorf("fm = %q, want %q", fm, c.wantFM)
			}
			if body != c.wantBody {
				t.Errorf("body = %q, want %q", body, c.wantBody)
			}
		})
	}
}

func TestParse_GuideFull(t *testing.T) {
	in := `---
kind: guide
id: idioms/rails/migrations
description: Migrations are reversible and small.
severity: must
tier: iron
globs:
  - "db/migrate/**"
  - "!db/migrate/legacy/**"
traces:
  - corpus/idioms/rails/migrations
---
# body unchanged
`
	fm, ok, err := Parse(in)
	if err != nil {
		t.Fatalf("Parse err: %v", err)
	}
	if !ok {
		t.Fatal("ok = false; want true")
	}
	if fm.Kind != "guide" {
		t.Errorf("Kind = %q", fm.Kind)
	}
	if fm.ID != "idioms/rails/migrations" {
		t.Errorf("ID = %q", fm.ID)
	}
	if fm.Severity != "must" {
		t.Errorf("Severity = %q", fm.Severity)
	}
	if len(fm.Globs) != 2 || fm.Globs[0] != "db/migrate/**" {
		t.Errorf("Globs = %v", fm.Globs)
	}
	if len(fm.Traces) != 1 || fm.Traces[0] != "corpus/idioms/rails/migrations" {
		t.Errorf("Traces = %v", fm.Traces)
	}
}

func TestParse_ArgsStringForm(t *testing.T) {
	in := `---
kind: command
id: review
description: Review PR.
args:
  - target
  - depth
---
body
`
	fm, _, err := Parse(in)
	if err != nil {
		t.Fatalf("Parse err: %v", err)
	}
	if len(fm.Args) != 2 {
		t.Fatalf("Args len = %d, want 2", len(fm.Args))
	}
	if fm.Args[0].Name != "target" || fm.Args[1].Name != "depth" {
		t.Errorf("Args = %+v", fm.Args)
	}
}

func TestParse_ArgsMappingForm(t *testing.T) {
	in := `---
kind: command
id: review
description: Review PR.
args:
  - name: target
    type: string
    required: true
    description: PR ref
---
body
`
	fm, _, err := Parse(in)
	if err != nil {
		t.Fatalf("Parse err: %v", err)
	}
	if len(fm.Args) != 1 {
		t.Fatalf("Args len = %d, want 1", len(fm.Args))
	}
	a := fm.Args[0]
	if a.Name != "target" || a.Type != "string" || !a.Required || a.Description != "PR ref" {
		t.Errorf("Args[0] = %+v", a)
	}
}

func TestParse_NoFrontmatter(t *testing.T) {
	_, ok, err := Parse("# heading only\n")
	if err != nil {
		t.Fatalf("Parse err: %v", err)
	}
	if ok {
		t.Error("ok = true; want false for body-only file")
	}
}
