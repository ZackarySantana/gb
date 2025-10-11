package main

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log/slog"
	"runtime"
	"strings"

	"github.com/urfave/cli/v3"
)

func (cmd *cmd) Root(logLevel *slog.LevelVar) *cli.Command {
	gitEmail, err := gitEmail(context.Background())
	if err != nil {
		cmd.logger.Error("unable to get git user email to generate a unique note", "error", err)
	}
	sum := sha1.Sum([]byte(gitEmail))
	gv := runtime.Version()
	notesRef := fmt.Sprintf("%s/%s-%s-%s", hex.EncodeToString(sum[:8]), runtime.GOOS, runtime.GOARCH, gv)
	return &cli.Command{
		Name:  "gb",
		Usage: "Go benchmark notes manager",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "v", Aliases: []string{"verbose"}},
			&cli.IntFlag{Name: "count", Value: 10, Usage: "benchmark count"},
			&cli.StringFlag{Name: "benchtime", Usage: "benchtime duration (e.g. 2s)"},
			&cli.StringFlag{Name: "bench", Value: ".", Usage: "benchmark regex"},
			&cli.StringFlag{Name: "pkgs", Value: "./...", Usage: "comma-separated package list"},
			&cli.StringFlag{Name: "notes-ref", Value: notesRef, Usage: "override notes ref (will always be prefixed with refs/notes/gb/)"},
		},
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			if c.Bool("v") {
				logLevel.Set(slog.LevelDebug)
			}
			// update notes-ref to be a full ref
			return ctx, nil
		},
		After: func(ctx context.Context, c *cli.Command) error {
			cmd.logger.DebugContext(ctx, "command complete, cleaning up worktrees")
			out, err := gitWorktreeList(ctx)
			if err != nil {
				cmd.logger.ErrorContext(ctx, "cleaning up leftover worktrees", "error", err)
				return nil
			}
			for _, l := range out {
				if !strings.HasPrefix(l.path, "/tmp/gb-wt-") {
					continue
				}
				cmd.logger.InfoContext(ctx, "removing leftover worktree", "path", l.path, "commit", l.commit)
				if err = gitWorktreeRemove(context.WithoutCancel(ctx), l.path); err != nil {
					cmd.logger.ErrorContext(ctx, "removing leftover worktree", "error", err, "path", l.path)
				}
			}
			cmd.logger.DebugContext(ctx, "worktree cleanup complete")
			return nil
		},
		Commands: []*cli.Command{
			cmd.Backfill(),
			cmd.Compare(),
			cmd.Export(),
			cmd.Show(),
			cmd.Sync(),
		},
	}
}

func getNotesRef(c *cli.Command) string {
	ref := c.String("notes-ref")
	if !strings.HasPrefix(ref, "refs/notes/gb/") {
		ref = "refs/notes/gb/" + ref
	}
	return ref
}
