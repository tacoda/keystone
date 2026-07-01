package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

func searchCmd() *cobra.Command {
	var (
		dir    string
		limit  int
		format string
	)
	c := &cobra.Command{
		Use:   "search <query>",
		Short: "Search every primitive by id, description, globs, traces, and body",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			abs, err := filepath.Abs(dir)
			if err != nil {
				return err
			}
			primitives, _, err := primitive.Walk(abs, config.DefaultCharterRoot)
			if err != nil {
				return err
			}
			query := ""
			for i, a := range args {
				if i > 0 {
					query += " "
				}
				query += a
			}
			hits := primitive.Search(abs, primitives, query, limit)
			switch format {
			case "json":
				out, _ := json.MarshalIndent(map[string]any{
					"query": query,
					"count": len(hits),
					"hits":  hits,
				}, "", "  ")
				fmt.Println(string(out))
			default:
				if len(hits) == 0 {
					fmt.Fprintf(os.Stdout, "no hits for %q\n", query)
					return nil
				}
				fmt.Fprintf(os.Stdout, "%d hit(s) for %q:\n\n", len(hits), query)
				for _, h := range hits {
					fmt.Fprintf(os.Stdout, "  [%d] %s/%s — %s\n", h.Score, h.Primitive.Kind, h.Primitive.ID, h.Primitive.Description)
					if h.Excerpt != "" {
						fmt.Fprintf(os.Stdout, "      %s\n", h.Excerpt)
					}
					fmt.Fprintf(os.Stdout, "      %s (%s)\n\n", h.Primitive.Path, h.Where)
				}
			}
			return nil
		},
	}
	c.Flags().StringVar(&dir, "dir", ".", "Project root (defaults to cwd).")
	c.Flags().IntVar(&limit, "limit", 25, "Max hits to return (0 = all).")
	c.Flags().StringVar(&format, "format", "text", "Output format: text | json.")
	return c
}
