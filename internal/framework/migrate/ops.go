package migrate

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// PlanOperation computes what op would do to the harness rooted at destDir,
// without writing anything. The caller decides whether to apply (writing
// Proposed to TargetPath) based on Status and user input.
func PlanOperation(op Operation, destDir string) (OpResult, error) {
	if op.Path == "" {
		return OpResult{}, fmt.Errorf("operation %q is missing path", op.Type)
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
	res := OpResult{Op: op, TargetPath: target}

	current, err := os.ReadFile(target)
	if err != nil && !os.IsNotExist(err) {
		return res, err
	}
	exists := err == nil
	if exists {
		res.Current = current
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

// WriteOpResult applies the planned change to disk. Dispatches by op type;
// no-op for conflicts.
func WriteOpResult(res OpResult) error {
	switch res.Op.Type {
	case "move_dir":
		return applyMoveDir(res)
	case "move_file":
		return applyMoveFile(res)
	case "delete_dir":
		return applyDeleteDir(res)
	case "delete_file":
		return applyDeleteFile(res)
	default:
		if err := os.MkdirAll(filepath.Dir(res.TargetPath), 0o755); err != nil {
			return err
		}
		return os.WriteFile(res.TargetPath, res.Proposed, 0o644)
	}
}

// planAddFile: create a brand-new file. If the target already exists, the
// migration's intent has already been carried out — but only if the content
// is unchanged. Otherwise the user has customized it; surface as conflict.
func planAddFile(res OpResult, exists bool) (OpResult, error) {
	want := []byte(res.Op.Content)
	if !exists {
		res.Status = OpCreate
		res.Proposed = want
		return res, nil
	}
	if string(res.Current) == string(want) {
		res.Status = OpNoop
		res.Note = "file already matches expected content"
		return res, nil
	}
	res.Status = OpConflict
	res.Note = "file exists with different content — review and merge manually"
	return res, nil
}

// planFrontmatterSet: set a YAML frontmatter key to a literal value, only if
// the key is not already present. The frontmatter is the leading `---`-fenced
// block; if absent, one is created. Existing values are never overwritten —
// the user has chosen them deliberately.
func planFrontmatterSet(res OpResult, exists bool) (OpResult, error) {
	if !exists {
		res.Status = OpConflict
		res.Note = "target file missing — cannot set frontmatter"
		return res, nil
	}
	if res.Op.Key == "" {
		return res, fmt.Errorf("frontmatter_set on %s requires key", res.Op.Path)
	}

	body := string(res.Current)
	fm, rest, hasFM := splitFrontmatter(body)

	if hasFM && frontmatterHasKey(fm, res.Op.Key) {
		res.Status = OpNoop
		res.Note = fmt.Sprintf("frontmatter already has %q", res.Op.Key)
		return res, nil
	}

	newLine := fmt.Sprintf("%s: %s", res.Op.Key, res.Op.Value)
	var newBody string
	if hasFM {
		newFM := strings.TrimRight(fm, "\n") + "\n" + newLine + "\n"
		newBody = "---\n" + newFM + "---\n" + rest
	} else {
		newBody = "---\n" + newLine + "\n---\n\n" + body
	}

	res.Status = OpChange
	res.Proposed = []byte(newBody)
	return res, nil
}

// planEnsureSection: append a heading + body to the file, anchored after a
// specific existing heading.
func planEnsureSection(res OpResult, exists bool) (OpResult, error) {
	if !exists {
		res.Status = OpConflict
		res.Note = "target file missing — cannot ensure section"
		return res, nil
	}
	if res.Op.Heading == "" {
		return res, fmt.Errorf("ensure_section on %s requires heading", res.Op.Path)
	}

	body := string(res.Current)
	if strings.Contains(body, res.Op.Heading) {
		res.Status = OpNoop
		res.Note = fmt.Sprintf("section %q already present", res.Op.Heading)
		return res, nil
	}

	insertion := strings.TrimRight(res.Op.Heading, "\n") + "\n\n" +
		strings.TrimRight(res.Op.Body, "\n") + "\n"

	var newBody string
	if res.Op.AfterHeading == "" {
		newBody = strings.TrimRight(body, "\n") + "\n\n" + insertion
	} else {
		idx := strings.Index(body, res.Op.AfterHeading)
		if idx < 0 {
			res.Status = OpConflict
			res.Note = fmt.Sprintf("anchor %q not found — section not inserted", res.Op.AfterHeading)
			return res, nil
		}
		level := headingLevel(res.Op.AfterHeading)
		insertAt := findNextHeadingAtOrAbove(body, idx+len(res.Op.AfterHeading), level)
		if insertAt < 0 {
			newBody = strings.TrimRight(body, "\n") + "\n\n" + insertion
		} else {
			before := strings.TrimRight(body[:insertAt], "\n") + "\n\n"
			after := body[insertAt:]
			newBody = before + insertion + "\n" + after
		}
	}

	res.Status = OpChange
	res.Proposed = []byte(newBody)
	return res, nil
}

// planReplaceBlock: replace `match` with `replacement` inside the target file,
// using exact-string match.
func planReplaceBlock(res OpResult, exists bool) (OpResult, error) {
	if !exists {
		res.Status = OpConflict
		res.Note = "target file missing — cannot replace block"
		return res, nil
	}
	if res.Op.Match == "" || res.Op.Replacement == "" {
		return res, fmt.Errorf("replace_block on %s requires match and replacement", res.Op.Path)
	}

	body := string(res.Current)
	if strings.Contains(body, res.Op.Replacement) {
		res.Status = OpNoop
		res.Note = "replacement already present"
		return res, nil
	}
	if !strings.Contains(body, res.Op.Match) {
		res.Status = OpConflict
		res.Note = "match block not found — file has diverged; update manually"
		return res, nil
	}

	newBody := strings.Replace(body, res.Op.Match, res.Op.Replacement, 1)
	res.Status = OpChange
	res.Proposed = []byte(newBody)
	return res, nil
}

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

func planMoveDir(op Operation, destDir string) (OpResult, error) {
	if op.Path == "" || op.To == "" {
		return OpResult{Op: op}, fmt.Errorf("move_dir requires both path and to")
	}
	res := OpResult{Op: op}
	src := filepath.Join(destDir, filepath.FromSlash(op.Path))
	dst := filepath.Join(destDir, filepath.FromSlash(op.To))
	res.TargetPath = src

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
		if dstExists && dstInfo.IsDir() {
			res.Status = OpNoop
			res.Note = fmt.Sprintf("source %s already relocated", op.Path)
		} else {
			res.Status = OpNoop
			res.Note = fmt.Sprintf("source %s not present — nothing to move", op.Path)
		}
		return res, nil
	}
	if !srcInfo.IsDir() {
		return res, fmt.Errorf("move_dir source %s is not a directory", op.Path)
	}

	worst := OpNoop
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
		plan := MovePlan{
			SourceAbs: path,
			DestAbs:   destPath,
			RelPath:   filepath.ToSlash(rel),
		}

		srcBytes, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}

		destBytes, destReadErr := os.ReadFile(destPath)
		switch {
		case os.IsNotExist(destReadErr):
			plan.Status = OpCreate
			if worst < OpCreate {
				worst = OpCreate
			}
		case destReadErr != nil:
			return destReadErr
		case string(destBytes) == string(srcBytes):
			plan.Status = OpNoop
			plan.Note = "destination matches source — will just remove source"
			if worst < OpChange {
				worst = OpChange
			}
		default:
			plan.Status = OpConflict
			plan.Note = "destination exists with different content"
			worst = OpConflict
		}

		res.MovePlans = append(res.MovePlans, plan)
		return nil
	})
	if err != nil {
		return res, err
	}

	if len(res.MovePlans) == 0 {
		res.Status = OpChange
		res.Note = "source directory is empty — will remove"
		return res, nil
	}
	res.Status = worst
	return res, nil
}

