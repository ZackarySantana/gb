package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
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
		run: func(ctx context.Context, stdout, stderr io.Writer, args []string) error {
			a, err := parseBackfill(ctx, stderr, args)
			if err != nil {
				return err
			}
			return Backfill(ctx, a, stdout, stderr)
		},
	})
}

const cmdBackfill = "backfill"

type BackfillArgs struct {
	Root  *RootFlags
	Since string
}

func parseBackfill(ctx context.Context, stderr io.Writer, args []string) (*BackfillArgs, error) {
	fs := flag.NewFlagSet(cmdBackfill, flag.ContinueOnError)
	fs.SetOutput(stderr)
	root := ParseRootFlags(fs)
	fs.Usage = func() { Usage(ctx, stderr) }
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	ref := fs.Args()
	if len(ref) < 1 {
		fmt.Fprintln(stderr, "backfill: missing REF")
		return nil, flag.ErrHelp
	}
	return &BackfillArgs{Root: root, Since: ref[0]}, nil
}

// Backfill walks commits since a ref and fills in missing notes.
func Backfill(ctx context.Context, a *BackfillArgs, stdout, stderr io.Writer) error {
	ref := a.Root.NotesRef

	if a.Root.Verbose {
		fmt.Fprintf(stderr, "notes ref: %s\n", ref)
		fmt.Fprintf(stderr, "range    : %s..HEAD\n", a.Since)
	}

	commits, err := gitRevList(ctx, a.Since+"..HEAD")
	if err != nil {
		return fmt.Errorf("git rev-list: %w", err)
	}
	if len(commits) == 0 {
		fmt.Fprintln(stdout, "no commits to backfill")
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
			fmt.Fprintf(stderr, "✖ %s: check note failed: %v\n", short(c), err)
			continue
		}
		if has {
			skipped++
			if a.Root.Verbose {
				fmt.Fprintf(stderr, "• %s: note exists, skipping\n", short(c))
			}
			continue
		}

		if a.Root.Verbose {
			fmt.Fprintf(stderr, "→ %s: benchmarking…\n", short(c))
		}

		raw, benchArgs, err := runBenchesInWorktree(ctx, c, a.Root)
		if err != nil {
			failed++
			fmt.Fprintf(stderr, "✖ %s: bench failed: %v\n", short(c), err)
			continue
		}

		payload, err := marshalNotePayload(c, benchArgs, raw)
		if err != nil {
			failed++
			fmt.Fprintf(stderr, "✖ %s: marshal note: %v\n", short(c), err)
			continue
		}

		if err := gitNotesAdd(ctx, ref, c, payload); err != nil {
			failed++
			fmt.Fprintf(stderr, "✖ %s: write note: %v\n", short(c), err)
			continue
		}

		done++
		if a.Root.Verbose {
			fmt.Fprintf(stderr, "✓ %s: noted\n", short(c))
		}
	}

	elapsed := time.Since(start).Truncate(time.Millisecond)
	fmt.Fprintf(stdout, "backfill complete: %d noted, %d skipped, %d failed in %s\n", done, skipped, failed, elapsed)
	if failed > 0 {
		return errors.New("some commits failed to backfill")
	}
	return nil
}

/* ------------------------------- helpers ---------------------------------- */

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


