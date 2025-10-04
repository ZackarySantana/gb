package main

import (
	"context"
	"flag"
	"fmt"
	"io"
)

func init() {
	cmds = append(cmds, command{
		name: cmdShow,
		usages: []string{
			fmt.Sprintf("%s REF\tShow stored note for a commit/ref", cmdShow),
		},
		run: func(ctx context.Context, stdout, stderr io.Writer, prog string, args []string) error {
			a, err := parseShow(stderr, prog, args)
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

func parseShow(stderr io.Writer, prog string, args []string) (*ShowArgs, error) {
	fs := flag.NewFlagSet(cmdShow, flag.ContinueOnError)
	fs.SetOutput(stderr)
	root := ParseRootFlags(fs)
	fs.Usage = func() { Usage(stderr, prog) }
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	pos := fs.Args()
	if len(pos) < 1 {
		fmt.Fprintln(stderr, "show: missing REF")
		return nil, flag.ErrHelp
	}
	return &ShowArgs{Root: root, Ref: pos[0]}, nil
}

func Show(ctx context.Context, a *ShowArgs, stdout, stderr io.Writer) error {
	fmt.Fprintf(stdout, "[show] ref=%s (todo)\n", a.Ref)
	return nil
}
