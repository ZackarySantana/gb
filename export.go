package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"
)

func (cmd *cmd) Export() *cli.Command {
	return &cli.Command{
		Name:  "export",
		Usage: "Export benchmark notes to a file",
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "since", UsageText: "<git ref>"},
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "single", Usage: "single commit to export (bypasses range)"},
			&cli.StringFlag{Name: "output", Value: "benchmarks", Usage: "output file path"},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			since := c.StringArg("since")
			single := c.Bool("single")
			output := c.String("output")

			if since == "" {
				return fmt.Errorf("missing required argument: since")
			}

			rangeSpec := since + "^..HEAD"
			if single {
				rangeSpec = since + "^.." + since
			}

			commits, err := gitRevList(ctx, rangeSpec)
			if err != nil {
				return fmt.Errorf("git rev-list: %w", err)
			}
			if len(commits) == 0 {
				cmd.logger.InfoContext(ctx, "no commits found in range")
				return nil
			}

			cmd.logger.InfoContext(ctx, "sync start", "range", rangeSpec)

			refs, err := listAllNotesRefs(ctx)
			if err != nil {
				return fmt.Errorf("listing notes refs: %w", err)
			}
			if len(refs) == 0 {
				cmd.logger.InfoContext(ctx, "no notes refs found, nothing to export")
				return nil
			}
			type note struct {
				fileName string
				value    []byte
			}
			ch := make(chan note, 10)
			group, errCtx := errgroup.WithContext(ctx)

			group.Go(func() error {
				for {
					select {
					case <-errCtx.Done():
						return errCtx.Err()
					case n, ok := <-ch:
						if !ok {
							return errCtx.Err()
						}
						if err := os.MkdirAll(filepath.Dir(n.fileName), 0o755); err != nil {
							return fmt.Errorf("mkdir at %s: %w", filepath.Dir(n.fileName), err)
						}
						if err := os.WriteFile(n.fileName, n.value, 0o644); err != nil {
							return fmt.Errorf("writing note at %s: %w", n.fileName, err)
						}
						cmd.logger.DebugContext(errCtx, "wrote note", "fileName", n.fileName)
					}
				}
			})

			commitsWithoutNotes := 0
			notesWrote := 0
			for _, commit := range commits {
				hasNote := false
				for _, ref := range refs {
					value, err := gitNotesShow(ctx, ref, commit)
					if err != nil {
						// This commit doesn't have this note ref, so we skip it.
						continue
					}
					hasNote = true
					select {
					case <-ctx.Done():
						close(ch)
						group.Wait()
						return ctx.Err()
					case ch <- note{
						fileName: fmt.Sprintf("%s/%s/%s.json", output, commit, ref),
						value:    value,
					}:
						notesWrote++
					}
				}
				if !hasNote {
					commitsWithoutNotes++
				}
			}

			close(ch)
			if err := group.Wait(); err != nil {
				return fmt.Errorf("exporting notes: %w", err)
			}

			cmd.logger.InfoContext(ctx, "export complete", "range", rangeSpec, "commits", len(commits), "notes", notesWrote, "commitsWithoutNotes", commitsWithoutNotes)
			return nil
		},
	}
}
