package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

func init() {
	cmds = append(cmds, command{
		name: cmdShow,
		usages: []string{
			fmt.Sprintf("%s REF\tShow stored note for a commit/ref", cmdShow),
		},
		run: func(ctx context.Context, stdout, stderr io.Writer, args []string) error {
			a, err := parseShow(ctx, stderr, args)
			if err != nil {
				return err
			}
			return Show(ctx, a, stdout, stderr)
		},
	})
}

const cmdShow = "show"

type ShowArgs struct {
	Root *RootFlags
	Ref  string
}

func parseShow(ctx context.Context, stderr io.Writer, args []string) (*ShowArgs, error) {
	fs := flag.NewFlagSet(cmdShow, flag.ContinueOnError)
	fs.SetOutput(stderr)
	root := ParseRootFlags(fs)
	fs.Usage = func() { Usage(ctx, stderr) }
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	ref := fs.Args()
	if len(ref) < 1 {
		fmt.Fprintln(stderr, "show: missing REF")
		return nil, flag.ErrHelp
	}
	return &ShowArgs{Root: root, Ref: ref[0]}, nil
}

// Show displays a stored note for a given commit/ref.
// It prints the exact JSON payload that Backfill stored (pretty-formatted).
func Show(ctx context.Context, a *ShowArgs, stdout, stderr io.Writer) error {
	notesRef := a.Root.NotesRef

	sha, err := resolveCommit(ctx, a.Ref)
	if err != nil {
		return fmt.Errorf("resolve %q: %w", a.Ref, err)
	}

	if a.Root.Verbose {
		fmt.Fprintf(stderr, "notes ref : %s\n", notesRef)
		fmt.Fprintf(stderr, "commit    : %s\n", sha)
	}

	raw, err := gitNotesShow(ctx, notesRef, sha)
	if err != nil {
		if errors.Is(err, errNoteMissing) {
			fmt.Fprintf(stderr, "no note found for %s in %s\n", sha, notesRef)
			return nil
		}
		return err
	}

	// Pretty-print the JSON payload.
	var js any
	if err := json.Unmarshal(raw, &js); err != nil {
		// If the payload isn't valid JSON for some reason, just print as-is.
		fmt.Fprintln(stdout, string(raw))
		return nil
	}
	b, _ := json.MarshalIndent(js, "", "  ")
	fmt.Fprintln(stdout, string(b))
	return nil
}

/* ------------------------------- helpers ---------------------------------- */

var errNoteMissing = errors.New("note missing")

func resolveCommit(ctx context.Context, ref string) (string, error) {
	out, err := runCmd(ctx, "", "git", "rev-parse", ref)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func gitNotesShow(ctx context.Context, notesRef, commit string) ([]byte, error) {
	out, err := runCmd(ctx, "", "git", "notes", "--ref", notesRef, "show", commit)
	if err != nil {
		// Missing note returns a non-zero exit; map to a friendlier error.
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return nil, errNoteMissing
		}
		return nil, err
	}
	return out, nil
}
