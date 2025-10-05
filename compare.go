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
		run: func(ctx context.Context, logger *slog.Logger, args []string) error {
			a, err := parseCompare(ctx, logger, args)
			if err != nil {
				return err
			}
			return Compare(ctx, a, logger)
		},
	})
}

const cmdCompare = "compare"

type CompareArgs struct {
	Root *RootFlags
	Base string
	Head string
}

func parseCompare(ctx context.Context, logger *slog.Logger, args []string) (*CompareArgs, error) {
	fs, root := setupFlags(ctx, logger)
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	base, head := "origin/main", "HEAD"
	if filled := optionalArgs(fs.Args(), &base, &head); !filled {
		logger.DebugContext(ctx, "using some defaults", "base", base, "head", head)
	}
	return &CompareArgs{Root: root, Base: base, Head: head}, nil
}

func Compare(ctx context.Context, a *CompareArgs, logger *slog.Logger) error {
	logger.InfoContext(ctx, "[compare] base=%s head=%s (todo)", a.Base, a.Head)
	return nil
}
