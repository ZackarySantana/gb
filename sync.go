package main

import (
	"context"
	"flag"
	"fmt"
	"io"
)

func init() {
	cmds = append(cmds, command{
		name: cmdSync,
		usages: []string{
			fmt.Sprintf("%s --remote NAME\tSync benchmark notes with remote (push/fetch)\n", cmdSync),
		},
		run: func(ctx context.Context, stdout, stderr io.Writer, args []string) error {
			a, err := parseSync(ctx, stderr, args)
			if err != nil {
				return err
			}
			return Sync(ctx, a, stdout, stderr)
		},
	})
}

const cmdSync = "sync"

type SyncArgs struct {
	Root   *RootFlags
	Remote string
}

func parseSync(ctx context.Context, stderr io.Writer, args []string) (*SyncArgs, error) {
	fs := flag.NewFlagSet(cmdSync, flag.ContinueOnError)
	fs.SetOutput(stderr)
	root := ParseRootFlags(fs)
	remote := fs.String("remote", "origin", "git remote to sync with")
	fs.Usage = func() { Usage(ctx, stderr) }
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return &SyncArgs{Root: root, Remote: *remote}, nil
}

func Sync(ctx context.Context, a *SyncArgs, stdout, stderr io.Writer) error {
	fmt.Fprintf(stdout, "[sync] remote=%s (todo)\n", a.Remote)
	return nil
}
