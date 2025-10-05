package main

import (
	"context"
	"log/slog"
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
		run: func(ctx context.Context, params *commandParams) error {
			a, err := parseDefault(ctx, params)
			if err != nil {
				return err
			}
			return Default(ctx, a, params.logger)
		},
	}
	cmds = append([]command{defaultCMD}, cmds...) // Default command is first
}

type DefaultArgs struct {
	Root *RootFlags
}

func parseDefault(_ context.Context, params *commandParams) (*DefaultArgs, error) {
	return &DefaultArgs{Root: params.root}, nil
}

func Default(ctx context.Context, a *DefaultArgs, logger *slog.Logger) error {
	logger.InfoContext(ctx, "[default] run benches on working tree; compare vs HEAD (todo)")
	// TODO: This should just call Compare?
	return nil
}
