package main

import (
	"fmt"
	"io"
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
	targetPath string // absolute path on disk (single-file ops); empty for multi-file
	status     opStatus
	current    []byte
	proposed   []byte
	note       string

	// movePlans is populated only by move_dir; one entry per file the op will
	// touch. Aggregate status above reflects the worst sub-status across these
	// (any conflict → conflict; all noop → noop; otherwise change).
	movePlans []movePlan
}

// movePlan is one file-level action inside a move_dir operation.
type movePlan struct {
	sourceAbs string
	destAbs   string
	relPath   string // for display
	status    opStatus
	note      string
}

// planOperation computes what op would do to the harness rooted at destDir,
// without writing anything. The caller decides whether to apply (writing
// proposed to targetPath) based on status and user input.
func planOperation(op Operation, destDir string) (opResult, error) {
	if op.Path == "" {
		return opResult{}, fmt.Errorf("operation %q is missing path", op.Type)
	}

	// Directory-scoped ops handle their own existence + content checks.
	switch op.Type {
	case "move_dir":
		return planMoveDir(op, destDir)
	case "delete_dir":
		return planDeleteDir(op, destDir)
	case "move_file":
		return planMoveFile(op, destDir)
	case "delete_file":
		return planDeleteFile(op, destDir)
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

// planMoveDir walks every file under op.Path and computes, for each, the
// transition into op.To. Idempotent: files already at the destination with
// matching content are no-ops; files at the destination with diverged content
// surface as conflicts. The aggregate status is the worst sub-status (any
// conflict bubbles up).
func planMoveDir(op Operation, destDir string) (opResult, error) {
	if op.Path == "" || op.To == "" {
		return opResult{op: op}, fmt.Errorf("move_dir requires both path and to")
	}
	res := opResult{op: op}
	src := filepath.Join(destDir, filepath.FromSlash(op.Path))
	dst := filepath.Join(destDir, filepath.FromSlash(op.To))
	res.targetPath = src // applyMoveDir uses this to attempt cleanup of the empty source dir

	srcInfo, srcErr := os.Stat(src)
	dstInfo, dstErr := os.Stat(dst)

	srcExists := srcErr == nil
	dstExists := dstErr == nil

	if srcErr != nil && !os.IsNotExist(srcErr) {
		return res, srcErr
	}
	if dstErr != nil && !os.IsNotExist(dstErr) {
		return res, dstErr
	}

	if !srcExists {
		// Source already gone. Either previously moved or never installed.
		if dstExists && dstInfo.IsDir() {
			res.status = opNoop
			res.note = fmt.Sprintf("source %s already relocated", op.Path)
		} else {
			res.status = opNoop
			res.note = fmt.Sprintf("source %s not present — nothing to move", op.Path)
		}
		return res, nil
	}
	if !srcInfo.IsDir() {
		return res, fmt.Errorf("move_dir source %s is not a directory", op.Path)
	}

	worst := opNoop
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(src, path)
		if relErr != nil {
			return relErr
		}
		destPath := filepath.Join(dst, rel)
		plan := movePlan{
			sourceAbs: path,
			destAbs:   destPath,
			relPath:   filepath.ToSlash(rel),
		}

		srcBytes, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}

		destBytes, destReadErr := os.ReadFile(destPath)
		switch {
		case os.IsNotExist(destReadErr):
			plan.status = opCreate
			if worst < opCreate {
				worst = opCreate
			}
		case destReadErr != nil:
			return destReadErr
		case string(destBytes) == string(srcBytes):
			plan.status = opNoop
			plan.note = "destination matches source — will just remove source"
			if worst < opChange {
				worst = opChange // we still have work to do (remove source)
			}
		default:
			plan.status = opConflict
			plan.note = "destination exists with different content"
			worst = opConflict
		}

		res.movePlans = append(res.movePlans, plan)
		return nil
	})
	if err != nil {
		return res, err
	}

	if len(res.movePlans) == 0 {
		// Source dir is empty — just need cleanup.
		res.status = opChange
		res.note = "source directory is empty — will remove"
		return res, nil
	}
	res.status = worst
	return res, nil
}

// planMoveFile reports the planned relocation of a single file from Path to
// To. Idempotent: missing source no-ops, matching destination removes the
// source, diverged destination conflicts.
func planMoveFile(op Operation, destDir string) (opResult, error) {
	if op.Path == "" || op.To == "" {
		return opResult{op: op}, fmt.Errorf("move_file requires both path and to")
	}
	res := opResult{op: op}
	src := filepath.Join(destDir, filepath.FromSlash(op.Path))
	dst := filepath.Join(destDir, filepath.FromSlash(op.To))
	res.targetPath = src

	srcBytes, srcErr := os.ReadFile(src)
	dstBytes, dstErr := os.ReadFile(dst)

	srcExists := srcErr == nil
	dstExists := dstErr == nil

	if srcErr != nil && !os.IsNotExist(srcErr) {
		return res, srcErr
	}
	if dstErr != nil && !os.IsNotExist(dstErr) {
		return res, dstErr
	}

	if !srcExists {
		if dstExists {
			res.status = opNoop
			res.note = fmt.Sprintf("source %s already relocated to %s", op.Path, op.To)
		} else {
			res.status = opNoop
			res.note = fmt.Sprintf("source %s not present — nothing to move", op.Path)
		}
		return res, nil
	}

	plan := movePlan{
		sourceAbs: src,
		destAbs:   dst,
		relPath:   op.To,
	}
	switch {
	case !dstExists:
		plan.status = opCreate
		res.status = opCreate
	case string(dstBytes) == string(srcBytes):
		plan.status = opNoop
		plan.note = "destination matches source — will just remove source"
		res.status = opChange
	default:
		plan.status = opConflict
		plan.note = "destination exists with different content"
		res.status = opConflict
		res.note = plan.note
	}
	res.movePlans = []movePlan{plan}
	return res, nil
}

