package main

import (
	"context"
	"fmt"
	"io"
)

func init() {
	defaultCMD = command{
		name: "",
		usages: []string{
			"\tRun benchmarks on working tree and compare vs HEAD note (default)",
		},
		examples: []string{
			"\tCompare working tree vs HEAD",
			"-bench '^BenchmarkFoo$'\tBenchmark one test",
		},
		run: func(ctx context.Context, stdout, stderr io.Writer, args []string) error {
			a, err := parseDefault(ctx, stderr, args)
			if err != nil {
				return err
			}
			return Default(ctx, a, stdout, stderr)
		},
	}
	cmds = append([]command{defaultCMD}, cmds...) // Default command is first
}

type DefaultArgs struct {
	Root *RootFlags
}

func parseDefault(ctx context.Context, stderr io.Writer, args []string) (*DefaultArgs, error) {
	fs, root := setupCommandFlags(ctx, "", stderr)
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return &DefaultArgs{Root: root}, nil
}

func Default(ctx context.Context, a *DefaultArgs, stdout, stderr io.Writer) error {
	fmt.Fprintln(stdout, "[default] run benches on working tree; compare vs HEAD (todo)")
	// TODO: This should just call Compare?
	return nil
}
