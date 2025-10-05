package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
)

func init() {
	cmds = append(cmds, command{
		name: cmdShow,
		usages: []string{
			fmt.Sprintf("%s REF\tShow stored note for a commit/ref", cmdShow),
		},
		run: func(ctx context.Context, params *commandParams) error {
			var ref string
			if err := requireArgs(params.fs.Args(), &ref); err != nil {
				return err
			}
			return Show(ctx, &ShowArgs{Root: params.root, Ref: ref}, params.logger)
		},
	})
}

const cmdShow = "show"

type ShowArgs struct {
	Root *RootFlags
	Ref  string
}

// Show displays a stored note for a given commit/ref.
// It prints the exact JSON payload that Backfill stored (pretty-formatted).
func Show(ctx context.Context, a *ShowArgs, logger *slog.Logger) error {
	notesRef := a.Root.NotesRef

	sha, err := resolveCommit(ctx, a.Ref)
	if err != nil {
		return fmt.Errorf("resolve %q: %w", a.Ref, err)
	}

	logger.DebugContext(ctx, "show start", "notes_ref", notesRef, "commit", sha)

	raw, err := gitNotesShow(ctx, notesRef, sha)
	if err != nil {
		if errors.Is(err, errNoteMissing) {
			logger.ErrorContext(ctx, "note missing", "commit", sha, "notes_ref", notesRef)
			return nil
		}
		return err
	}

	// Pretty-print the JSON payload.
	var js map[string]any
	if err := json.Unmarshal(raw, &js); err != nil {
		logger.ErrorContext(ctx, "unmarshal notes", "error", err, "commit", sha, "notes_ref", notesRef)
		return nil
	}
	logger.InfoContext(ctx, "result", "commit", sha, "notes_ref", notesRef)
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