// applyMoveFile performs a single planned file relocation. Conflicts skip
// silently (source left in place for the user to handle).
func applyMoveFile(res opResult) error {
	if len(res.movePlans) == 0 {
		return nil
	}
	p := res.movePlans[0]
	switch p.status {
	case opConflict:
		return nil
	case opNoop:
		if err := os.Remove(p.sourceAbs); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove %s: %w", p.sourceAbs, err)
		}
		return nil
	case opCreate:
		if err := os.MkdirAll(filepath.Dir(p.destAbs), 0o755); err != nil {
			return err
		}
		if err := copyFile(p.sourceAbs, p.destAbs); err != nil {
			return err
		}
		if err := os.Remove(p.sourceAbs); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove %s: %w", p.sourceAbs, err)
		}
	}
	return nil
}

// planDeleteFile reports the planned removal of a single file. Idempotent:
// missing target is a no-op.
func planDeleteFile(op Operation, destDir string) (opResult, error) {
	res := opResult{op: op}
	target := filepath.Join(destDir, filepath.FromSlash(op.Path))
	res.targetPath = target

	info, err := os.Stat(target)
	if err != nil {
		if os.IsNotExist(err) {
			res.status = opNoop
			res.note = "file already absent"
			return res, nil
		}
		return res, err
	}
	if info.IsDir() {
		return res, fmt.Errorf("delete_file target %s is a directory", op.Path)
	}
	res.status = opChange
	return res, nil
}

// applyDeleteFile removes the file at the target path.
func applyDeleteFile(res opResult) error {
	if err := os.Remove(res.targetPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// planDeleteDir reports the planned removal of an empty directory. Conflicts
// if the directory still contains files (so the user notices residue).
func planDeleteDir(op Operation, destDir string) (opResult, error) {
	res := opResult{op: op}
	target := filepath.Join(destDir, filepath.FromSlash(op.Path))
	res.targetPath = target

	info, err := os.Stat(target)
	if err != nil {
		if os.IsNotExist(err) {
			res.status = opNoop
			res.note = "directory already absent"
			return res, nil
		}
		return res, err
	}
	if !info.IsDir() {
		return res, fmt.Errorf("delete_dir target %s is not a directory", op.Path)
	}

	entries, err := os.ReadDir(target)
	if err != nil {
		return res, err
	}
	if len(entries) > 0 {
		res.status = opConflict
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		res.note = "directory not empty: " + strings.Join(names, ", ")
		return res, nil
	}
	res.status = opChange
	return res, nil
}

// applyMoveDir performs the planned moves on disk: copies each source file to
// its destination, then unlinks the source. Removes the source directory if
// emptied. Conflicts in movePlans are skipped (leaving the source file in
// place for the user to handle).
func applyMoveDir(res opResult) error {
	src := filepath.Join(filepath.Dir(res.targetPath), "") // unused; we use plans
	_ = src
	for _, p := range res.movePlans {
		if p.status == opConflict {
			continue
		}
		if p.status == opNoop {
			// Destination already matches — just remove the source.
			if err := os.Remove(p.sourceAbs); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("remove %s: %w", p.sourceAbs, err)
			}
			continue
		}
		// opCreate: copy then remove source.
		if err := os.MkdirAll(filepath.Dir(p.destAbs), 0o755); err != nil {
			return err
		}
		if err := copyFile(p.sourceAbs, p.destAbs); err != nil {
			return err
		}
		if err := os.Remove(p.sourceAbs); err != nil {
			return fmt.Errorf("remove %s: %w", p.sourceAbs, err)
		}
	}
	// Attempt to remove the (now-empty) source directory and any empty parents
	// up to but not including the harness root. Silent if anything is still
	// occupied — `delete_dir` exists for explicit cleanup.
	_ = removeIfEmpty(res.targetPath)
	return nil
}

// applyDeleteDir removes a directory that planDeleteDir already verified is
// empty.
func applyDeleteDir(res opResult) error {
	if err := os.Remove(res.targetPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// removeIfEmpty unlinks dir if it has no entries left. Silent if dir is gone
// or non-empty.
func removeIfEmpty(dir string) error {
	if dir == "" {
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	if len(entries) > 0 {
		return nil
	}
	return os.Remove(dir)
}

// copyFile streams src to dst. Truncates dst if it exists (planMoveDir should
// have flagged a content mismatch already, so reaching here means destination
// is absent or content already matches).
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
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
