package web

import (
	"path/filepath"
	"strings"
)

// sseTopic is the SSE `event:` name a watcher publish carries. Each
// live dashboard widget subscribes to the narrowest topic it cares
// about. `charter-changed` is the coarse fallback every change
// emits, so generic listeners always update.
type sseTopic string

const (
	topicCharter    sseTopic = "charter-changed"
	topicPrimitives sseTopic = "primitives-changed"
	topicInbox      sseTopic = "inbox-changed"
	topicPrune      sseTopic = "prune-changed"
)

// topicsForPath classifies a single dirty path into the SSE topics
// it should emit. The coarse `charter-changed` topic is always
// present, so widgets that don't care which subsystem moved still
// fire. Multiple narrow topics may also apply — e.g. a touch to
// `policies/<x>/.charter/...` is `primitives-changed` (the
// primitives a policy ships moved).
//
// Paths are matched on forward-slash form so the classification is
// platform-independent.
func topicsForPath(projectDir, path string) []sseTopic {
	rel := relPath(projectDir, path)
	rel = filepath.ToSlash(rel)

	out := []sseTopic{topicCharter}

	switch {
	case strings.Contains(rel, "/learning/inbox/") || strings.HasSuffix(rel, "/learning/inbox"):
		out = append(out, topicInbox, topicPrimitives)
	case strings.Contains(rel, "/.charter/policies/") || strings.Contains(rel, "/policies/"):
		out = append(out, topicPrimitives)
	case strings.HasSuffix(rel, "/INDEX.json") || strings.HasSuffix(rel, "INDEX.json"):
		out = append(out, topicPrimitives)
	// Specific files under the charter root must match before the generic
	// charter-dir case below — the lockfile lives at .charter/lockfile.json.
	case strings.Contains(rel, "/prune") || strings.HasSuffix(rel, "lockfile.json"):
		out = append(out, topicPrune)
	case strings.Contains(rel, "/.charter/") || strings.HasPrefix(rel, ".charter/"):
		out = append(out, topicPrimitives)
	}

	return out
}

// relPath returns `path` relative to `projectDir`, or `path` itself
// if it isn't under the project. fsnotify hands us absolute paths
// on every platform, so this is the common case.
func relPath(projectDir, path string) string {
	if rel, err := filepath.Rel(projectDir, path); err == nil && !strings.HasPrefix(rel, "..") {
		return rel
	}
	return path
}

// unionTopics collapses a list of per-path topic slices into a
// stable, de-duplicated set. `charter-changed` always sorts first;
// the rest sort alphabetically. Order is for testability — the
// SSE hub does not care about order.
func unionTopics(perPath [][]sseTopic) []sseTopic {
	seen := map[sseTopic]struct{}{}
	for _, ts := range perPath {
		for _, t := range ts {
			seen[t] = struct{}{}
		}
	}
	if len(seen) == 0 {
		return nil
	}
	out := make([]sseTopic, 0, len(seen))
	if _, ok := seen[topicCharter]; ok {
		out = append(out, topicCharter)
		delete(seen, topicCharter)
	}
	rest := make([]sseTopic, 0, len(seen))
	for t := range seen {
		rest = append(rest, t)
	}
	// Simple insertion sort — set is small (≤ 5 in practice).
	for i := 1; i < len(rest); i++ {
		for j := i; j > 0 && rest[j-1] > rest[j]; j-- {
			rest[j-1], rest[j] = rest[j], rest[j-1]
		}
	}
	return append(out, rest...)
}
