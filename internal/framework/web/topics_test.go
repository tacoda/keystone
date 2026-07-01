package web

import (
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

// TestTopicsForPath exercises the path → SSE topic classifier.
// Every input emits the coarse `charter-changed` topic; narrow
// topics layer on per pattern.
func TestTopicsForPath(t *testing.T) {
	project := filepath.FromSlash("/proj")
	cases := []struct {
		name string
		path string
		want []sseTopic
	}{
		{
			name: "inbox candidate triggers inbox + primitives",
			path: filepath.FromSlash("/proj/.charter/learning/inbox/2026-06-18.md"),
			want: []sseTopic{topicCharter, topicInbox, topicPrimitives},
		},
		{
			name: "policies tree triggers primitives",
			path: filepath.FromSlash("/proj/.charter/policies/tacoda-org/charter/guides/x.md"),
			want: []sseTopic{topicCharter, topicPrimitives},
		},
		{
			name: "INDEX.json triggers primitives",
			path: filepath.FromSlash("/proj/.charter/INDEX.json"),
			want: []sseTopic{topicCharter, topicPrimitives},
		},
		{
			name: "charter tree triggers primitives",
			path: filepath.FromSlash("/proj/.charter/guides/process/spec.md"),
			want: []sseTopic{topicCharter, topicPrimitives},
		},
		{
			name: "lockfile triggers prune",
			path: filepath.FromSlash("/proj/.charter/lockfile.json"),
			want: []sseTopic{topicCharter, topicPrune},
		},
		{
			name: "random doc only triggers charter",
			path: filepath.FromSlash("/proj/notes.md"),
			want: []sseTopic{topicCharter},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := topicsForPath(project, c.path)
			// Compare as sorted sets — the classifier is order-stable
			// today, but the contract is set-equality.
			sortTopics(got)
			sortTopics(c.want)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("topicsForPath(%q):\n  got  %v\n  want %v", c.path, got, c.want)
			}
		})
	}
}

// TestUnionTopics confirms multiple per-path topic lists collapse
// into a stable set with `charter-changed` sorted first.
func TestUnionTopics(t *testing.T) {
	in := [][]sseTopic{
		{topicCharter, topicPrimitives},
		{topicCharter, topicPrune},
		{topicCharter, topicInbox, topicPrimitives},
	}
	got := unionTopics(in)
	if len(got) == 0 || got[0] != topicCharter {
		t.Fatalf("charter-changed must lead the union, got %v", got)
	}
	if len(got) != 4 {
		t.Errorf("union should contain charter + primitives + prune + inbox; got %v", got)
	}
}

// TestUnionTopics_Empty makes sure an empty input returns nil so
// callers can safely fall back to a coarse-topic-only publish.
func TestUnionTopics_Empty(t *testing.T) {
	if got := unionTopics(nil); got != nil {
		t.Errorf("expected nil on empty input, got %v", got)
	}
	if got := unionTopics([][]sseTopic{}); got != nil {
		t.Errorf("expected nil on empty slice input, got %v", got)
	}
}

func sortTopics(ts []sseTopic) {
	sort.Slice(ts, func(i, j int) bool { return string(ts[i]) < string(ts[j]) })
}
