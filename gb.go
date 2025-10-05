package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"slices"
	"text/tabwriter"
)

func init() {
	cmds = append(cmds,
		command{
			name:    "version",
			aliases: []string{"--version"},
			usages: []string{
				"version\tShow version",
			},
			run: func(ctx context.Context, stdout, stderr io.Writer, prog string, args []string) error {
				fmt.Fprintln(stdout, version)
				return nil
			},
		},
		command{
			name:    "help",
			aliases: []string{"-h", "--help"},
			usages: []string{
				"help\tShow help",
			},
			run: func(ctx context.Context, stdout, stderr io.Writer, prog string, args []string) error {
				Usage(stdout, prog)
				return nil
			},
		},
	)
}

type command struct {
	name     string
	aliases  []string
	usages   []string
	examples []string
	run      func(ctx context.Context, stdout, stderr io.Writer, prog string, args []string) error
}

var (
	defaultCMD command
	cmds       = []command{}
)

type RootFlags struct {
	Verbose   bool
	Count     int
	Benchtime string
	Bench     string
	Pkgs      string
	NotesRef  string
	Force     bool
}

// ParseRootFlags wires common flags into fs and returns the struct pointer.
func ParseRootFlags(fs *flag.FlagSet) *RootFlags {
	cfg := &RootFlags{}
	fs.BoolVar(&cfg.Verbose, "v", false, "verbose output")
	fs.IntVar(&cfg.Count, "count", 10, "benchmark count")
	fs.StringVar(&cfg.Benchtime, "benchtime", "", "benchtime duration (e.g. 2s)")
	fs.StringVar(&cfg.Bench, "bench", ".", "benchmark regex")
	fs.StringVar(&cfg.Pkgs, "pkgs", "./...", "comma-separated package list")
	fs.StringVar(&cfg.NotesRef, "notes-ref", "", "override notes ref (default derived from env)")
	fs.BoolVar(&cfg.Force, "force", false, "allow cross-env comparisons")
	return cfg
}

func Usage(stderr io.Writer, invokedName string) {
	name := invokedName
	if version == "dev" {
		name = fmt.Sprintf("%s_dev", tool)
	}
	w := tabwriter.NewWriter(stderr, 0, 8, 2, ' ', 0)
	fmt.Fprintf(w, "Usage:\n")
	for _, cmd := range cmds {
		for _, u := range cmd.usages {
			fmt.Fprintf(w, "\t%s %s\n", name, u)
		}
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flags:")
	fmt.Fprintln(w, "\t-v\tVerbose output")
	fmt.Fprintln(w, "\t-count N\tBenchmark count (default 10)")
	fmt.Fprintln(w, "\t-benchtime D\tBenchtime duration (e.g. 2s)")
	fmt.Fprintln(w, "\t-bench PAT\tBenchmark regex (default '.')")
	fmt.Fprintln(w, "\t-pkgs LIST\tComma-separated package list (default './...')")
	fmt.Fprintln(w, "\t-notes-ref R\tOverride notes ref (default derived from env)")
	fmt.Fprintln(w, "\t-force\tAllow cross-environment comparisons")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "Examples:\n")
	for _, cmd := range cmds {
		for _, e := range cmd.examples {
			fmt.Fprintf(w, "\t%s %s\n", name, e)
		}
	}
	_ = w.Flush()
}

// Run dispatches based on argv[1]. It then parses the rest of the arguments
// and calls the appropriate command function.
func Run(ctx context.Context, argv []string, stdout, stderr io.Writer) error {
	prog := programName(argv)

	if len(argv) == 1 {
		return defaultCMD.run(ctx, stdout, stderr, prog, argv[1:])
	}

	targetCMD := argv[1]

	for _, cmd := range cmds {
		if targetCMD != cmd.name && !slices.Contains(cmd.aliases, targetCMD) {
			continue
		}
		return cmd.run(ctx, stdout, stderr, prog, argv[2:])
	}

	return defaultCMD.run(ctx, stdout, stderr, prog, argv[1:])
}

func programName(argv []string) string {
	if len(argv) > 0 && argv[0] != "" {
		return argv[0]
	}
	return tool
}
