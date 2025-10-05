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
			base, head := "origin/main", "HEAD"
			if filled := optionalArgs(params.fs.Args(), &base, &head); !filled {
				params.logger.DebugContext(ctx, "using some defaults", "base", base, "head", head)
			}
			return Compare(ctx, &CompareArgs{Root: params.root, Base: base, Head: head}, params.logger)
		},
	})
}

const cmdCompare = "compare"

type CompareArgs struct {
	Root *RootFlags
	Base string
	Head string
}

func Compare(ctx context.Context, a *CompareArgs, logger *slog.Logger) error {
	logger.InfoContext(ctx, "[compare] base=%s head=%s (todo)", a.Base, a.Head)
	return nil
}
