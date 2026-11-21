package main

import (
	"github.com/spf13/cobra"
)

// learningCmd is the Learning flywheel's inbox group: the candidates
// awaiting synthesis, newest first.
func learningCmd() *cobra.Command {
	c := &cobra.Command{
		Use:                "learning",
		Short:              "The Learning flywheel's inbox (candidates awaiting synthesis)",
		DisableFlagParsing: true,
	}
	c.AddCommand(&cobra.Command{
		Use:                "list",
		Short:              "List the learning inbox candidates with their proposed layers (promotion is synthesize's judgment)",
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLearningList(args)
		},
	})
	return c
}
