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

			note, err := loadNote(ctx, notesRef, sha)
			if err != nil {
				return fmt.Errorf("reading note for commit %s: %w", sha, err)
			}

			cmd.logger.InfoContext(ctx, "result", "commit", sha, "notes_ref", notesRef)
			for _, bench := range note.Parsed.Benches {
				stats := bench.Stats
				cmd.logger.InfoContext(ctx, "bench result",
					"benchmark", bench.Name,
					"ns_per_op_mean", stats.NsPerOpMean,
					"ns_per_op_median", stats.NsPerOpMedian,
					"bytes_per_op_mean", stats.BytesPerOpMean,
					"allocs_per_op_mean", stats.AllocsPerOpMean,
					"samples", stats.Count,
				)
			}
			return nil
		},
	}
}

func loadNote(ctx context.Context, notesRef, sha string) (*Note, error) {
	raw, err := gitNotesShow(ctx, notesRef, sha)
	if err != nil {
		return nil, fmt.Errorf("reading note for commit %s: %w", sha, err)
	}

	var note Note
	if err := json.Unmarshal(raw, &note); err != nil {
		return nil, fmt.Errorf("reading note for commit %s: %w", sha, err)
	}
	return &note, nil
}
