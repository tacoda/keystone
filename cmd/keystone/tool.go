package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
	"github.com/tacoda/keystone/internal/framework/toolrun"
)

// runTool dispatches `keystone tool run|list`. `run` is how a cli or plugin
// tool reaches a host: the host (or the agent, from AGENTS.md) runs this
// command, and keystone runs the tool's `run:` exactly as its MCP server
// does for an mcp tool.
func runTool(args []string) error {
	if len(args) == 0 {
		printToolUsage(os.Stdout)
		return nil
	}
	switch args[0] {
	case "help", "--help", "-h":
		printToolUsage(os.Stdout)
		return nil
	case "run":
		return runToolRun(args[1:])
	case "list":
		return runToolList(args[1:])
	}
	return fmt.Errorf("unknown tool subcommand %q (use: run, list)", args[0])
}

func runToolRun(args []string) error {
	id, dir, toolArgs, err := parseToolRun(args)
	if err != nil {
		return err
	}
	p, err := findTool(dir, id)
	if err != nil {
		return err
	}
	out, err := toolrun.Run(context.Background(), dir, p, toolArgs)
	fmt.Fprint(os.Stdout, out)
	return err
}

// parseToolRun reads `<id> [--dir D] [--<arg> <value> | --<arg>=<value>]...`.
func parseToolRun(args []string) (id, dir string, toolArgs map[string]string, err error) {
	dir, toolArgs = ".", map[string]string{}
	var rest []string
	for i := 0; i < len(args); i++ {
		name, value, ok, next := toolFlag(args, i)
		switch {
		case !ok:
			rest = append(rest, args[i])
		case next < 0:
			return "", "", nil, fmt.Errorf("flag --%s requires a value", name)
		case name == "dir":
			dir = value
		default:
			toolArgs[name] = value
		}
		i = max(i, next)
	}
	if len(rest) != 1 {
		return "", "", nil, fmt.Errorf("`keystone tool run` requires exactly one <id>")
	}
	abs, err := filepath.Abs(dir)
	return rest[0], abs, toolArgs, err
}

// toolFlag reads the flag at args[i]: its name, its value, whether it is a
// flag at all, and the index of the last arg it used (-1 when the value is
// missing).
func toolFlag(args []string, i int) (name, value string, ok bool, next int) {
	name, isFlag := strings.CutPrefix(args[i], "--")
	if !isFlag {
		return "", "", false, i
	}
	if n, v, hasEq := strings.Cut(name, "="); hasEq {
		return n, v, true, i
	}
	if i+1 >= len(args) {
		return name, "", true, -1
	}
	return name, args[i+1], true, i + 1
}

// findTool returns the tool primitive with id.
func findTool(dir, id string) (primitive.Primitive, error) {
	prims, _, err := primitive.Walk(dir, config.DefaultAgentsRoot)
	if err != nil {
		return primitive.Primitive{}, err
	}
	for _, p := range prims {
		if p.Kind == string(primitive.KindTool) && p.ID == id {
			return p, nil
		}
	}
	return primitive.Primitive{}, fmt.Errorf("no tool %q in %s/tools/", id, config.DefaultAgentsRoot)
}

func runToolList(args []string) error {
	dir := "."
	help, err := parseFlags(args, map[string]*string{"--dir": &dir}, nil, func() { printToolUsage(os.Stdout) })
	if help || err != nil {
		return err
	}
	prims, _, err := primitive.Walk(dir, config.DefaultAgentsRoot)
	if err != nil {
		return err
	}
	for _, p := range prims {
		if p.Kind == string(primitive.KindTool) {
			fmt.Fprintf(os.Stdout, "%-24s %-7s %s\n    %s\n", p.ID, p.Transport, p.Description, primitive.ToolCommand(p))
		}
	}
	return nil
}

func printToolUsage(w io.Writer) {
	fmt.Fprint(w, `keystone tool — run the charter's tools

Usage:
    keystone tool list [--dir <path>]
    keystone tool run <id> [--<arg> <value>]... [--dir <path>]

A tool (.agents/tools/<id>.md) declares a run: handler and its args:.
Each argument reaches the handler as KEYSTONE_ARG_<NAME>.

By transport:
  mcp     the keystone MCP server serves it as keystone_tool_<id>
  cli     AGENTS.md lists it with its keystone tool run command
  plugin  as cli, plus a native host tool where the host has one
          (opencode: .opencode/tools/<id>.ts)
`)
}
