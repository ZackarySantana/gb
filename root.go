package main

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"

	"github.com/urfave/cli/v3"
)

func (cmd *cmd) Root(logLevel *slog.LevelVar) *cli.Command {
	gv := runtime.Version()
	notesRef := fmt.Sprintf("refs/notes/benches/%s-%s-%s", runtime.GOOS, runtime.GOARCH, gv)
	return &cli.Command{
		Name:  "gb",
		Usage: "Go benchmark notes manager",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "v", Aliases: []string{"verbose"}},
			&cli.IntFlag{Name: "count", Value: 10, Usage: "benchmark count"},
			&cli.StringFlag{Name: "benchtime", Usage: "benchtime duration (e.g. 2s)"},
			&cli.StringFlag{Name: "bench", Value: ".", Usage: "benchmark regex"},
			&cli.StringFlag{Name: "pkgs", Value: "./...", Usage: "comma-separated package list"},
			&cli.StringFlag{Name: "notes-ref", Value: notesRef, Usage: "override notes ref (default derived from env)"},
			&cli.BoolFlag{Name: "force", Usage: "allow cross-environment comparisons"},
		},
		Action: cmd.Default,
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			if c.Bool("v") {
				logLevel.Set(slog.LevelDebug)
			}
			return ctx, nil
		},
		Commands: []*cli.Command{
			cmd.Backfill(),
			cmd.Compare(),
			cmd.Show(),
			cmd.Sync(),
		},
	}
}

func (cmd *cmd) Default(ctx context.Context, c *cli.Command) error {
	for _, flag := range c.Flags {
		slog.Default().Info("flag", "name", flag.Names()[0], "value", fmt.Sprintf("%v", flag.Get()))
	}
	return nil
}
