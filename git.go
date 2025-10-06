package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	errExit        = errors.New("exit")
	errNoteMissing = errors.New("note missing")
)

func gitWorktreeRunCommand(ctx context.Context, commit string, cmd []string) ([]byte, error) {
	if len(cmd) == 0 {
		return nil, fmt.Errorf("no benchmark command provided")
	}
	base := cmd[0]

	tmp, err := gitWorktreeAdd(ctx, commit)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = gitWorktreeRemove(context.WithoutCancel(ctx), tmp)
	}()

	var args []string
	if len(cmd) > 1 {
		args = cmd[1:]
	}

	out, err := runCmd(ctx, tmp, base, args...)
	if err != nil {
		// Include tail of output for debugging.
		msg := string(out)
		if len(msg) > 600 {
			msg = msg[len(msg)-600:]
		}
		return out, fmt.Errorf("go %s failed: %v\n…%s", strings.Join(cmd, " "), err, msg)
	}
	return out, nil
}

func gitWorktreeAdd(ctx context.Context, ref string) (string, error) {
	tmp := filepath.Join(os.TempDir(), "gb-wt-"+ref[:8]+"-"+fmt.Sprint(time.Now().UnixNano()))
	_, err := runCmd(ctx, "", "git", "worktree", "add", "--detach", tmp, ref)
	return tmp, err
}

type worktreeInfo struct {
	path   string
	commit string
}

func gitWorktreeList(ctx context.Context) ([]worktreeInfo, error) {
	out, err := runCmd(ctx, "", "git", "worktree", "list")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var infos []worktreeInfo
	for _, l := range lines {
		parts := strings.Fields(l)
		if len(parts) < 2 {
			continue
		}
		infos = append(infos, worktreeInfo{path: parts[0], commit: parts[1]})
	}

	return infos, nil
}

func gitWorktreeRemove(ctx context.Context, dir string) error {
	_, err := runCmd(ctx, "", "git", "worktree", "remove", "--force", dir)
	return err
}

func gitRevList(ctx context.Context, rangeSpec string) ([]string, error) {
	out, err := runCmd(ctx, "", "git", "rev-list", rangeSpec)
	if err != nil {
		return nil, err
	}
	return strings.Fields(string(out)), nil
}

func gitNotesAdd(ctx context.Context, notesRef, commit string, payload []byte) error {
	// We use -f to overwrite if a concurrent run added one; normally it won't exist.
	_, err := runCmd(ctx, "", "git", "notes", "--ref", notesRef, "add", "-f", "-m", string(payload), commit)
	return err
}

func gitNotesShow(ctx context.Context, notesRef, commit string) ([]byte, error) {
	out, err := runCmd(ctx, "", "git", "notes", "--ref", notesRef, "show", commit)
	if err != nil {
		if errors.Is(err, errExit) {
			return nil, errNoteMissing
		}
		return nil, err
	}
	return out, nil
}

func listAllNotesRefs(ctx context.Context, commit string) (map[string][]byte, error) {
	out, err := runCmd(ctx, "", "git", "for-each-ref", "--format=%(refname)", "refs/notes")
	if err != nil {
		return nil, fmt.Errorf("git for-each-ref refs/notes: %w", err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	notes := make(map[string][]byte)

	for _, ref := range lines {
		ref = strings.TrimSpace(ref)
		if ref == "" {
			continue
		}

		value, err := gitNotesShow(ctx, ref, commit)
		if err == nil {
			notes[ref] = value
		}
	}
	return notes, nil
}

func gitResolveCommit(ctx context.Context, ref string) (string, error) {
	out, err := runCmd(ctx, "", "git", "rev-parse", ref)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func gitEmail(ctx context.Context) (string, error) {
	out, err := runCmd(ctx, "", "git", "config", "user.email")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func runCmd(ctx context.Context, dir string, bin string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, bin, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return nil, errExit
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		return nil, fmt.Errorf("command \"%s %s\" returned: %w", bin, strings.Join(args, " "), err)
	}
	return out, nil
}
