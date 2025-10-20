package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/urfave/cli/v3"
)

func (cmd *cmd) Backfill() *cli.Command {
	return &cli.Command{
		Name:  "backfill",
		Usage: "Backfill benchmark notes with missing commits since a ref",
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "since", UsageText: "<git ref>"},
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "force", Aliases: []string{"f"}, Usage: "re-benchmark even if note(s) exists"},
			&cli.BoolFlag{Name: "single", Usage: "single commit to benchmark (bypasses range)"},
			&cli.BoolFlag{Name: "fast-abort", Usage: "abort backfill on first benchmark failure"},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			since := c.StringArg("since")
			force := c.Bool("force")
			single := c.Bool("single")
			fastAbort := c.Bool("fast-abort")

			if since == "" {
				return fmt.Errorf("missing required argument: since")
			}

			rangeSpec := since + "^..HEAD"
			if single {
				rangeSpec = since + "^.." + since
			}

			notesRef := getNotesRef(c)

			cmd.logger.InfoContext(ctx, "backfill start", "notes_ref", notesRef, "range", rangeSpec)

			commits, err := gitRevList(ctx, rangeSpec)
			if err != nil {
				return fmt.Errorf("git rev-list: %w", err)
			}
			if len(commits) == 0 {
				cmd.logger.InfoContext(ctx, "no commits to backfill")
				return nil
			}
			benchmarkArgs := benchmarkCommand(c.String("pkgs"), c.String("bench"), c.String("benchtime"), c.Int("count"))

			cmd.logger.DebugContext(ctx, "found commits", "benchmark command", strings.Join(benchmarkArgs, " "), "commits", strings.Join(commits, ", "))

			var done, skipped int
			var failed []error
			start := time.Now()
			for _, commit := range commits {
				isSkipped, err := cmd.benchmark(ctx, notesRef, commit, benchmarkArgs, force)
				if err != nil {
					failed = append(failed, fmt.Errorf("commit %s: %w", commit, err))
					if fastAbort {
						break
					}
					continue
				}
				if isSkipped {
					skipped++
					continue
				}
				done++
			}

			elapsed := time.Since(start).Truncate(time.Millisecond)
			cmd.logger.InfoContext(ctx, "backfill complete", "noted", done, "skipped", skipped, "failed", len(failed), "elapsed", elapsed)
			if len(failed) > 0 {
				return fmt.Errorf("some benchmarks failed: %v", failed)
			}
			return nil
		},
	}
}

func (cmd *cmd) benchmark(ctx context.Context, notesRef, commit string, benchmarkArgs []string, force bool) (skipped bool, failed error) {
	_, err := gitNotesShow(ctx, notesRef, commit)
	if err != nil && !errors.Is(err, errNoteMissing) {
		return false, fmt.Errorf("checking note for commit %s: %w", commit, err)
	}
	if !errors.Is(err, errNoteMissing) {
		if !force {
			cmd.logger.DebugContext(ctx, "note exists, skipping", "commit", commit)
			return true, nil
		}
		cmd.logger.DebugContext(ctx, "note exists but force enabled, re-benchmarking", "commit", commit)
	}

	cmd.logger.DebugContext(ctx, "benchmarking", "commit", commit)

	raw, err := gitWorktreeRunCommand(ctx, commit, benchmarkArgs)
	if err != nil {
		return false, fmt.Errorf("benchmark failed for commit %s: %w", commit, err)
	}

	payload, err := marshalNotePayload(commit, benchmarkArgs, raw)
	if err != nil {
		return false, fmt.Errorf("marshal note failed for commit %s: %w", commit, err)
	}

	if err := gitNotesAdd(ctx, notesRef, commit, payload); err != nil {
		return false, fmt.Errorf("git notes add failed for commit %s: %w", commit, err)
	}

	cmd.logger.DebugContext(ctx, "noted", "commit", commit)
	return false, nil
}

func benchmarkCommand(pkgs, bench, benchtime string, count int) []string {
	if strings.TrimSpace(pkgs) == "" {
		pkgs = "./..."
	}
	args := []string{"go", "test", pkgs, "-run=^$", "-bench", bench, "-benchmem"}
	if count > 0 {
		args = append(args, "-count", fmt.Sprint(count))
	}
	if strings.TrimSpace(benchtime) != "" {
		args = append(args, "-benchtime", benchtime)
	}
	return args
}
