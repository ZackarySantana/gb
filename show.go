package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/urfave/cli/v3"
)

func (cmd *cmd) Show() *cli.Command {
	return &cli.Command{
		Name:      "show",
		Usage:     "Show stored note for a commit/ref",
		Arguments: []cli.Argument{&cli.StringArg{Name: "ref", UsageText: "<git ref>"}},
		Action: func(ctx context.Context, c *cli.Command) error {
			notesRef := c.String("notes-ref")
			ref := c.StringArg("ref")

			if ref == "" {
				return fmt.Errorf("missing required argument: ref")
			}

			sha, err := resolveCommit(ctx, ref)
			if err != nil {
				return fmt.Errorf("resolve %q: %w", ref, err)
			}

			cmd.logger.DebugContext(ctx, "show start", "notes_ref", notesRef, "commit", sha)

			raw, err := gitNotesShow(ctx, notesRef, sha)
			if err != nil {
				return fmt.Errorf("reading note for commit %s: %w", sha, err)
			}

			var js map[string]any
			if err := json.Unmarshal(raw, &js); err != nil {
				cmd.logger.ErrorContext(ctx, "unmarshal notes", "error", err, "commit", sha, "notes_ref", notesRef)
				return fmt.Errorf("reading note for commit %s: %w", sha, err)
			}
			// TODO: use a table format for non-JSON output?
			cmd.logger.InfoContext(ctx, "result", "commit", sha, "notes_ref", notesRef)
			return nil
		},
	}
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
