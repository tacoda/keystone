package main

import (
	"testing"
)

func TestAudit_LearningAndPruningSides(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		// A rule that is cited by another rule, has paths:, and triggers:
		// not a pruning candidate.
		".agents/rules/process/used.md": `---
kind: rule
id: process/used
description: used rule
paths:
  - "cmd/**"
triggers:
  - keystone:verify
corpus:
  - reference/process/backing
---
`,
		".agents/references/process/backing.md": `---
kind: reference
id: process/backing
description: the why
---
`,
		// Uncited, unscoped, untriggered: a pruning candidate.
		".agents/rules/process/dead.md": `---
kind: rule
id: process/dead
description: dead rule
---
`,
		// Cited by the hook below but unscoped and untriggered: only two
		// of three reasons — not a candidate (citation is a liveness
		// signal on its own).
		".agents/rules/process/cited.md": `---
kind: rule
id: process/cited
description: cited but unscoped
---
`,
		".agents/hooks/watch.md": `---
kind: hook
id: watch
description: watches
event: on-verify
corpus:
  - rule/process/cited
---
`,
		".keystone/learning/inbox/2026-11-27T10-00-00-x.md": `---
captured: 2026-11-27
source: review
proposed-layer: rules/process/x
---

body
`,
		".keystone/learning/inbox/2026-11-27T11-00-00-y.md": `---
captured: 2026-11-27
status: accepted
proposed-layer: hook
---

body
`,
	})
	rep, err := auditProject(root)
	if err != nil {
		t.Fatalf("auditProject: %v", err)
	}
	if len(rep.Learning.Pending) != 1 || len(rep.Learning.Accepted) != 1 {
		t.Errorf("learning = %+v / %+v", rep.Learning.Pending, rep.Learning.Accepted)
	}
	if rep.Pruning.Total != 5 {
		t.Errorf("total = %d, want 5", rep.Pruning.Total)
	}
	if len(rep.Pruning.Candidates) != 1 {
		t.Fatalf("pruning candidates = %+v, want 1", rep.Pruning.Candidates)
	}
	c := rep.Pruning.Candidates[0]
	if c.ID != "process/dead" {
		t.Errorf("candidate = %+v, want process/dead", c)
	}
	if len(c.Reasons) != 3 {
		t.Errorf("reasons = %v, want all three", c.Reasons)
	}
}

func TestAudit_EmptyCharterIsNotAnError(t *testing.T) {
	root := t.TempDir()
	rep, err := auditProject(root)
	if err != nil {
		t.Fatalf("auditProject on empty dir: %v", err)
	}
	if len(rep.Pruning.Candidates) != 0 || len(rep.Learning.Pending) != 0 {
		t.Errorf("report = %+v, want empty", rep)
	}
}
