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

// eqStr / eqSlice / nilSlice keep the field-assertion branching out of the
// parse tests so each test body stays flat.
func eqStr(t *testing.T, name, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %q, want %q", name, got, want)
	}
}

func wantSlice(t *testing.T, name string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%s = %v, want %v", name, got, want)
		return
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("%s = %v, want %v", name, got, want)
			return
		}
	}
}

func nilSlice(t *testing.T, name string, got []string) {
	t.Helper()
	if got != nil {
		t.Errorf("%s = %v, want nil", name, got)
	}
}

func mustParse(t *testing.T, in string) Frontmatter {
	t.Helper()
	fm, ok, err := Parse(in)
	if err != nil {
		t.Fatalf("Parse err: %v", err)
	}
	if !ok {
		t.Fatal("ok = false; want true")
	}
	return fm
}

func TestParse_RuleFull(t *testing.T) {
	fm := mustParse(t, `---
kind: guide
id: idioms/rails/migrations
description: Migrations are reversible and small.
severity: must
globs:
  - "db/migrate/**"
  - "!db/migrate/legacy/**"
corpus:
  - corpus/idioms/rails/migrations
supersedes:
  - idioms/rails/old-migrations
---
# body unchanged
`)
	eqStr(t, "Kind", fm.Kind, "guide")
	eqStr(t, "ID", fm.ID, "idioms/rails/migrations")
	eqStr(t, "Severity", fm.Severity, "must")
	wantSlice(t, "Globs", fm.Globs, []string{"db/migrate/**", "!db/migrate/legacy/**"})
	// `corpus:` is the renamed `traces:` — cite the reasoning behind this rule.
	wantSlice(t, "Corpus", fm.Corpus, []string{"corpus/idioms/rails/migrations"})
	wantSlice(t, "Supersedes", fm.Supersedes, []string{"idioms/rails/old-migrations"})
}

func TestParse_DocumentFields(t *testing.T) {
	fm := mustParse(t, `---
kind: document
id: implementation-plan
description: The central planning artifact.
type: feature
produced_by: orient
consumes:
  - explore-findings
produces:
  - review
gates:
  - draft
  - in-review
  - approved
  - executed
  - done
stop: Acceptance criteria are all checkable and the plan is approved.
---
# body
`)
	eqStr(t, "Kind", fm.Kind, "document")
	eqStr(t, "Type", fm.Type, "feature")
	eqStr(t, "ProducedBy", fm.ProducedBy, "orient")
	wantSlice(t, "Consumes", fm.Consumes, []string{"explore-findings"})
	wantSlice(t, "Produces", fm.Produces, []string{"review"})
	wantSlice(t, "Gates", fm.Gates, []string{"draft", "in-review", "approved", "executed", "done"})
	if fm.Stop == "" {
		t.Errorf("Stop = %q, want non-empty", fm.Stop)
	}
}

// TestParse_NoNewFields guards back-compat: a descriptor that declares none of
// the 3.0 fields parses clean with zero values.
func TestParse_NoNewFields(t *testing.T) {
	fm := mustParse(t, `---
kind: guide
id: process/spec
description: Capture intent first.
---
body
`)
	nilSlice(t, "Produces", fm.Produces)
	nilSlice(t, "Consumes", fm.Consumes)
	nilSlice(t, "Gates", fm.Gates)
	nilSlice(t, "Supersedes", fm.Supersedes)
	nilSlice(t, "Corpus", fm.Corpus)
	eqStr(t, "Stop", fm.Stop, "")
	eqStr(t, "ProducedBy", fm.ProducedBy, "")
	eqStr(t, "Type", fm.Type, "")
	eqStr(t, "Mode", fm.Mode, "")
	eqStr(t, "Event", fm.Event, "")
	eqStr(t, "Run", fm.Run, "")
	eqStr(t, "Agent", fm.Agent, "")
	eqStr(t, "Returns", fm.Returns, "")
}

// TestParse_HookFields — the 3.0 hook/sensor action fields decode:
// computational binds an event to a `run:` script; inferential dispatches an
// `agent:` that must emit a `returns:`-schema'd result.
func TestParse_HookFields(t *testing.T) {
	comp := mustParse(t, `---
kind: hook
id: gofmt-on-save
description: Format Go files after every edit.
mode: computational
event: PostToolUse
run: gofmt -w "$KEYSTONE_FILE"
---
body
`)
	eqStr(t, "Kind", comp.Kind, "hook")
	eqStr(t, "Mode", comp.Mode, "computational")
	eqStr(t, "Event", comp.Event, "PostToolUse")
	eqStr(t, "Run", comp.Run, `gofmt -w "$KEYSTONE_FILE"`)

	inf := mustParse(t, `---
kind: sensor
id: security-review
description: Review the diff for OWASP issues.
mode: inferential
agent: security-reviewer
returns: review-findings
---
body
`)
	eqStr(t, "Mode", inf.Mode, "inferential")
	eqStr(t, "Agent", inf.Agent, "security-reviewer")
	eqStr(t, "Returns", inf.Returns, "review-findings")
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
