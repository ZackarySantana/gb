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
		flags: func(ctx context.Context, params *commandParams) {
			remote := params.fs.String("remote", "origin", "git remote to sync with")
			params.flags["remote"] = remote
		},
		run: func(ctx context.Context, params *commandParams) error {
			a, err := parseSync(ctx, params)
			if err != nil {
				return err
			}
			return Sync(ctx, a, params.logger)
		},
	})
}

const cmdSync = "sync"

type SyncArgs struct {
	Root   *RootFlags
	Remote string
}

func parseSync(_ context.Context, params *commandParams) (*SyncArgs, error) {
	return &SyncArgs{Root: params.root, Remote: *params.flags["remote"]}, nil
}

func Sync(ctx context.Context, a *SyncArgs, logger *slog.Logger) error {
	logger.InfoContext(ctx, "[sync] (todo)", "remote", a.Remote)
	return nil
}
