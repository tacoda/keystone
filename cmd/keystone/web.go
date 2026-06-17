package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/tacoda/keystone/internal/framework/web"
)

// webCmd is the parent of `keystone web <sub>`. Currently only
// `serve` is wired; future verbs (open, snapshot, …) hang off the
// same tree.
func webCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "web",
		Short: "Local HTMX dashboard + read-only REST API for the harness",
	}
	c.AddCommand(webServeCmd())
	return c
}

func webServeCmd() *cobra.Command {
	var (
		port int
		dir  string
	)
	c := &cobra.Command{
		Use:   "serve",
		Short: "Run the dashboard + REST API on localhost",
		Long: `Run the keystone dashboard on localhost. Single port; same
origin for HTML, REST API, and SSE.

The dashboard at http://127.0.0.1:<port>/ shows the harness inventory,
configured external sources, and primitive detail pages. The REST API
under /api/ is read-only. Server-Sent Events at /events push HTMX
fragments whenever a file in .keystone/ changes — the dashboard
updates without polling.

Default port 4773 ("KEYS" on a phone keypad). Override with --port.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer cancel()
			return web.Serve(ctx, web.Options{ProjectDir: dir, Port: port})
		},
	}
	c.Flags().IntVar(&port, "port", web.DefaultPort, "Localhost port to bind.")
	c.Flags().StringVar(&dir, "dir", ".", "Project root (defaults to cwd).")
	return c
}
