package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
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
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			notesRef := c.String("notes-ref")
			since := c.StringArg("since")
			force := c.Bool("force")

			cmd.logger.DebugContext(ctx, "backfill start", "notes_ref", notesRef, "range", since+"^..HEAD")

			commits, err := gitRevList(ctx, since+"^..HEAD")
			if err != nil {
				return fmt.Errorf("git rev-list: %w", err)
			}
			if len(commits) == 0 {
				cmd.logger.InfoContext(ctx, "no commits to backfill")
				return nil
			}
			benchmarkArgs := benchmarkCommand(c)

			cmd.logger.DebugContext(ctx, "found commits", "benchmark command", strings.Join(benchmarkArgs, " "), "commits", strings.Join(commits, ", "))

			var done, skipped, failed int
			start := time.Now()
			for _, commit := range commits {
				isSkipped, err := cmd.benchmark(ctx, notesRef, commit, benchmarkArgs, force)
				if err != nil {
					cmd.logger.ErrorContext(ctx, "benchmark failed", "error", err, "commit", commit)
					failed++
					continue
				}
				if isSkipped {
					skipped++
					continue
				}
				done++
			}

			elapsed := time.Since(start).Truncate(time.Millisecond)
			cmd.logger.InfoContext(ctx, "backfill complete", "noted", done, "skipped", skipped, "failed", failed, "elapsed", elapsed)
			if failed > 0 {
				return errors.New("some commits failed to backfill")
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

func benchmarkCommand(c *cli.Command) []string {
	pkgs := c.String("pkgs")
	if strings.TrimSpace(pkgs) == "" {
		pkgs = "./..."
	}
	args := []string{"go", "test", pkgs, "-run=^$", "-bench", c.String("bench"), "-benchmem"}
	if c.Int("count") > 0 {
		args = append(args, "-count", fmt.Sprint(c.Int("count")))
	}
	if strings.TrimSpace(c.String("benchtime")) != "" {
		args = append(args, "-benchtime", c.String("benchtime"))
	}
	return args
}

func marshalNotePayload(commit string, benchArgs []string, raw []byte) ([]byte, error) {
	host, _ := os.Hostname()
	doc := map[string]any{
		"schema":     1,
		"commit":     commit,
		"created_at": time.Now().UTC().Format(time.RFC3339),
		"env": map[string]any{
			"go_version": runtime.Version(),
			"goos":       runtime.GOOS,
			"goarch":     runtime.GOARCH,
			"gomaxprocs": runtime.GOMAXPROCS(0),
			"host":       host,
			"cpus":       runtime.NumCPU(),
		},
		"bench_args": benchArgs,
		"raw":        string(raw),
	}
	return json.Marshal(doc)
}
