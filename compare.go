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
			&cli.StringArg{Name: "base", UsageText: "<git ref>", Value: "HEAD~1"},
			&cli.StringArg{Name: "head", UsageText: "<git ref>", Value: "HEAD"},
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "create", Aliases: []string{"c"}, Usage: "creates notes if missing (runs benchmarks)"},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			notesRef := c.String("notes-ref")
			baseRef := c.StringArg("base")
			headRef := c.StringArg("head")
			create := c.Bool("create")

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

			baseNote, err := loadNote(ctx, notesRef, baseSHA)
			if err != nil {
				if !errors.Is(err, errNoteMissing) || !create {
					return fmt.Errorf("load base note: %w", err)
				}

				cmd.logger.InfoContext(ctx, "base note missing, creating", "commit", baseSHA)
				benchmarkArgs := benchmarkCommand(c.String("pkgs"), c.String("bench"), c.String("benchtime"), c.Int("count"))
				if _, err = cmd.benchmark(ctx, notesRef, baseSHA, benchmarkArgs, false); err != nil {
					return fmt.Errorf("create base note: %w", err)
				}
				baseNote, err = loadNote(ctx, notesRef, baseSHA)
				if err != nil {
					return fmt.Errorf("load created base note: %w", err)
				}
			}
			headNote, err := loadNote(ctx, notesRef, headSHA)
			if err != nil {
				if !errors.Is(err, errNoteMissing) || !create {
					return fmt.Errorf("load head note: %w", err)
				}

				cmd.logger.InfoContext(ctx, "head note missing, creating", "commit", headSHA)
				benchmarkArgs := benchmarkCommand(c.String("pkgs"), c.String("bench"), c.String("benchtime"), c.Int("count"))
				if _, err = cmd.benchmark(ctx, notesRef, headSHA, benchmarkArgs, false); err != nil {
					return fmt.Errorf("create head note: %w", err)
				}
				baseNote, err = loadNote(ctx, notesRef, headSHA)
				if err != nil {
					return fmt.Errorf("load created head note: %w", err)
				}
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
