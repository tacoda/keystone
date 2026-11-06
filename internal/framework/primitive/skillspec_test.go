package primitive

import (
	"strings"
	"testing"
)

func skillAt(dir string, fm Frontmatter) Primitive {
	fm.Kind = "skill"
	if fm.ID == "" {
		fm.ID = dir
	}
	if fm.Description == "" {
		fm.Description = "d"
	}
	return Primitive{Frontmatter: fm, Path: ".agents/skills/" + dir + "/SKILL.md"}
}

func TestLintSkillSpec(t *testing.T) {
	cases := []struct {
		name string
		p    Primitive
		sev  FindingSeverity
		want string // "" = no finding
	}{
		{"spec clean", skillAt("deploy", Frontmatter{Name: "deploy", Keys: []string{"name", "description"}}), "", ""},
		{"missing name", skillAt("deploy", Frontmatter{Keys: []string{"description"}}), FindingWarning, "add `name: deploy`"},
		{"bad charset", skillAt("deploy", Frontmatter{Name: "Deploy_It", Keys: []string{"name"}}), FindingError, "invalid"},
		{"double hyphen", skillAt("a--b", Frontmatter{Name: "a--b", Keys: []string{"name"}}), FindingError, "invalid"},
		{"dir mismatch", skillAt("deploy", Frontmatter{Name: "ship", Keys: []string{"name"}}), FindingError, "must match its directory"},
		{"long description", skillAt("deploy", Frontmatter{Name: "deploy", Description: strings.Repeat("x", 1025), Keys: []string{"name"}}), FindingError, "1024"},
		{"4.x keys", skillAt("deploy", Frontmatter{Name: "deploy", Keys: []string{"name", "triggers", "kind"}}), FindingWarning, "kind, triggers"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var got []Finding
			lintSkillSpec(c.p, func(sev FindingSeverity, msg string) {
				got = append(got, Finding{Severity: sev, Message: msg})
			})
			if c.want == "" {
				if len(got) != 0 {
					t.Errorf("expected no findings, got %v", got)
				}
				return
			}
			if !find(t, got, c.sev, c.want) {
				t.Errorf("expected %s containing %q, got %v", c.sev, c.want, got)
			}
		})
	}
}

func TestSpecSkillFrontmatter(t *testing.T) {
	p := skillAt("keystone-verify", Frontmatter{
		ID:          "keystone:verify",
		Description: "Check the cascade.",
		Triggers:    []string{"keystone verify", "audit policy drift"},
		Tools:       []string{"Read", "Bash"},
		Model:       "opus",
		Tags:        []string{"audit", "cascade"},
		Keys:        []string{"kind", "id", "description", "triggers", "model", "tools", "tags"},
	})
	got, err := SpecSkillFrontmatter(p)
	if err != nil {
		t.Fatal(err)
	}
	want := `name: keystone-verify
description: 'Check the cascade. Triggers: keystone verify; audit policy drift.'
allowed-tools: Read Bash
metadata:
  model: opus
  tags: audit,cascade
`
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
	// Round trip: the rewritten frontmatter parses back to the same tools,
	// model, and tags, with ID taken from name.
	fm, _, err := Parse("---\n" + got + "---\nbody\n")
	if err != nil {
		t.Fatal(err)
	}
	roundTrip := fm.ID + "|" + strings.Join(fm.Tools, " ") + "|" + fm.Model + "|" + strings.Join(fm.Tags, ",")
	if roundTrip != "keystone-verify|Read Bash|opus|audit,cascade" {
		t.Errorf("round trip lost data: %s", roundTrip)
	}
}

func TestSpecSkillFrontmatter_OverlongTriggersGoToMetadata(t *testing.T) {
	p := skillAt("big", Frontmatter{
		Description: strings.Repeat("d", 1000),
		Triggers:    []string{strings.Repeat("t", 100)},
		Keys:        []string{"description", "triggers"},
	})
	got, err := SpecSkillFrontmatter(p)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "Triggers:") || !strings.Contains(got, "triggers: "+strings.Repeat("t", 100)) {
		t.Errorf("overlong triggers should move to metadata, got:\n%s", got)
	}
}

func TestSpecSkillFrontmatter_RefusesUncarriedKeys(t *testing.T) {
	p := skillAt("x", Frontmatter{Keys: []string{"description", "phase", "args"}})
	if _, err := SpecSkillFrontmatter(p); err == nil || !strings.Contains(err.Error(), "args, phase") {
		t.Errorf("expected refusal naming args, phase; got %v", err)
	}
}

func TestInlineIncludes(t *testing.T) {
	raw := "kind: agent\nid: code-reviewer\ndescription: d\ntools:\n  - Read\nincludes:\n  - reads-diff\ntags:\n  - review\nmodel: opus\n"
	composed := Primitive{Frontmatter: Frontmatter{Tools: []string{"Bash", "Read"}, Tags: []string{"diff", "review"}}}
	got, err := InlineIncludes(raw, composed)
	if err != nil {
		t.Fatal(err)
	}
	want := "kind: agent\nid: code-reviewer\ndescription: d\ntools:\n  - Bash\n  - Read\ntags:\n  - diff\n  - review\nmodel: opus\n"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}
