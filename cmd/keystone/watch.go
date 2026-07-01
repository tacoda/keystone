package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"

	"github.com/tacoda/keystone/internal/framework/adapters/agnostic"
	"github.com/tacoda/keystone/internal/framework/adapters/aider"
	"github.com/tacoda/keystone/internal/framework/adapters/claudecode"
	"github.com/tacoda/keystone/internal/framework/adapters/continueide"
	"github.com/tacoda/keystone/internal/framework/adapters/cursor"
	"github.com/tacoda/keystone/internal/framework/adapters/opencode"
	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

func watchCmd() *cobra.Command {
	var (
		dir      string
		lint     bool
		project  bool
		debounce time.Duration
	)
	c := &cobra.Command{
		Use:   "watch",
		Short: "Watch .charter/ and re-run index + project + lint on change",
		Long: `Long-running file watcher. Debounces bursts of writes and
runs the configured pipeline:

  - keystone index           (always)
  - keystone project         (if --project, default on)
  - keystone lint            (if --lint, default off)

Pair with the dashboard at /metrics or /insights — both pick up the
fresh INDEX.json via their existing SSE channel.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			abs, err := filepath.Abs(dir)
			if err != nil {
				return err
			}
			root := filepath.Join(abs, config.DefaultCharterRoot)
			if _, err := os.Stat(root); err != nil {
				return fmt.Errorf("no charter at %s — run `keystone init` first", root)
			}

			w, err := fsnotify.NewWatcher()
			if err != nil {
				return err
			}
			defer w.Close()

			if err := watchRecursive(w, root); err != nil {
				return err
			}

			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			fmt.Fprintf(os.Stdout, "▸ watching %s (debounce %s)\n", root, debounce)

			var (
				mu      sync.Mutex
				timer   *time.Timer
				cancel2 context.CancelFunc
			)
			schedule := func() {
				mu.Lock()
				defer mu.Unlock()
				if timer != nil {
					timer.Stop()
				}
				timer = time.AfterFunc(debounce, func() {
					if cancel2 != nil {
						cancel2()
					}
					var runCtx context.Context
					runCtx, cancel2 = context.WithCancel(ctx)
					defer cancel2()
					runPipeline(runCtx, abs, lint, project)
				})
			}

			for {
				select {
				case <-ctx.Done():
					return nil
				case ev, ok := <-w.Events:
					if !ok {
						return nil
					}
					if !relevantEvent(ev) {
						continue
					}
					if ev.Op&fsnotify.Create == fsnotify.Create {
						if info, err := os.Stat(ev.Name); err == nil && info.IsDir() {
							_ = watchRecursive(w, ev.Name)
						}
					}
					schedule()
				case err, ok := <-w.Errors:
					if !ok {
						return nil
					}
					fmt.Fprintf(os.Stderr, "keystone watch: %v\n", err)
				}
			}
		},
	}
	c.Flags().StringVar(&dir, "dir", ".", "Project root (defaults to cwd).")
	c.Flags().BoolVar(&lint, "lint", false, "Also run `keystone lint` after index.")
	c.Flags().BoolVar(&project, "project", true, "Also run `keystone project` after index.")
	c.Flags().DurationVar(&debounce, "debounce", 300*time.Millisecond, "Debounce window before running the pipeline.")
	return c
}

func relevantEvent(ev fsnotify.Event) bool {
	base := filepath.Base(ev.Name)
	if strings.HasPrefix(base, ".") && base != filepath.Base(config.DefaultCharterRoot) {
		return false
	}
	if strings.HasSuffix(base, "~") || strings.HasSuffix(base, ".swp") {
		return false
	}
	return true
}

func watchRecursive(w *fsnotify.Watcher, root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if info.IsDir() {
			return w.Add(path)
		}
		return nil
	})
}

// runPipeline does what each cycle of the watcher runs. Inline —
// avoids shelling out so the watcher stays responsive on bursty
// edits.
//
// Pipeline order:
//  1. Walk primitives (one disk scan for the whole cycle)
//  2. Write INDEX.json + INDEX.lite.json
//  3. (if doProject) project primitives → .claude/* via primitive.Project
//  4. (if doProject) project agnostic AGENTS.md
//  5. (if doProject) project claudecode hooks → .claude/settings.json
//  6. (if doProject) per-host adapters from keystone.json `adapters:` list
//  7. (if doLint) run primitive.Lint and report
//
// Every adapter is wired here so a user editing or saving a guide /
// sensor / persona sees the projection update in one debounce window
// — no manual `keystone project` invocation needed.
func runPipeline(ctx context.Context, projectDir string, doLint, doProject bool) {
	t0 := time.Now()
	primitives, _, err := primitive.Walk(projectDir, config.DefaultCharterRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ walk: %v\n", err)
		return
	}
	composed, composeErrs := primitive.Compose(primitives)
	for _, e := range composeErrs {
		fmt.Fprintf(os.Stderr, "  ⚠ %s\n", e.Error())
	}
	primitives = composed
	idx := primitive.Build(primitives, time.Now())
	keystoneDir := filepath.Join(projectDir, config.KeystoneDir(config.DefaultCharterRoot))
	outPath := filepath.Join(keystoneDir, config.IndexName)
	if err := primitive.Write(outPath, idx); err != nil {
		fmt.Fprintf(os.Stderr, "✗ write index: %v\n", err)
		return
	}
	if err := primitive.WriteLite(filepath.Join(keystoneDir, config.IndexLiteName), primitive.BuildLite(idx)); err != nil {
		fmt.Fprintf(os.Stderr, "✗ write index.lite: %v\n", err)
		return
	}
	rel, _ := filepath.Rel(projectDir, outPath)
	fmt.Fprintf(os.Stdout, "✓ %s (%d primitives", rel, len(idx.Primitives))

	if doProject {
		if _, err := primitive.Project(projectDir, primitives); err != nil {
			fmt.Fprintf(os.Stderr, "; ✗ project: %v", err)
		} else {
			fmt.Fprint(os.Stdout, ", projections refreshed")
		}
		runWatchAdapters(projectDir, primitives)
	}
	if doLint {
		findings := primitive.Lint(primitives)
		errCount := 0
		for _, f := range findings {
			if f.Severity == primitive.FindingError {
				errCount++
			}
		}
		if errCount > 0 {
			fmt.Fprintf(os.Stdout, ", ✗ %d lint error(s)", errCount)
		} else {
			fmt.Fprint(os.Stdout, ", lint clean")
		}
	}
	fmt.Fprintf(os.Stdout, ") in %s\n", time.Since(t0).Round(time.Millisecond))
}

// runWatchAdapters mirrors the adapter fan-out in `keystone project`.
// Failures from any single adapter log a warning but don't abort the
// watch loop — a half-finished cross-host projection is better than
// the watcher dying on a transient I/O error.
func runWatchAdapters(projectDir string, primitives []primitive.Primitive) {
	// Agnostic AGENTS.md — always emitted.
	if _, err := agnostic.ProjectAgentsMD(projectDir, agnostic.DefaultBody()); err != nil {
		fmt.Fprintf(os.Stderr, "; ✗ agnostic AGENTS.md: %v", err)
	}
	// Claude Code settings.json — hooks + posture, always emitted
	// (claude-code is the default host).
	watchClaudeCode(projectDir, primitives)
	// Opt-in adapters from keystone.json.
	cfg, cfgErr := config.ReadProjectConfig(projectDir)
	if cfgErr != nil && !errors.Is(cfgErr, os.ErrNotExist) {
		fmt.Fprintf(os.Stderr, "; ✗ read keystone.json: %v", cfgErr)
		return
	}
	if cfg == nil {
		return
	}
	if cfg.HasAdapter(config.AdapterCursor) {
		if _, err := cursor.ProjectRules(projectDir, primitives); err != nil {
			fmt.Fprintf(os.Stderr, "; ✗ cursor: %v", err)
		}
	}
	if cfg.HasAdapter(config.AdapterAider) {
		if _, err := aider.ProjectAider(projectDir, agnostic.DefaultBody()); err != nil {
			fmt.Fprintf(os.Stderr, "; ✗ aider: %v", err)
		}
	}
	if cfg.HasAdapter(config.AdapterContinue) {
		if _, err := continueide.ProjectRules(projectDir, primitives); err != nil {
			fmt.Fprintf(os.Stderr, "; ✗ continue: %v", err)
		}
	}
	if cfg.HasAdapter(config.AdapterOpenCode) {
		if _, err := opencode.ProjectAgents(projectDir, primitives); err != nil {
			fmt.Fprintf(os.Stderr, "; ✗ opencode: %v", err)
		}
	}
}

// watchClaudeCode re-projects the Claude Code settings.json regions (hooks +
// posture) during a watch cycle, logging failures to stderr without aborting.
func watchClaudeCode(projectDir string, primitives []primitive.Primitive) {
	if _, err := claudecode.ProjectHooks(projectDir, primitives); err != nil {
		fmt.Fprintf(os.Stderr, "; ✗ claudecode hooks: %v", err)
	}
	if _, err := claudecode.ProjectPosture(projectDir, primitives); err != nil {
		fmt.Fprintf(os.Stderr, "; ✗ claudecode posture: %v", err)
	}
}
