package main

import (
	"context"
	"errors"
	"fmt"

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
			notesRef := c.String("notes-ref")
			baseRef := c.StringArg("base")
			headRef := c.StringArg("head")

			if baseRef == "" || headRef == "" {
				return errors.New("missing required arguments: base and/or head")
			}

			baseSHA, err := gitResolveCommit(ctx, baseRef)
			if err != nil {
				return fmt.Errorf("resolve base %s: %w", baseRef, err)
			}
			headSHA, err := gitResolveCommit(ctx, headRef)
			if err != nil {
				return fmt.Errorf("resolve head %s: %w", headRef, err)
			}

			cmd.logger.InfoContext(ctx, "compare start",
				"notes_ref", notesRef, "base_ref", baseRef, "base", baseSHA, "head_ref", headRef, "head", headSHA,
			)

			// TODO: There should be a flag to just load benchmark + save the notes, and then compare.
			baseNote, err := loadNote(ctx, notesRef, baseSHA)
			if err != nil {
				return fmt.Errorf("load base note: %w", err)
			}
			headNote, err := loadNote(ctx, notesRef, headSHA)
			if err != nil {
				return fmt.Errorf("load head note: %w", err)
			}

			baseBenches := benchesByName(baseNote.Parsed.Benches)
			headBenches := benchesByName(headNote.Parsed.Benches)

			for name, headBench := range headBenches {
				baseBench, ok := baseBenches[name]
				if !ok {
					cmd.logger.InfoContext(ctx, "new benchmark", "benchmark", name)
					continue
				}
				if !c.Bool("verbose") {
					cmd.logger.InfoContext(ctx, "benchmark comparison", "benchmark", name,
						"ns_per_op_percentage", (headBench.Stats.NsPerOpMean-baseBench.Stats.NsPerOpMean)/baseBench.Stats.NsPerOpMean*100.0,

						"bytes_per_op_percentage", (headBench.Stats.BytesPerOpMean-baseBench.Stats.BytesPerOpMean)/baseBench.Stats.BytesPerOpMean*100.0,

						"allocs_per_op_percentage", (headBench.Stats.AllocsPerOpMean-baseBench.Stats.AllocsPerOpMean)/baseBench.Stats.AllocsPerOpMean*100.0,
					)
					continue
				}

				cmd.logger.DebugContext(ctx, "benchmark comparison", "benchmark", name,
					"ns_per_op_base", baseBench.Stats.NsPerOpMean,
					"ns_per_op_head", headBench.Stats.NsPerOpMean,
					"ns_per_op_diff", headBench.Stats.NsPerOpMean-baseBench.Stats.NsPerOpMean,
					"ns_per_op_percentage", (headBench.Stats.NsPerOpMean-baseBench.Stats.NsPerOpMean)/baseBench.Stats.NsPerOpMean*100.0,

					"bytes_per_op_base", baseBench.Stats.BytesPerOpMean,
					"bytes_per_op_head", headBench.Stats.BytesPerOpMean,
					"bytes_per_op_diff", headBench.Stats.BytesPerOpMean-baseBench.Stats.BytesPerOpMean,
					"bytes_per_op_percentage", (headBench.Stats.BytesPerOpMean-baseBench.Stats.BytesPerOpMean)/baseBench.Stats.BytesPerOpMean*100.0,

					"allocs_per_op_base", baseBench.Stats.AllocsPerOpMean,
					"allocs_per_op_head", headBench.Stats.AllocsPerOpMean,
					"allocs_per_op_diff", headBench.Stats.AllocsPerOpMean-baseBench.Stats.AllocsPerOpMean,
					"allocs_per_op_percentage", (headBench.Stats.AllocsPerOpMean-baseBench.Stats.AllocsPerOpMean)/baseBench.Stats.AllocsPerOpMean*100.0,

					"count_base", baseBench.Stats.Count,
					"count_head", headBench.Stats.Count,
				)
			}

			for name := range baseBenches {
				if _, ok := headBenches[name]; !ok {
					cmd.logger.InfoContext(ctx, "removed benchmark", "benchmark", name)
				}
			}

			return nil
		},
	}
}

func benchesByName(benches []BenchCase) map[string]BenchCase {
	m := make(map[string]BenchCase, len(benches))
	for _, b := range benches {
		m[b.Name] = b
	}
	return m
}