func planMoveFile(op Operation, destDir string) (OpResult, error) {
	if op.Path == "" || op.To == "" {
		return OpResult{Op: op}, fmt.Errorf("move_file requires both path and to")
	}
	res := OpResult{Op: op}
	src := filepath.Join(destDir, filepath.FromSlash(op.Path))
	dst := filepath.Join(destDir, filepath.FromSlash(op.To))
	res.TargetPath = src

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
			res.Status = OpNoop
			res.Note = fmt.Sprintf("source %s already relocated to %s", op.Path, op.To)
		} else {
			res.Status = OpNoop
			res.Note = fmt.Sprintf("source %s not present — nothing to move", op.Path)
		}
		return res, nil
	}

	plan := MovePlan{
		SourceAbs: src,
		DestAbs:   dst,
		RelPath:   op.To,
	}
	switch {
	case !dstExists:
		plan.Status = OpCreate
		res.Status = OpCreate
	case string(dstBytes) == string(srcBytes):
		plan.Status = OpNoop
		plan.Note = "destination matches source — will just remove source"
		res.Status = OpChange
	default:
		plan.Status = OpConflict
		plan.Note = "destination exists with different content"
		res.Status = OpConflict
		res.Note = plan.Note
	}
	res.MovePlans = []MovePlan{plan}
	return res, nil
}

