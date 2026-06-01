package main

import (
	"embed"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func runInit(args []string, assets embed.FS) error {
	flagArgs, positional, err := splitArgs(args)
	if err != nil {
		return err
	}

	flags := flag.NewFlagSet("init", flag.ContinueOnError)
	agent := flags.String("agent", "", "agent target to install")
	force := flags.Bool("force", false, "overwrite existing harness/ without prompting")
	if err := flags.Parse(flagArgs); err != nil {
		return err
	}

	dir := "."
	if len(positional) > 0 {
		dir = positional[0]
	}
	if len(positional) > 1 {
		return fmt.Errorf("init takes at most one positional argument (got %d)", len(positional))
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}
	if info, err := os.Stat(absDir); err != nil {
		return fmt.Errorf("dir %s: %w", absDir, err)
	} else if !info.IsDir() {
		return fmt.Errorf("dir %s is not a directory", absDir)
	}

	resolved := *agent
	if resolved == "" {
		resolved = detectAgent(absDir)
		if resolved == "" {
			return fmt.Errorf("no --agent provided and detection found no marker files in %s; pass --agent <name>", absDir)
		}
		fmt.Fprintf(os.Stdout, "▸ detected agent: %s\n", resolved)
	}
	if !isSupportedAgent(resolved) {
		return fmt.Errorf("unknown agent %q (supported: %v)", resolved, supportedAgents)
	}

	preflight(absDir)

	harnessDest := filepath.Join(absDir, "harness")
	if _, err := os.Stat(harnessDest); err == nil {
		if !*force {
			return fmt.Errorf("%s already exists; pass --force to overwrite", harnessDest)
		}
		fmt.Fprintf(os.Stdout, "▸ overwriting existing harness/\n")
	} else if !os.IsNotExist(err) {
		return err
	}

	fmt.Fprintf(os.Stdout, "▸ installing corpus to %s\n", harnessDest)
	if err := copyTree(assets, "harness", harnessDest, overwrite); err != nil {
		return fmt.Errorf("copy harness: %w", err)
	}

	targetRoot := filepath.Join("targets", resolved)
	if _, err := assets.Open(targetRoot); err != nil {
		fmt.Fprintf(os.Stderr, "! no target bundle for %s; corpus installed, configure activation manually\n", resolved)
	} else {
		fmt.Fprintf(os.Stdout, "▸ installing %s target\n", resolved)
		if err := copyTree(assets, targetRoot, absDir, skipIfExists); err != nil {
			return fmt.Errorf("copy target: %w", err)
		}
	}

	printNextSteps(resolved)
	return nil
}

// splitArgs separates flag tokens from positional tokens so that flags may
// appear before or after the positional dir argument.
func splitArgs(args []string) (flagArgs, positional []string, err error) {
	booleanFlags := map[string]bool{"--force": true, "-force": true}
	valueFlags := map[string]bool{"--agent": true, "-agent": true}

	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case booleanFlags[a]:
			flagArgs = append(flagArgs, a)
		case valueFlags[a]:
			if i+1 >= len(args) {
				return nil, nil, fmt.Errorf("flag %s requires a value", a)
			}
			flagArgs = append(flagArgs, a, args[i+1])
			i++
		case len(a) > 0 && a[0] == '-':
			// Support --flag=value form for both kinds.
			flagArgs = append(flagArgs, a)
		default:
			positional = append(positional, a)
		}
	}
	return flagArgs, positional, nil
}

func preflight(dir string) {
	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
		fmt.Fprintf(os.Stderr, "! %s is not a git repository; keystone assumes git-tracked projects\n", dir)
	}

	ciCandidates := []string{
		".github/workflows",
		".gitlab-ci.yml",
		".circleci/config.yml",
		".travis.yml",
		"azure-pipelines.yml",
		"bitbucket-pipelines.yml",
		"Jenkinsfile",
	}
	for _, c := range ciCandidates {
		if _, err := os.Stat(filepath.Join(dir, c)); err == nil {
			return
		}
	}
	fmt.Fprintf(os.Stderr, "! no CI config detected; the release phase assumes a CI pipeline\n")
}

func printNextSteps(agent string) {
	fmt.Fprintf(os.Stdout, `
✓ keystone installed for %s.

Next steps:

  1. Read harness/README.md
  2. Run the bootstrap action in your agent to populate
     harness/state/CODEBASE_STATE.md and harness/idioms/<your-stack>/
     from your project.
  3. Commit harness/ and any agent-specific files this installer created.
`, agent)
}
