package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// runCmd executes a command in the given directory and returns the combined output.
func runCmd(ctx context.Context, dir string, bin string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, bin, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, fmt.Errorf("%s %s: %w", bin, strings.Join(args, " "), err)
	}
	return out, nil
}

// short returns the first 8 characters of a SHA string, or the full string if shorter.
func short(sha string) string {
	if len(sha) > 8 {
		return sha[:8]
	}
	return sha
}

// setupCommandFlags creates and configures a flag.FlagSet for a command.
// It returns the flagset and root flags, with common setup already done.
func setupCommandFlags(ctx context.Context, cmdName string, stderr io.Writer) (*flag.FlagSet, *RootFlags) {
	fs := flag.NewFlagSet(cmdName, flag.ContinueOnError)
	fs.SetOutput(stderr)
	root := ParseRootFlags(fs)
	fs.Usage = func() { Usage(ctx, stderr) }
	return fs, root
}

// requireSingleArg validates that exactly one positional argument is provided.
// Returns the argument or an error with appropriate message.
func requireSingleArg(fs *flag.FlagSet, stderr io.Writer, cmdName, argName string) (string, error) {
	args := fs.Args()
	if len(args) < 1 {
		fmt.Fprintf(stderr, "%s: missing %s\n", cmdName, argName)
		return "", flag.ErrHelp
	}
	return args[0], nil
}