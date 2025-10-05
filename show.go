package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/urfave/cli/v3"
)

func (cmd *cmd) Show() *cli.Command {
	return &cli.Command{
		Name:      "show",
		Usage:     "Show stored note for a commit/ref",
		Arguments: []cli.Argument{&cli.StringArg{Name: "ref", UsageText: "<git ref>"}},
		Action: func(ctx context.Context, c *cli.Command) error {
			notesRef := c.String("notes-ref")
			ref := c.StringArg("ref")

			if ref == "" {
				return fmt.Errorf("missing required argument: ref")
			}

			sha, err := gitResolveCommit(ctx, ref)
			if err != nil {
				return fmt.Errorf("resolving commit %s: %w", ref, err)
			}

			cmd.logger.DebugContext(ctx, "show start", "notes_ref", notesRef, "commit", sha)

			raw, err := gitNotesShow(ctx, notesRef, sha)
			if err != nil {
				return fmt.Errorf("reading note for commit %s: %w", sha, err)
			}

			var js map[string]any
			if err := json.Unmarshal(raw, &js); err != nil {
				return fmt.Errorf("reading note for commit %s: %w", sha, err)
			}
			// TODO: use a table format for non-JSON output?
			cmd.logger.InfoContext(ctx, "result", "commit", sha, "notes_ref", notesRef)
			return nil
		},
	}
}
