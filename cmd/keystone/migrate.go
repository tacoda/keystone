package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/tacoda/keystone/internal/framework/migrate"
)

type migrateFlags struct {
	dir         string
	apply       bool // skip prompts; write every non-conflict change
	dryRun      bool // never write, never prompt
	from        string
	harnessRoot string
}

func runMigrate(args []string, assets fs.FS) error {
	harnessRoot, args, err := extractHarnessRoot(args)
	if err != nil {
		return err
	}
	flags, err := parseMigrateArgs(args)
	if err != nil {
		return err
	}
	flags.harnessRoot = harnessRoot

	absDir, err := filepath.Abs(flags.dir)
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}
	if _, err := os.Stat(filepath.Join(absDir, harnessRoot)); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no %s/ in %s — run `keystone init` first", harnessRoot, absDir)
		}
		return err
	}

	from := flags.from
	if from == "" {
		from, err = readKeystoneVersion(absDir, harnessRoot)
		if err != nil {
			return fmt.Errorf("read install profile: %w", err)
		}
	}
	if from == "dev" {
		fmt.Fprintf(os.Stdout, "▸ harness is on a dev version; skipping migration (use --from <version> to override)\n")
		return nil
	}

	migrations, err := migrate.Load(assets, from)
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
			if err := updateKeystoneVersion(absDir, harnessRoot, vg.version); err != nil {
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
	migrations []migrate.Migration
}

func groupByVersion(ms []migrate.Migration) []versionGroup {
	var out []versionGroup
	for _, m := range ms {
		if len(out) == 0 || out[len(out)-1].version != m.Version {
			out = append(out, versionGroup{version: m.Version})
		}
		out[len(out)-1].migrations = append(out[len(out)-1].migrations, m)
	}
	return out
}

func processMigration(m migrate.Migration, destDir string, flags *migrateFlags, reader *bufio.Reader, tty bool, quit *bool) error {
	fmt.Fprintf(os.Stdout, "▸ %s/%s", m.Version, m.ID)
	if m.Description != "" {
		fmt.Fprintf(os.Stdout, " — %s", m.Description)
	}
	fmt.Fprintln(os.Stdout)

	for i, op := range m.Operations {
		res, err := migrate.PlanOperation(op, destDir)
		if err != nil {
			return fmt.Errorf("%s op %d: %w", m.SourcePath, i+1, err)
		}
		label := fmt.Sprintf("  [%d/%d] %s", i+1, len(m.Operations), op.Path)
		switch res.Status {
		case migrate.OpNoop:
			fmt.Fprintf(os.Stdout, "%s  (no change — %s)\n", label, res.Note)
			continue
		case migrate.OpConflict:
			fmt.Fprintf(os.Stdout, "%s\n    ! conflict: %s\n", label, res.Note)
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
			if err := migrate.WriteOpResult(res); err != nil {
				return err
			}
			switch res.Op.Type {
			case "move_dir", "move_file":
				fmt.Fprintf(os.Stdout, "    ✓ moved %s → %s\n", res.Op.Path, res.Op.To)
			case "delete_dir", "delete_file":
				fmt.Fprintf(os.Stdout, "    ✓ removed %s\n", res.Op.Path)
			default:
				fmt.Fprintf(os.Stdout, "    ✓ wrote %s\n", res.TargetPath)
			}
		}
	}
	fmt.Fprintln(os.Stdout)
	return nil
}

// printOpPreview renders a small, op-type-aware diff. For creates we show the
// proposed content with `+` prefix; for changes we show the relevant slice
// (added section, frontmatter line, or before/after match block).
func printOpPreview(res migrate.OpResult) {
	switch res.Op.Type {
	case "add_file":
		fmt.Fprintln(os.Stdout, "    add_file (new):")
		printPrefixed(string(res.Proposed), "    + ")
	case "frontmatter_set":
		fmt.Fprintf(os.Stdout, "    frontmatter_set: %s = %s\n", res.Op.Key, res.Op.Value)
		fmt.Fprintf(os.Stdout, "    + %s: %s\n", res.Op.Key, res.Op.Value)
	case "ensure_section":
		fmt.Fprintf(os.Stdout, "    ensure_section: %q\n", res.Op.Heading)
		printPrefixed(strings.TrimRight(res.Op.Heading, "\n")+"\n\n"+strings.TrimRight(res.Op.Body, "\n")+"\n", "    + ")
	case "replace_block":
		fmt.Fprintf(os.Stdout, "    replace_block in %s:\n", res.Op.Path)
		printPrefixed(res.Op.Match, "    - ")
		printPrefixed(res.Op.Replacement, "    + ")
	case "move_dir":
		fmt.Fprintf(os.Stdout, "    move_dir: %s → %s\n", res.Op.Path, res.Op.To)
		for _, p := range res.MovePlans {
			switch p.Status {
			case migrate.OpCreate:
				fmt.Fprintf(os.Stdout, "    + %s\n", p.RelPath)
			case migrate.OpConflict:
				fmt.Fprintf(os.Stdout, "    ! %s (%s)\n", p.RelPath, p.Note)
			case migrate.OpNoop:
				fmt.Fprintf(os.Stdout, "      %s (already at destination)\n", p.RelPath)
			}
		}
	case "move_file":
		fmt.Fprintf(os.Stdout, "    move_file: %s → %s\n", res.Op.Path, res.Op.To)
		if len(res.MovePlans) == 1 {
			p := res.MovePlans[0]
			switch p.Status {
			case migrate.OpCreate:
				fmt.Fprintf(os.Stdout, "    + %s\n", p.RelPath)
			case migrate.OpConflict:
				fmt.Fprintf(os.Stdout, "    ! %s (%s)\n", p.RelPath, p.Note)
			case migrate.OpNoop:
				fmt.Fprintf(os.Stdout, "      %s (already at destination)\n", p.RelPath)
			}
		}
	case "delete_dir":
		fmt.Fprintf(os.Stdout, "    delete_dir: %s\n", res.Op.Path)
		fmt.Fprintf(os.Stdout, "    - %s/\n", res.Op.Path)
	case "delete_file":
		fmt.Fprintf(os.Stdout, "    delete_file: %s\n", res.Op.Path)
		fmt.Fprintf(os.Stdout, "    - %s\n", res.Op.Path)
	}
}

func printPrefixed(s, prefix string) {
	for _, line := range strings.Split(strings.TrimRight(s, "\n"), "\n") {
		fmt.Fprintf(os.Stdout, "%s%s\n", prefix, line)
	}
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
