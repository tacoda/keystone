package main

import (
	"bufio"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type migrateFlags struct {
	dir     string
	apply   bool // skip prompts; write every non-conflict change
	dryRun  bool // never write, never prompt
	from    string
}

func runMigrate(args []string, assets embed.FS) error {
	flags, err := parseMigrateArgs(args)
	if err != nil {
		return err
	}

	absDir, err := filepath.Abs(flags.dir)
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}
	if _, err := os.Stat(filepath.Join(absDir, "harness")); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no harness/ in %s — run `keystone init` first", absDir)
		}
		return err
	}

	from := flags.from
	if from == "" {
		from, err = readKeystoneVersion(absDir)
		if err != nil {
			return fmt.Errorf("read install profile: %w", err)
		}
	}
	if from == "dev" {
		fmt.Fprintf(os.Stdout, "▸ harness is on a dev version; skipping migration (use --from <version> to override)\n")
		return nil
	}

	migrations, err := loadMigrations(assets, from)
	if err != nil {
		return fmt.Errorf("load migrations: %w", err)
	}
	if len(migrations) == 0 {
		if from == "" {
			fmt.Fprintf(os.Stdout, "✓ no migrations available\n")
		} else {
			fmt.Fprintf(os.Stdout, "✓ harness is up to date (at %s)\n", from)
		}
		return nil
	}

	versions := groupByVersion(migrations)
	fmt.Fprintf(os.Stdout, "keystone migrate — %d migration(s) pending across %d version(s)\n",
		len(migrations), len(versions))
	if from != "" {
		fmt.Fprintf(os.Stdout, "  current keystone_version: %s\n", from)
	}
	if flags.dryRun {
		fmt.Fprintf(os.Stdout, "  mode: dry-run (nothing will be written)\n")
	}
	fmt.Fprintln(os.Stdout)

	reader := bufio.NewReader(os.Stdin)
	tty := isTerminal(os.Stdin)

	for _, vg := range versions {
		fmt.Fprintf(os.Stdout, "═══ %s ═══\n\n", vg.version)
		quit := false
		for _, m := range vg.migrations {
			if err := processMigration(m, absDir, flags, reader, tty, &quit); err != nil {
				return err
			}
			if quit {
				break
			}
		}
		if quit {
			fmt.Fprintf(os.Stdout, "\n! halted at %s (keystone_version not bumped)\n", vg.version)
			return nil
		}
		if !flags.dryRun {
			if err := updateKeystoneVersion(absDir, vg.version); err != nil {
				return fmt.Errorf("bump keystone_version: %w", err)
			}
			fmt.Fprintf(os.Stdout, "  ✓ keystone_version → %s\n\n", vg.version)
		} else {
			fmt.Fprintf(os.Stdout, "  (dry-run: would bump keystone_version → %s)\n\n", vg.version)
		}
	}

	fmt.Fprintf(os.Stdout, "✓ migration complete\n")
	return nil
}

type versionGroup struct {
	version    string
	migrations []Migration
}

func groupByVersion(ms []Migration) []versionGroup {
	var out []versionGroup
	for _, m := range ms {
		if len(out) == 0 || out[len(out)-1].version != m.Version {
			out = append(out, versionGroup{version: m.Version})
		}
		out[len(out)-1].migrations = append(out[len(out)-1].migrations, m)
	}
	return out
}

func processMigration(m Migration, destDir string, flags *migrateFlags, reader *bufio.Reader, tty bool, quit *bool) error {
	fmt.Fprintf(os.Stdout, "▸ %s/%s", m.Version, m.ID)
	if m.Description != "" {
		fmt.Fprintf(os.Stdout, " — %s", m.Description)
	}
	fmt.Fprintln(os.Stdout)

	for i, op := range m.Operations {
		res, err := planOperation(op, destDir)
		if err != nil {
			return fmt.Errorf("%s op %d: %w", m.SourcePath, i+1, err)
		}
		label := fmt.Sprintf("  [%d/%d] %s", i+1, len(m.Operations), op.Path)
		switch res.status {
		case opNoop:
			fmt.Fprintf(os.Stdout, "%s  (no change — %s)\n", label, res.note)
			continue
		case opConflict:
			fmt.Fprintf(os.Stdout, "%s\n    ! conflict: %s\n", label, res.note)
			continue
		}

		// opCreate / opChange — show preview then prompt (or auto-apply).
		fmt.Fprintf(os.Stdout, "%s\n", label)
		printOpPreview(res)

		if flags.dryRun {
			fmt.Fprintln(os.Stdout, "    (dry-run: not written)")
			continue
		}

		apply := flags.apply
		if !apply {
			if !tty {
				return fmt.Errorf("interactive prompt needed but stdin is not a TTY — pass --apply or --dry-run")
			}
			choice, err := promptYesNoQuit(reader)
			if err != nil {
				return err
			}
			switch choice {
			case "y":
				apply = true
			case "n":
				fmt.Fprintln(os.Stdout, "    skipped")
				continue
			case "q":
				*quit = true
				return nil
			}
		}

		if apply {
			if err := writeOpResult(res); err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "    ✓ wrote %s\n", res.targetPath)
		}
	}
	fmt.Fprintln(os.Stdout)
	return nil
}

