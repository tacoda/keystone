// Package sensors hosts the small subset of sensor implementations
// that keystone owns end-to-end — checks that aren't usefully delegated
// to an external tool. Other sensors (build, test, lint, vuln-scan) are
// invoked by their host_triggers as direct shell commands (`go build`,
// `go test`, `golangci-lint`, `govulncheck`); keystone has no value to
// add by wrapping them.
//
// Each sensor has a Runner that consumes a Context and returns a
// Result. The CLI exit code derives from Result.Block:
//
//	Block == false  → exit 0 (advisory or pass)
//	Block == true   → exit 2 (Claude Code's block-with-message code)
//
// New sensors register themselves via Register() in their own init().
package sensors

import (
	"fmt"
	"io"
	"sort"
)

// Context carries inputs the host hook hands the sensor at runtime.
// Fields beyond ProjectDir are best-effort: Claude Code's hook protocol
// surfaces tool_input.file_path and tool_input.content via stdin JSON,
// but a sensor may run with neither (e.g., Stop-phase batch checks).
type Context struct {
	ProjectDir string
	// FilePath is the file the host is about to edit/write. Empty for
	// hooks that don't have a single file target (Stop, Bash matchers).
	FilePath string
	// FileContent is the post-edit content the host is about to write.
	// Pre-populated for PreToolUse Edit/Write hooks; empty otherwise.
	FileContent []byte
	// BashCommand is the shell command a PreToolUse:Bash hook is about
	// to run. Empty for non-Bash matchers.
	BashCommand string
}

// Result is what a sensor reports back. Message is rendered to stderr
// on block; informational messages render to stdout regardless.
type Result struct {
	Block   bool
	Message string
}

// Runner is the contract every registered sensor implements.
type Runner func(ctx Context, out io.Writer) (Result, error)

var registry = map[string]Runner{}

// Register adds a sensor runner to the registry. Duplicate ids panic
// at init time — sensors must have unique ids matching their primitive
// frontmatter.
func Register(id string, run Runner) {
	if _, exists := registry[id]; exists {
		panic(fmt.Sprintf("sensors.Register: duplicate id %q", id))
	}
	registry[id] = run
}

// Run dispatches a sensor by id. Returns ErrUnknownSensor if no runner
// is registered — the caller decides whether that's a hard error or a
// soft skip (see verify.go).
func Run(id string, ctx Context, out io.Writer) (Result, error) {
	run, ok := registry[id]
	if !ok {
		return Result{}, ErrUnknownSensor{ID: id}
	}
	return run(ctx, out)
}

// IDs returns the set of registered sensor ids, sorted, for help output.
func IDs() []string {
	ids := make([]string, 0, len(registry))
	for id := range registry {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// ErrUnknownSensor is returned by Run when a caller asks for a sensor
// id with no registered runner. Hook plumbing should treat this as a
// soft skip (exit 0) so a partial implementation doesn't block edits.
type ErrUnknownSensor struct{ ID string }

func (e ErrUnknownSensor) Error() string {
	return fmt.Sprintf("no sensor runner registered for %q", e.ID)
}
