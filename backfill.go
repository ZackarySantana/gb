package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func init() {
	cmds = append(cmds, command{
		name: cmdBackfill,
		usages: []string{
			fmt.Sprintf("%s REF\tBackfill missing notes in REF..HEAD", cmdBackfill),
		},
		examples: []string{
			fmt.Sprintf("%s origin/main\tBackfill history", cmdBackfill),
		},
		run: func(ctx context.Context, params *commandParams) error {
			var ref string
			if err := requireArgs(params.fs.Args(), &ref); err != nil {
				return err
			}
			return Backfill(ctx, &BackfillArgs{Root: params.root, Since: ref}, params.logger)
		},
	})
}

const cmdBackfill = "backfill"

type BackfillArgs struct {
	Root  *RootFlags
	Since string
}

// Backfill walks commits since a ref and fills in missing notes.
func Backfill(ctx context.Context, a *BackfillArgs, logger *slog.Logger) error {
	ref := a.Root.NotesRef

	logger.DebugContext(ctx, "backfill start", "notes_ref", ref, "range", a.Since+"..HEAD")

	commits, err := gitRevList(ctx, a.Since+"..HEAD")
	if err != nil {
		return fmt.Errorf("git rev-list: %w", err)
	}
	if len(commits) == 0 {
		logger.InfoContext(ctx, "no commits to backfill")
		return nil
	}

	var done, skipped, failed int
	start := time.Now()

	for _, c := range commits {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		has, err := gitNoteExists(ctx, ref, c)
		if err != nil {
			failed++
			logger.ErrorContext(ctx, "git note exists failed", "commit", c, "error", err)
			continue
		}
		if has {
			skipped++
			logger.DebugContext(ctx, "note exists, skipping", "commit", c)
			continue
		}

		logger.DebugContext(ctx, "benchmarking", "commit", c)

		raw, benchArgs, err := runBenchesInWorktree(ctx, c, a.Root)
		if err != nil {
			failed++
			logger.ErrorContext(ctx, "benchmark failed", "commit", c, "error", err)
			continue
		}

		payload, err := marshalNotePayload(c, benchArgs, raw)
		if err != nil {
			failed++
			logger.ErrorContext(ctx, "marshal note failed", "commit", c, "error", err)
			continue
		}

		if err := gitNotesAdd(ctx, ref, c, payload); err != nil {
			failed++
			logger.ErrorContext(ctx, "git notes add failed", "commit", c, "error", err)
			continue
		}

		done++
		logger.DebugContext(ctx, "noted", "commit", c)
	}

	elapsed := time.Since(start).Truncate(time.Millisecond)
	logger.InfoContext(ctx, "backfill complete", "noted", done, "skipped", skipped, "failed", failed, "elapsed", elapsed)
	if failed > 0 {
		return errors.New("some commits failed to backfill")
	}
	return nil
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

func runBenchesInWorktree(ctx context.Context, commit string, root *RootFlags) ([]byte, []string, error) {
	// create temp worktree
	tmp := filepath.Join(os.TempDir(), "gb-wt-"+commit[:8]+"-"+fmt.Sprint(time.Now().UnixNano()))
	if _, err := runCmd(ctx, "", "git", "worktree", "add", "--detach", tmp, commit); err != nil {
		return nil, nil, fmt.Errorf("git worktree add: %w", err)
	}
	defer func() {
		_, _ = runCmd(context.Background(), "", "git", "worktree", "remove", "--force", tmp)
	}()

	args := benchArgsFor(root)

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

func benchArgsFor(root *RootFlags) []string {
	pkgs := root.Pkgs
	if strings.TrimSpace(pkgs) == "" {
		pkgs = "./..."
	}
	args := []string{"test", pkgs, "-run=^$", "-bench", root.Bench, "-benchmem"}
	if root.Count > 0 {
		args = append(args, "-count", fmt.Sprint(root.Count))
	}
	if strings.TrimSpace(root.Benchtime) != "" {
		args = append(args, "-benchtime", root.Benchtime)
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
