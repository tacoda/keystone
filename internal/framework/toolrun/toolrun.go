// Package toolrun runs a tool primitive's `run:` handler. Every transport
// goes through it: the keystone MCP server (transport: mcp), `keystone tool
// run` (transport: cli), and host plugins that call that command
// (transport: plugin).
package toolrun

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

// Run executes p's `run:` in projectDir with each argument exported as
// KEYSTONE_ARG_<NAME>, and returns the combined output. An argument p does
// not declare, or a required one that is missing, is an error.
func Run(ctx context.Context, projectDir string, p primitive.Primitive, args map[string]string) (string, error) {
	if err := Check(p, args); err != nil {
		return "", err
	}
	cmd := exec.CommandContext(ctx, "bash", "-c", p.Run)
	cmd.Dir = projectDir
	cmd.Env = append(os.Environ(), Env(p, args)...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	return buf.String(), err
}

// Env is the KEYSTONE_ARG_<NAME> variables for the declared arguments; an
// argument not given is exported empty.
func Env(p primitive.Primitive, args map[string]string) []string {
	var env []string
	for _, a := range p.Args {
		env = append(env, "KEYSTONE_ARG_"+strings.ToUpper(a.Name)+"="+args[a.Name])
	}
	return env
}

// missingRequired is the first required argument not given, "" if none.
func missingRequired(p primitive.Primitive, args map[string]string) string {
	for _, a := range p.Args {
		if a.Required && args[a.Name] == "" {
			return a.Name
		}
	}
	return ""
}

// Check rejects an undeclared argument and a missing required one.
func Check(p primitive.Primitive, args map[string]string) error {
	declared := map[string]bool{}
	for _, a := range p.Args {
		declared[a.Name] = true
	}
	if missing := missingRequired(p, args); missing != "" {
		return fmt.Errorf("tool %s: missing required argument --%s", p.ID, missing)
	}
	for name := range args {
		if !declared[name] {
			return fmt.Errorf("tool %s: unknown argument --%s", p.ID, name)
		}
	}
	return nil
}
