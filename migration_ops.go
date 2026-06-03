package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// opStatus classifies the planned effect of a single operation.
type opStatus int

const (
	opNoop     opStatus = iota // target state already present; skip silently
	opCreate                   // file doesn't exist; will be written
	opChange                   // file exists; will be modified
	opConflict                 // can't apply automatically — user must intervene
)

// opResult is what planOperation returns: the read-out of current state, the
// proposed new state (nil for no-op/conflict), a status, and a human-readable
// note used for user-facing messages.
type opResult struct {
	op         Operation
	targetPath string // absolute path on disk
	status     opStatus
	current    []byte
	proposed   []byte
	note       string
}

// planOperation computes what op would do to the harness rooted at destDir,
// without writing anything. The caller decides whether to apply (writing
// proposed to targetPath) based on status and user input.
func planOperation(op Operation, destDir string) (opResult, error) {
	if op.Path == "" {
		return opResult{}, fmt.Errorf("operation %q is missing path", op.Type)
	}
	target := filepath.Join(destDir, filepath.FromSlash(op.Path))
	res := opResult{op: op, targetPath: target}

	current, err := os.ReadFile(target)
	if err != nil && !os.IsNotExist(err) {
		return res, err
	}
	exists := err == nil
	if exists {
		res.current = current
	}

	switch op.Type {
	case "add_file":
		return planAddFile(res, exists)
	case "frontmatter_set":
		return planFrontmatterSet(res, exists)
	case "ensure_section":
		return planEnsureSection(res, exists)
	case "replace_block":
		return planReplaceBlock(res, exists)
	default:
		return res, fmt.Errorf("unknown operation type %q", op.Type)
	}
}

// planAddFile: create a brand-new file. If the target already exists, the
// migration's intent has already been carried out — but only if the content
// is unchanged. Otherwise the user has customized it; surface as conflict.
func planAddFile(res opResult, exists bool) (opResult, error) {
	want := []byte(res.op.Content)
	if !exists {
		res.status = opCreate
		res.proposed = want
		return res, nil
	}
	if string(res.current) == string(want) {
		res.status = opNoop
		res.note = "file already matches expected content"
		return res, nil
	}
	res.status = opConflict
	res.note = "file exists with different content — review and merge manually"
	return res, nil
}

// planFrontmatterSet: set a YAML frontmatter key to a literal value, only if
// the key is not already present. The frontmatter is the leading `---`-fenced
// block; if absent, one is created. Existing values are never overwritten —
// the user has chosen them deliberately.
func planFrontmatterSet(res opResult, exists bool) (opResult, error) {
	if !exists {
		res.status = opConflict
		res.note = "target file missing — cannot set frontmatter"
		return res, nil
	}
	if res.op.Key == "" {
		return res, fmt.Errorf("frontmatter_set on %s requires key", res.op.Path)
	}

	body := string(res.current)
	fm, rest, hasFM := splitFrontmatter(body)

	if hasFM && frontmatterHasKey(fm, res.op.Key) {
		res.status = opNoop
		res.note = fmt.Sprintf("frontmatter already has %q", res.op.Key)
		return res, nil
	}

	newLine := fmt.Sprintf("%s: %s", res.op.Key, res.op.Value)
	var newBody string
	if hasFM {
		newFM := strings.TrimRight(fm, "\n") + "\n" + newLine + "\n"
		newBody = "---\n" + newFM + "---\n" + rest
	} else {
		newBody = "---\n" + newLine + "\n---\n\n" + body
	}

	res.status = opChange
	res.proposed = []byte(newBody)
	return res, nil
}

