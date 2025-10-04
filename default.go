package main

import (
	"context"
	"flag"
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
		run: func(ctx context.Context, stdout, stderr io.Writer, prog string, args []string) error {
			a, err := parseDefault(stderr, prog, args)
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

func parseDefault(stderr io.Writer, prog string, args []string) (*DefaultArgs, error) {
	fs := flag.NewFlagSet(prog, flag.ContinueOnError)
	fs.SetOutput(stderr)
	root := ParseRootFlags(fs)
	fs.Usage = func() { Usage(stderr, prog) }
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
