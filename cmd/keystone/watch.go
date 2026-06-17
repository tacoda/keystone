package main

import (
	"context"
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

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

func watchCmd() *cobra.Command {
	var (
		dir       string
		lint      bool
		project   bool
		debounce  time.Duration
	)
	c := &cobra.Command{
		Use:   "watch",
		Short: "Watch .keystone/harness/ and re-run index + project + lint on change",
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
			root := filepath.Join(abs, config.DefaultHarnessRoot)
			if _, err := os.Stat(root); err != nil {
				return fmt.Errorf("no harness at %s — run `keystone init` first", root)
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
				mu     sync.Mutex
				timer  *time.Timer
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
	if strings.HasPrefix(base, ".") && base != ".keystone" {
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
func runPipeline(ctx context.Context, projectDir string, doLint, doProject bool) {
	t0 := time.Now()
	primitives, _, err := primitive.Walk(projectDir, config.DefaultHarnessRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ walk: %v\n", err)
		return
	}
	idx := primitive.Build(primitives, time.Now())
	outPath := filepath.Join(projectDir, config.KeystoneDir(config.DefaultHarnessRoot), config.IndexName)
	if err := primitive.Write(outPath, idx); err != nil {
		fmt.Fprintf(os.Stderr, "✗ write index: %v\n", err)
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