// printOpPreview renders a small, op-type-aware diff. For creates we show the
// proposed content with `+` prefix; for changes we show the relevant slice
// (added section, frontmatter line, or before/after match block).
func printOpPreview(res opResult) {
	switch res.op.Type {
	case "add_file":
		fmt.Fprintln(os.Stdout, "    add_file (new):")
		printPrefixed(string(res.proposed), "    + ")
	case "frontmatter_set":
		fmt.Fprintf(os.Stdout, "    frontmatter_set: %s = %s\n", res.op.Key, res.op.Value)
		fmt.Fprintf(os.Stdout, "    + %s: %s\n", res.op.Key, res.op.Value)
	case "ensure_section":
		fmt.Fprintf(os.Stdout, "    ensure_section: %q\n", res.op.Heading)
		printPrefixed(strings.TrimRight(res.op.Heading, "\n")+"\n\n"+strings.TrimRight(res.op.Body, "\n")+"\n", "    + ")
	case "replace_block":
		fmt.Fprintf(os.Stdout, "    replace_block in %s:\n", res.op.Path)
		printPrefixed(res.op.Match, "    - ")
		printPrefixed(res.op.Replacement, "    + ")
	}
}

func printPrefixed(s, prefix string) {
	for _, line := range strings.Split(strings.TrimRight(s, "\n"), "\n") {
		fmt.Fprintf(os.Stdout, "%s%s\n", prefix, line)
	}
}

func writeOpResult(res opResult) error {
	if err := os.MkdirAll(filepath.Dir(res.targetPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(res.targetPath, res.proposed, 0o644)
}

func promptYesNoQuit(reader *bufio.Reader) (string, error) {
	for {
		fmt.Fprint(os.Stdout, "    Apply? [y/N/q]: ")
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		switch strings.ToLower(strings.TrimSpace(line)) {
		case "y", "yes":
			return "y", nil
		case "", "n", "no":
			return "n", nil
		case "q", "quit":
			return "q", nil
		default:
			fmt.Fprintln(os.Stdout, "    please answer y, n, or q")
		}
	}
}

func parseMigrateArgs(args []string) (*migrateFlags, error) {
	flags := &migrateFlags{dir: "."}
	positional := []string{}

	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--help" || a == "-h":
			printMigrateUsage(os.Stdout)
			os.Exit(0)
		case a == "--apply" || a == "-y":
			flags.apply = true
		case a == "--dry-run":
			flags.dryRun = true
		case a == "--from":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--from requires a value")
			}
			flags.from = args[i+1]
			i++
		case strings.HasPrefix(a, "--from="):
			flags.from = strings.TrimPrefix(a, "--from=")
		case strings.HasPrefix(a, "-"):
			return nil, fmt.Errorf("unknown flag %s", a)
		default:
			positional = append(positional, a)
		}
	}
	if len(positional) > 1 {
		return nil, fmt.Errorf("migrate takes at most one positional argument (got %d)", len(positional))
	}
	if len(positional) == 1 {
		flags.dir = positional[0]
	}
	if flags.apply && flags.dryRun {
		return nil, fmt.Errorf("--apply and --dry-run are mutually exclusive")
	}
	return flags, nil
}

func printMigrateUsage(w *os.File) {
	fmt.Fprint(w, `keystone migrate — apply pending harness migrations

Usage:
  keystone migrate [<dir>] [--apply|-y] [--dry-run] [--from <version>]

Reads harness/corpus/state/INSTALL_PROFILE.md to find the current
keystone_version, then applies every embedded migration newer than that.
By default each planned change is previewed and prompted for. After all
migrations in a version directory are processed, keystone_version is bumped.

Flags:
  --apply, -y    Apply every non-conflict change without prompting.
  --dry-run      Preview every change; write nothing.
  --from <v>     Override the version recorded in INSTALL_PROFILE.md.

Per-prompt options:
  y    apply this change
  n    skip this change (default)
  q    quit; the current version is not bumped, re-run to resume
`)
}
