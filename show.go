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
			notesRef := getNotesRef(c)
			ref := c.StringArg("ref")
			all := c.Bool("all")

			if ref == "" {
				return fmt.Errorf("missing required argument: ref")
			}

			sha, err := gitResolveCommit(ctx, ref)
			if err != nil {
				return fmt.Errorf("resolving commit %s: %w", ref, err)
			}

			cmd.logger.InfoContext(ctx, "show start", "notes_ref", notesRef, "commit", sha)

			if !all {
				note, err := loadNote(ctx, notesRef, sha)
				if err != nil {
					return err
				}

				cmd.logNote(ctx, note, notesRef)
				return nil
			}

			// If it's all, we
			notes, err := listAllNotesRefs(ctx, sha)
			if err != nil {
				return fmt.Errorf("listing all notes: %w", err)
			}

			if len(notes) == 0 {
				return fmt.Errorf("no notes found for commit %s", sha)
			}

			for ref, raw := range notes {
				var note Note
				if err := json.Unmarshal(raw, &note); err != nil {
					cmd.logger.WarnContext(ctx, "unknown note format", "notes_ref", ref, "error", err)
				} else {
					cmd.logNote(ctx, &note, ref)
				}
			}

			return nil
		},
	}
}

func (cmd *cmd) logNote(ctx context.Context, note *Note, notesRef string) {
	for _, bench := range note.Parsed.Benches {
		stats := bench.Stats
		cmd.logger.InfoContext(ctx, "bench result",
			"benchmark", bench.Name,
			"notes_ref", notesRef,
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