// planEnsureSection: append a heading + body to the file, anchored after a
// specific existing heading (the new section is inserted at the end of the
// region beginning at after_heading and ending at the next sibling heading,
// or at EOF). No-op if heading is already present anywhere in the file —
// presence is the idempotency marker.
func planEnsureSection(res opResult, exists bool) (opResult, error) {
	if !exists {
		res.status = opConflict
		res.note = "target file missing — cannot ensure section"
		return res, nil
	}
	if res.op.Heading == "" {
		return res, fmt.Errorf("ensure_section on %s requires heading", res.op.Path)
	}

	body := string(res.current)
	if strings.Contains(body, res.op.Heading) {
		res.status = opNoop
		res.note = fmt.Sprintf("section %q already present", res.op.Heading)
		return res, nil
	}

	insertion := strings.TrimRight(res.op.Heading, "\n") + "\n\n" +
		strings.TrimRight(res.op.Body, "\n") + "\n"

	var newBody string
	if res.op.AfterHeading == "" {
		// Append at EOF.
		newBody = strings.TrimRight(body, "\n") + "\n\n" + insertion
	} else {
		idx := strings.Index(body, res.op.AfterHeading)
		if idx < 0 {
			res.status = opConflict
			res.note = fmt.Sprintf("anchor %q not found — section not inserted", res.op.AfterHeading)
			return res, nil
		}
		// Find the next heading at the same level as after_heading (or any
		// heading at a level ≤ after_heading's). Simplest workable rule:
		// next line starting with the same `#`-prefix or fewer.
		level := headingLevel(res.op.AfterHeading)
		insertAt := findNextHeadingAtOrAbove(body, idx+len(res.op.AfterHeading), level)
		if insertAt < 0 {
			// no next heading — append at EOF
			newBody = strings.TrimRight(body, "\n") + "\n\n" + insertion
		} else {
			before := strings.TrimRight(body[:insertAt], "\n") + "\n\n"
			after := body[insertAt:]
			newBody = before + insertion + "\n" + after
		}
	}

	res.status = opChange
	res.proposed = []byte(newBody)
	return res, nil
}

// planReplaceBlock: replace `match` with `replacement` inside the target file,
// using exact-string match. If `replacement` is already present, no-op. If
// `match` cannot be found exactly, conflict — user must update manually.
func planReplaceBlock(res opResult, exists bool) (opResult, error) {
	if !exists {
		res.status = opConflict
		res.note = "target file missing — cannot replace block"
		return res, nil
	}
	if res.op.Match == "" || res.op.Replacement == "" {
		return res, fmt.Errorf("replace_block on %s requires match and replacement", res.op.Path)
	}

	body := string(res.current)
	if strings.Contains(body, res.op.Replacement) {
		res.status = opNoop
		res.note = "replacement already present"
		return res, nil
	}
	if !strings.Contains(body, res.op.Match) {
		res.status = opConflict
		res.note = "match block not found — file has diverged; update manually"
		return res, nil
	}

	newBody := strings.Replace(body, res.op.Match, res.op.Replacement, 1)
	res.status = opChange
	res.proposed = []byte(newBody)
	return res, nil
}

// splitFrontmatter pulls off a leading `---`-fenced block. Returns the
// frontmatter (without fences), the rest of the document, and a flag.
func splitFrontmatter(s string) (fm, rest string, ok bool) {
	if !strings.HasPrefix(s, "---\n") && !strings.HasPrefix(s, "---\r\n") {
		return "", s, false
	}
	body := strings.TrimPrefix(strings.TrimPrefix(s, "---\r\n"), "---\n")
	end := strings.Index(body, "\n---")
	if end < 0 {
		return "", s, false
	}
	fm = body[:end+1]
	// Skip past the closing fence + newline.
	tail := body[end+len("\n---"):]
	tail = strings.TrimPrefix(strings.TrimPrefix(tail, "\r\n"), "\n")
	return fm, tail, true
}

func frontmatterHasKey(fm, key string) bool {
	prefix := key + ":"
	for _, line := range strings.Split(fm, "\n") {
		trimmed := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(trimmed, prefix) {
			return true
		}
	}
	return false
}

// headingLevel returns the count of leading `#` in a markdown heading line,
// or 0 if it doesn't look like one.
func headingLevel(line string) int {
	n := 0
	for n < len(line) && line[n] == '#' {
		n++
	}
	if n == 0 || n >= len(line) || line[n] != ' ' {
		return 0
	}
	return n
}

// findNextHeadingAtOrAbove returns the byte index in body of the next heading
// (starting from after offset) whose level is ≤ maxLevel. Returns -1 if none.
// "At or above" in document terms = same level or shallower (#=1 > ##=2).
func findNextHeadingAtOrAbove(body string, offset, maxLevel int) int {
	if maxLevel <= 0 {
		return -1
	}
	i := offset
	for i < len(body) {
		nl := strings.IndexByte(body[i:], '\n')
		if nl < 0 {
			return -1
		}
		lineStart := i
		i += nl + 1
		line := body[lineStart : i-1]
		if lvl := headingLevel(line); lvl > 0 && lvl <= maxLevel {
			return lineStart
		}
	}
	return -1
}
