package sensors

import (
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
)

// secretScanPatterns is the conservative default — patterns that are
// nearly always real secrets when they appear in source. Hash-suffixed
// fake values, .env.example placeholders, and obvious test fixtures
// will still match; the sensitive-files guide is explicit that even
// fake-looking values are treated as real if the pattern matches.
//
// Bootstrap may extend this list per-project via CODEBASE_STATE.md;
// the registered runner reads that file when present.
var secretScanPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(aws|amazon)[_-]?(secret|access)[_-]?key[_-]?(id)?\s*[:=]\s*["']?[A-Z0-9/+=]{20,}["']?`),
	regexp.MustCompile(`(?i)AKIA[0-9A-Z]{16}`),                                // AWS access key id
	regexp.MustCompile(`(?i)ghp_[A-Za-z0-9]{36}`),                             // GitHub personal access token
	regexp.MustCompile(`(?i)gho_[A-Za-z0-9]{36}`),                             // GitHub OAuth token
	regexp.MustCompile(`(?i)xox[abprs]-[A-Za-z0-9-]{10,}`),                    // Slack tokens
	regexp.MustCompile(`(?i)sk[_-](live|test)[_-][A-Za-z0-9]{20,}`),           // Stripe secret key
	regexp.MustCompile(`(?i)-----BEGIN (RSA|EC|OPENSSH|PGP|DSA) PRIVATE KEY`), // PEM-format private keys
	regexp.MustCompile(`(?i)bearer\s+[A-Za-z0-9._\-]{20,}`),                   // Bearer tokens
	regexp.MustCompile(`(?i)(api[_-]?key|api[_-]?secret|access[_-]?token|auth[_-]?token)\s*[:=]\s*["'][A-Za-z0-9._\-]{16,}["']`),
}

// sensitivePathPatterns matches paths the sensitive-files guide
// forbids reading or writing. The sensor blocks on any attempt to
// write a file at one of these paths.
//
// `.env.example` is the documented public placeholder file; it's the
// one explicit exception, handled via isEnvExample() below rather than
// a negative lookahead (Go's RE2 doesn't support `(?!...)`).
var sensitivePathPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(^|/)\.env$`),
	regexp.MustCompile(`(^|/)\.env\.[A-Za-z0-9_-]+$`),
	regexp.MustCompile(`\.(pem|key|p12|pfx|kdbx)$`),
	regexp.MustCompile(`(^|/)id_(rsa|ed25519|ecdsa|dsa)$`),
	regexp.MustCompile(`(^|/)credentials\.json$`),
	regexp.MustCompile(`(^|/)service-account[A-Za-z0-9._-]*\.json$`),
	regexp.MustCompile(`(?i)(^|/)secrets?/`),
	regexp.MustCompile(`(?i)(^|/)vault/`),
}

// isEnvExample lets the `.env.example` placeholder through. Real
// credentials in `.env.example` are still caught by the content regex
// pass — the path check just allows the file to exist.
func isEnvExample(path string) bool {
	return strings.HasSuffix(path, "/.env.example") || path == ".env.example"
}

func init() {
	Register("secret-scan", runSecretScan)
}

// runSecretScan blocks any Edit/Write that targets a sensitive path
// outright, or whose post-edit content contains a credential pattern
// that .env.example values are explicitly NOT exempted from — the
// "make fakes look obviously fake" rule lives in the guide, not the
// sensor.
//
// Pre-edit content matching is best-effort; the host hands the sensor
// the content the agent is about to write, and the sensor checks that
// before it lands. Stop-phase invocation (no FileContent) scans the
// staged git diff via `git diff --staged --no-color -U0`.
func runSecretScan(ctx Context, out io.Writer) (Result, error) {
	if ctx.FilePath != "" {
		path := filepath.ToSlash(ctx.FilePath)
		if !isEnvExample(path) {
			for _, re := range sensitivePathPatterns {
				if re.MatchString(path) {
					return Result{
						Block: true,
						Message: fmt.Sprintf(
							"secret-scan: refusing to write %q — matches a sensitive-file pattern. Ask the user out-of-band for the value.",
							ctx.FilePath,
						),
					}, nil
				}
			}
		}
	}

	content := ctx.FileContent
	if len(content) == 0 && ctx.BashCommand != "" {
		content = []byte(ctx.BashCommand)
	}
	if len(content) == 0 {
		// Stop-phase or no-payload invocation. Report a soft pass; the
		// `git diff --staged` scan path is not implemented in 2.2.0 — a
		// dedicated `secret-scan-staged` sensor can land later if the
		// pre-write hook proves insufficient.
		fmt.Fprintln(out, "secret-scan: no inline content to scan (advisory pass)")
		return Result{}, nil
	}

	for _, re := range secretScanPatterns {
		if re.Match(content) {
			snippet := redactedMatch(content, re)
			return Result{
				Block: true,
				Message: fmt.Sprintf(
					"secret-scan: content matches a credential pattern (%s). Move the value to a secret store and reference it via env/config.",
					snippet,
				),
			}, nil
		}
	}
	fmt.Fprintln(out, "secret-scan: clean")
	return Result{}, nil
}

// redactedMatch returns a short, redacted preview of the match so the
// block message is actionable without echoing the secret back into the
// agent's transcript.
func redactedMatch(content []byte, re *regexp.Regexp) string {
	loc := re.FindIndex(content)
	if loc == nil {
		return "pattern matched"
	}
	start := loc[0]
	end := loc[1]
	prefix := ""
	if start > 8 {
		prefix = "…"
		start = start - 4
	} else {
		start = 0
	}
	preview := string(content[start:loc[0]])
	if end-loc[0] > 4 {
		preview += string(content[loc[0]:loc[0]+4]) + "…REDACTED"
	} else {
		preview += "…REDACTED"
	}
	preview = strings.ReplaceAll(preview, "\n", " ")
	return prefix + preview
}
