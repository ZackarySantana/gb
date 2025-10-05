package main

import (
	"context"
	"fmt"
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