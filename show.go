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
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "all", Aliases: []string{"a"}, Usage: "show all notes (not just the one targetted by 'notes-ref')"},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			notesRef := c.String("notes-ref")
			ref := c.StringArg("ref")
			all := c.Bool("all")

			if ref == "" {
				return fmt.Errorf("missing required argument: ref")
			}

			sha, err := gitResolveCommit(ctx, ref)
			if err != nil {
				return fmt.Errorf("resolving commit %s: %w", ref, err)
			}

			if !all {
				return cmd.showNote(ctx, notesRef, sha)
			}

			// If it's all, we
			notes, err := listAllNotesRefs(ctx, sha)
			if err != nil {
				return fmt.Errorf("listing all notes: %w", err)
			}

			for ref, raw := range notes {
				var note Note
				if err := json.Unmarshal(raw, &note); err != nil {
					return fmt.Errorf("reading note for commit %s ref %s: %w", sha, ref, err)
				}
				cmd.logger.InfoContext(ctx, "result", "commit", sha, "notes_ref", ref)
				cmd.logNote(ctx, &note)
			}

			return nil
		},
	}
}

func (cmd *cmd) showNote(ctx context.Context, notesRef, sha string) error {
	cmd.logger.DebugContext(ctx, "show start", "notes_ref", notesRef, "commit", sha)

	note, err := loadNote(ctx, notesRef, sha)
	if err != nil {
		return err
	}

	cmd.logger.InfoContext(ctx, "result", "commit", sha, "notes_ref", notesRef)
	cmd.logNote(ctx, note)
	return nil
}

func (cmd *cmd) logNote(ctx context.Context, note *Note) {
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
