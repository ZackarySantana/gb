package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
			&cli.BoolFlag{Name: "force", Usage: "re-benchmark even if note(s) exists"},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			notesRef := c.String("notes-ref")
			since := c.StringArg("since")
			force := c.Bool("force")

			cmd.logger.DebugContext(ctx, "backfill start", "notes_ref", notesRef, "range", since+"..HEAD")

			commits, err := gitRevList(ctx, since+"..HEAD")
			if err != nil {
				return fmt.Errorf("git rev-list: %w", err)
			}
			if len(commits) == 0 {
				cmd.logger.InfoContext(ctx, "no commits to backfill")
				return nil
			}

			var done, skipped, failed int
			start := time.Now()

			for _, commit := range commits {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				has, err := gitNoteExists(ctx, notesRef, commit)
				if err != nil {
					failed++
					cmd.logger.ErrorContext(ctx, "git note exists failed", "commit", commit, "error", err)
					continue
				}
				if has {
					if !force {
						skipped++
						cmd.logger.DebugContext(ctx, "note exists, skipping", "commit", commit)
						continue
					}
					cmd.logger.DebugContext(ctx, "note exists but force enabled, re-benchmarking", "commit", commit)
				}

				cmd.logger.DebugContext(ctx, "benchmarking", "commit", commit)

				raw, benchArgs, err := runBenchesInWorktree(ctx, commit, c)
				if err != nil {
					failed++
					cmd.logger.ErrorContext(ctx, "benchmark failed", "commit", commit, "error", err)
					continue
				}

				payload, err := marshalNotePayload(commit, benchArgs, raw)
				if err != nil {
					failed++
					cmd.logger.ErrorContext(ctx, "marshal note failed", "commit", commit, "error", err)
					continue
				}

				if err := gitNotesAdd(ctx, notesRef, commit, payload); err != nil {
					failed++
					cmd.logger.ErrorContext(ctx, "git notes add failed", "commit", commit, "error", err)
					continue
				}

				done++
				cmd.logger.DebugContext(ctx, "noted", "commit", commit)
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

/* ------------------------------- helpers ---------------------------------- */

func gitRevList(ctx context.Context, rangeSpec string) ([]string, error) {
	out, err := runCmd(ctx, "", "git", "rev-list", "--reverse", rangeSpec)
	if err != nil {
		return nil, err
	}
	lines := strings.Fields(string(out))
	return lines, nil
}

func gitNoteExists(ctx context.Context, notesRef, commit string) (bool, error) {
	_, err := runCmd(ctx, "", "git", "notes", "--ref", notesRef, "show", commit)
	if err == nil {
		return true, nil
	}
	// exit code != 0 when note missing; differentiate from other errors
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return false, nil
	}
	return false, err
}

func gitNotesAdd(ctx context.Context, notesRef, commit string, payload []byte) error {
	// We use -f to overwrite if a concurrent run added one; normally it won't exist.
	_, err := runCmd(ctx, "", "git", "notes", "--ref", notesRef, "add", "-f", "-m", string(payload), commit)
	return err
}

func runBenchesInWorktree(ctx context.Context, commit string, c *cli.Command) ([]byte, []string, error) {
	// create temp worktree
	tmp := filepath.Join(os.TempDir(), "gb-wt-"+commit[:8]+"-"+fmt.Sprint(time.Now().UnixNano()))
	if _, err := runCmd(ctx, "", "git", "worktree", "add", "--detach", tmp, commit); err != nil {
		return nil, nil, fmt.Errorf("git worktree add: %w", err)
	}
	defer func() {
		_, _ = runCmd(context.Background(), "", "git", "worktree", "remove", "--force", tmp)
	}()

	args := benchArgsFor(c)

	// run benchmark in that worktree
	out, err := runCmd(ctx, tmp, "go", args...)
	if err != nil {
		// include tail of output for debugging
		msg := string(out)
		if len(msg) > 600 {
			msg = msg[len(msg)-600:]
		}
		return out, args, fmt.Errorf("go %s failed: %v\n…%s", strings.Join(args, " "), err, msg)
	}
	return out, args, nil
}

func benchArgsFor(c *cli.Command) []string {
	pkgs := c.String("pkgs")
	if strings.TrimSpace(pkgs) == "" {
		pkgs = "./..."
	}
	args := []string{"test", pkgs, "-run=^$", "-bench", c.String("bench"), "-benchmem"}
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

func runCmd(ctx context.Context, dir string, bin string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, bin, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, fmt.Errorf("%s %s: %w", bin, strings.Join(args, " "), err)
	}
	return out, nil
}

func short(sha string) string {
	if len(sha) > 8 {
		return sha[:8]
	}
	return sha
}
