package main

import (
	"context"
	"fmt"
	"io"
)

func init() {
	cmds = append(cmds, command{
		name: cmdCompare,
		usages: []string{
			fmt.Sprintf("%s [base [head]]\tCompare stored notes for two refs", cmdCompare),
		},
		examples: []string{
			fmt.Sprintf("%s origin/main HEAD\tCompare two commits", cmdCompare),
		},
		run: func(ctx context.Context, stdout, stderr io.Writer, args []string) error {
			a, err := parseCompare(ctx, stderr, args)
			if err != nil {
				return err
			}
			return Compare(ctx, a, stdout, stderr)
		},
	})
}

const cmdCompare = "compare"

type CompareArgs struct {
	Root *RootFlags
	Base string
	Head string
}

func parseCompare(ctx context.Context, stderr io.Writer, args []string) (*CompareArgs, error) {
	fs, root := setupCommandFlags(ctx, cmdCompare, stderr)
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	pos := fs.Args()
	base, head := "origin/main", "HEAD"
	if len(pos) >= 1 {
		base = pos[0]
	}
	if len(pos) >= 2 {
		head = pos[1]
	}
	return &CompareArgs{Root: root, Base: base, Head: head}, nil
}

func Compare(ctx context.Context, a *CompareArgs, stdout, stderr io.Writer) error {
	fmt.Fprintf(stdout, "[compare] base=%s head=%s (todo)\n", a.Base, a.Head)
	return nil
}
