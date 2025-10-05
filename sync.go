package main

import (
	"context"
	"fmt"
	"log/slog"
)

func init() {
	cmds = append(cmds, command{
		name: cmdSync,
		usages: []string{
			fmt.Sprintf("%s --remote NAME\tSync benchmark notes with remote (push/fetch)\n", cmdSync),
		},
		run: func(ctx context.Context, logger *slog.Logger, args []string) error {
			a, err := parseSync(ctx, logger, args)
			if err != nil {
				return err
			}
			return Sync(ctx, a, logger)
		},
	})
}

const cmdSync = "sync"

type SyncArgs struct {
	Root   *RootFlags
	Remote string
}

func parseSync(ctx context.Context, logger *slog.Logger, args []string) (*SyncArgs, error) {
	fs, root := setupFlags(ctx, logger)
	remote := fs.String("remote", "origin", "git remote to sync with")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return &SyncArgs{Root: root, Remote: *remote}, nil
}

func Sync(ctx context.Context, a *SyncArgs, logger *slog.Logger) error {
	logger.InfoContext(ctx, "[sync] (todo)", "remote", a.Remote)
	return nil
}
