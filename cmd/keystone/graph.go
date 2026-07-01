package main

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

func graphCmd() *cobra.Command {
	var (
		dir    string
		format string
	)
	c := &cobra.Command{
		Use:   "graph",
		Short: "Print the primitive-relationship graph (deps + traces)",
		RunE: func(cmd *cobra.Command, args []string) error {
			abs, err := filepath.Abs(dir)
			if err != nil {
				return err
			}
			primitives, _, err := primitive.Walk(abs, config.DefaultCharterRoot)
			if err != nil {
				return err
			}
			fmt.Println(primitive.RenderGraph(primitives, primitive.GraphFormat(format)))
			return nil
		},
	}
	c.Flags().StringVar(&dir, "dir", ".", "Project root (defaults to cwd).")
	c.Flags().StringVar(&format, "format", "mermaid", "Output format: mermaid | dot.")
	return c
}
