package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"
)

func init() {
	cmds = append(cmds, command{
		name: cmdBackfill,
		usages: []string{
			fmt.Sprintf("%s --since REF\tBackfill missing notes in REF..HEAD", cmdBackfill),
		},
		examples: []string{
			fmt.Sprintf("%s --since origin/main\tBackfill history", cmdBackfill),
		},
		run: func(ctx context.Context, stdout, stderr io.Writer, prog string, args []string) error {
			a, err := parseBackfill(stderr, prog, args)
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

func parseBackfill(stderr io.Writer, prog string, args []string) (*BackfillArgs, error) {
	fs := flag.NewFlagSet(cmdBackfill, flag.ContinueOnError)
	fs.SetOutput(stderr)
	root := ParseRootFlags(fs)
	since := fs.String("since", "", "start ref (required)")
	fs.Usage = func() { Usage(stderr, prog) }
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if strings.TrimSpace(*since) == "" {
		fmt.Fprintln(stderr, "--since REF is required")
		return nil, flag.ErrHelp
	}
	return &BackfillArgs{Root: root, Since: *since}, nil
}

func Backfill(ctx context.Context, a *BackfillArgs, stdout, stderr io.Writer) error {
	fmt.Fprintf(stdout, "[backfill] since=%s (todo)\n", a.Since)
	return nil
}
