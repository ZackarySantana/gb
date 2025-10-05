package main

import (
	"context"
	"fmt"
	"log/slog"
)

func init() {
	cmds = append(cmds, command{
		name: cmdCompare,
		usages: []string{
			fmt.Sprintf("%s [base [head]]\tCompare stored notes for two refs", cmdCompare),
		},
		examples: []string{
			fmt.Sprintf("%s origin/main HEAD\tCompare two commits", cmdCompare),
		},
		run: func(ctx context.Context, params *commandParams) error {
			a, err := parseCompare(ctx, params)
			if err != nil {
				return err
			}
			return Compare(ctx, a, params.logger)
		},
	})
}

const cmdCompare = "compare"

type CompareArgs struct {
	Root *RootFlags
	Base string
	Head string
}

func parseCompare(ctx context.Context, params *commandParams) (*CompareArgs, error) {
	if err := params.fs.Parse(params.args); err != nil {
		return nil, err
	}
	base, head := "origin/main", "HEAD"
	if filled := optionalArgs(params.fs.Args(), &base, &head); !filled {
		params.logger.DebugContext(ctx, "using some defaults", "base", base, "head", head)
	}
	return &CompareArgs{Root: params.root, Base: base, Head: head}, nil
}

func Compare(ctx context.Context, a *CompareArgs, logger *slog.Logger) error {
	logger.InfoContext(ctx, "[compare] base=%s head=%s (todo)", a.Base, a.Head)
	return nil
}