func applyMoveFile(res OpResult) error {
	if len(res.MovePlans) == 0 {
		return nil
	}
	p := res.MovePlans[0]
	switch p.Status {
	case OpConflict:
		return nil
	case OpNoop:
		if err := os.Remove(p.SourceAbs); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove %s: %w", p.SourceAbs, err)
		}
		return nil
	case OpCreate:
		if err := os.MkdirAll(filepath.Dir(p.DestAbs), 0o755); err != nil {
			return err
		}
		if err := copyFile(p.SourceAbs, p.DestAbs); err != nil {
			return err
		}
		if err := os.Remove(p.SourceAbs); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove %s: %w", p.SourceAbs, err)
		}
	}
	return nil
}

func planDeleteFile(op Operation, destDir string) (OpResult, error) {
	res := OpResult{Op: op}
	target := filepath.Join(destDir, filepath.FromSlash(op.Path))
	res.TargetPath = target

	info, err := os.Stat(target)
	if err != nil {
		if os.IsNotExist(err) {
			res.Status = OpNoop
			res.Note = "file already absent"
			return res, nil
		}
		return res, err
	}
	if info.IsDir() {
		return res, fmt.Errorf("delete_file target %s is a directory", op.Path)
	}
	res.Status = OpChange
	return res, nil
}

func applyDeleteFile(res OpResult) error {
	if err := os.Remove(res.TargetPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func planDeleteDir(op Operation, destDir string) (OpResult, error) {
	res := OpResult{Op: op}
	target := filepath.Join(destDir, filepath.FromSlash(op.Path))
	res.TargetPath = target

	info, err := os.Stat(target)
	if err != nil {
		if os.IsNotExist(err) {
			res.Status = OpNoop
			res.Note = "directory already absent"
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
		res.Status = OpConflict
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		res.Note = "directory not empty: " + strings.Join(names, ", ")
		return res, nil
	}
	res.Status = OpChange
	return res, nil
}

func applyMoveDir(res OpResult) error {
	for _, p := range res.MovePlans {
		if p.Status == OpConflict {
			continue
		}
		if p.Status == OpNoop {
			if err := os.Remove(p.SourceAbs); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("remove %s: %w", p.SourceAbs, err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(p.DestAbs), 0o755); err != nil {
			return err
		}
		if err := copyFile(p.SourceAbs, p.DestAbs); err != nil {
			return err
		}
		if err := os.Remove(p.SourceAbs); err != nil {
			return fmt.Errorf("remove %s: %w", p.SourceAbs, err)
		}
	}
	_ = removeIfEmpty(res.TargetPath)
	return nil
}

func applyDeleteDir(res OpResult) error {
	if err := os.Remove(res.TargetPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

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
