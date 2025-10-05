package main

import (
	"context"
	"errors"
	"os/exec"
	"strings"
)

var errNoteMissing = errors.New("note missing")

// gitRevList returns a list of commit SHAs in reverse chronological order for the given range.
func gitRevList(ctx context.Context, rangeSpec string) ([]string, error) {
	out, err := runCmd(ctx, "", "git", "rev-list", "--reverse", rangeSpec)
	if err != nil {
		return nil, err
	}
	lines := strings.Fields(string(out))
	return lines, nil
}

// gitNoteExists checks if a note exists for the given commit in the specified notes ref.
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

// gitNotesAdd adds a note to the given commit in the specified notes ref.
func gitNotesAdd(ctx context.Context, notesRef, commit string, payload []byte) error {
	// We use -f to overwrite if a concurrent run added one; normally it won't exist.
	_, err := runCmd(ctx, "", "git", "notes", "--ref", notesRef, "add", "-f", "-m", string(payload), commit)
	return err
}

// gitNotesShow retrieves the note content for the given commit from the specified notes ref.
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

// resolveCommit resolves a git reference to its full SHA.
func resolveCommit(ctx context.Context, ref string) (string, error) {
	out, err := runCmd(ctx, "", "git", "rev-parse", ref)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}