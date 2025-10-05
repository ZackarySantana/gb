package main

import (
	"context"

	"github.com/urfave/cli/v3"
)

func (cmd *cmd) Compare() *cli.Command {
	return &cli.Command{
		Name:  "compare",
		Usage: "Compare stored notes for two refs",
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "base", UsageText: "<git ref>", Value: "origin/main"},
			&cli.StringArg{Name: "head", UsageText: "<git ref>", Value: "HEAD"},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			cmd.logger.InfoContext(ctx, "[compare] base=%s head=%s (todo)", c.String("base"), c.String("head"))
			return nil
		},
	}
}
