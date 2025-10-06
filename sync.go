package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"
)

func (cmd *cmd) Sync() *cli.Command {
	return &cli.Command{
		Name:  "sync",
		Usage: "Sync benchmark notes with remote (push/fetch)",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "remote", Value: "origin", Usage: "git remote to sync with"},
			&cli.BoolFlag{Name: "force", Aliases: []string{"f"}, Usage: "force overwrite remote refs"},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			remote := c.String("remote")
			force := c.Bool("force")

			cmd.logger.InfoContext(ctx, "sync start", "remote", remote, "force", force)

			cmd.logger.InfoContext(ctx, "pushing", "remote", remote, "force", force)
			refs, err := listAllNotesRefs(ctx)
			if err != nil {
				return fmt.Errorf("listing notes refs: %w", err)
			}
			if len(refs) == 0 {
				cmd.logger.InfoContext(ctx, "no notes refs found, nothing to push")
			} else {
				group, ctx := errgroup.WithContext(ctx)
				for _, ref := range refs {
					group.Go(func() error {
						args := []string{"push"}
						if force {
							args = append(args, "--force")
						}
						args = append(args, remote, ref)
						cmd.logger.DebugContext(ctx, "git", "args", strings.Join(args, " "))
						if out, err := runCmd(ctx, "", "git", args...); err != nil {
							return fmt.Errorf("git push %s: %v: %s", ref, err, string(out))
						}
						return nil
					})
				}

				if err = group.Wait(); err != nil {
					return err
				}
			}

			cmd.logger.InfoContext(ctx, "fetching", "remote", remote, "force", force)
			args := []string{"fetch", remote, "+refs/notes/gb/*:refs/notes/gb/*"}
			cmd.logger.DebugContext(ctx, "git", "args", strings.Join(args, " "))
			if out, err := runCmd(ctx, "", "git", args...); err != nil {
				return fmt.Errorf("git fetch notes: %v: %s", err, string(out))
			}

			cmd.logger.InfoContext(ctx, "sync complete", "remote", remote)

			return nil
		},
	}
}
