package main

import (
	"context"

	"github.com/urfave/cli/v3"
)

func (cmd *cmd) Sync() *cli.Command {
	return &cli.Command{
		Name:  "sync",
		Usage: "Sync benchmark notes with remote (push/fetch)",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "remote", Value: "origin", Usage: "git remote to sync with"},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cmd.logger.InfoContext(ctx, "[sync] (todo)", "remote", c.String("remote"))
			return nil
		},
	}
}
