package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/tacoda/keystone/internal/framework/eval"
)

func evalCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "eval",
		Short: "Run harness evals (static + sensor levels)",
		Long: `keystone eval — measure how the harness behaves against
known scenarios. Each eval is a primitive of kind: eval, living at
.keystone/harness/evals/<id>/EVAL.md with a sibling expected.json
and optional fixture/.

Subcommands:
  run     Execute every eval (or a --filter subset) and write a report.
  list    Print every eval id discovered in the harness.

Phase A (this release) implements static + sensor levels. Agent-level
evals (LLM-driven, judge-graded) land in 2.1.`,
	}
	c.AddCommand(evalRunCmd())
	c.AddCommand(evalListCmd())
	return c
}

func evalRunCmd() *cobra.Command {
	var (
		dir      string
		filter   string
		report   string
		out      string
		baseline string
	)
	c := &cobra.Command{
		Use:   "run",
		Short: "Execute every eval and write a report",
		Long: `Run every declared eval (or a --filter subset) and emit a
report. Pass --baseline <git-ref> to also run the same evals against
a worktree of <ref> and report regressions/fixes/new/removed evals.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			abs, err := filepath.Abs(dir)
			if err != nil {
				return err
			}
			if baseline != "" {
				diff, err := eval.RunWithBaseline(context.Background(), abs, baseline, filter)
				if err != nil {
					return err
				}
				var body string
				if report == "json" {
					raw, _ := json.MarshalIndent(diff, "", "  ")
					body = string(raw)
				} else {
					body = eval.RenderBaselineMarkdown(diff)
				}
				if out != "" {
					if err := os.WriteFile(out, []byte(body), 0o644); err != nil {
						return err
					}
					fmt.Fprintf(os.Stdout, "  wrote: %s\n", out)
				} else {
					fmt.Fprintln(os.Stdout, body)
				}
				if diff.Summary.Regressions > 0 {
					return fmt.Errorf("baseline diff: %d regression(s) vs %s", diff.Summary.Regressions, baseline)
				}
				return nil
			}
			specs, err := eval.LoadAll(abs)
			if err != nil {
				return err
			}
			if len(specs) == 0 {
				fmt.Fprintln(os.Stdout, "no evals declared — `keystone new eval <id>` to add one")
				return nil
			}
			rep := eval.Run(context.Background(), abs, specs, filter)
			body, err := renderReport(rep, report)
			if err != nil {
				return err
			}
			if out != "" {
				if err := os.WriteFile(out, []byte(body), 0o644); err != nil {
					return err
				}
				fmt.Fprintf(os.Stdout, "  wrote: %s\n", out)
			} else {
				fmt.Fprintln(os.Stdout, body)
			}
			if rep.Failed > 0 {
				return fmt.Errorf("eval failed: %d eval(s) did not meet expectations", rep.Failed)
			}
			return nil
		},
	}
	c.Flags().StringVar(&dir, "dir", ".", "Project root (defaults to cwd).")
	c.Flags().StringVar(&filter, "filter", "", "Substring filter on eval ids.")
	c.Flags().StringVar(&report, "report", "md", "Report format: md | json.")
	c.Flags().StringVar(&out, "out", "", "Write the report to this file instead of stdout.")
	c.Flags().StringVar(&baseline, "baseline", "", "Compare current eval results against a git ref (e.g. main, v1.2.3). Materializes the ref in a worktree, runs both, reports regressions + fixes + new + removed.")
	return c
}

func evalListCmd() *cobra.Command {
	var dir string
	c := &cobra.Command{
		Use:   "list",
		Short: "Print every eval id discovered in the harness",
		RunE: func(cmd *cobra.Command, args []string) error {
			abs, err := filepath.Abs(dir)
			if err != nil {
				return err
			}
			ids, err := eval.ListIDs(abs)
			if err != nil {
				return err
			}
			for _, id := range ids {
				fmt.Println(id)
			}
			return nil
		},
	}
	c.Flags().StringVar(&dir, "dir", ".", "Project root (defaults to cwd).")
	return c
}

func renderReport(rep eval.Report, format string) (string, error) {
	switch format {
	case "json":
		out, err := json.MarshalIndent(rep, "", "  ")
		if err != nil {
			return "", err
		}
		return string(out), nil
	case "md", "markdown", "":
		return eval.RenderMarkdown(rep), nil
	default:
		return "", fmt.Errorf("unknown --report format %q (json|md)", format)
	}
}
