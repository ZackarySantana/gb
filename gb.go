package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"slices"
	"text/tabwriter"
)

func main() {
	logLevel := &slog.LevelVar{}
	logger := newSplitLogger(os.Stdout, os.Stderr, logLevel)

	if err := Run(context.Background(), os.Args, logger, logLevel); err != nil {
		os.Exit(1)
	}
}

var (
	// version is set at build time with -ldflags "-X main.version=v0.1.0".
	version = "dev"
	tool    = "gb"

	defaultCMD command
	cmds       = []command{}
)

type commandParams struct {
	logger *slog.Logger
	fs     *flag.FlagSet
	root   *RootFlags
	args   []string

	flags map[string]*string
}

type command struct {
	name     string
	aliases  []string
	usages   []string
	examples []string
	flags    func(ctx context.Context, params *commandParams)
	run      func(ctx context.Context, params *commandParams) error
}

func (c *command) execute(ctx context.Context, logger *slog.Logger, logLeveler *slog.LevelVar, args []string) error {
	fs, root := setupFlags(ctx, logger)
	params := &commandParams{logger: logger, fs: fs, root: root, args: args}
	if c.flags != nil {
		c.flags(ctx, params)
	}
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("%s: parse flags: %w", c.name, err)
	}
	if root.Verbose {
		logLeveler.Set(slog.LevelDebug)
	}

	return c.run(ctx, &commandParams{logger: logger, fs: fs, root: root, args: fs.Args()})
}

func init() {
	cmds = append(cmds,
		command{
			name:    "version",
			aliases: []string{"--version"},
			usages: []string{
				"version\tShow version",
			},
			run: func(ctx context.Context, params *commandParams) error {
				params.logger.Info("version", "version", version, "go", runtime.Version(), "os", runtime.GOOS, "arch", runtime.GOARCH)
				return nil
			},
		},
		command{
			name:    "help",
			aliases: []string{"-h", "--help"},
			usages: []string{
				"help\tShow help",
			},
			run: func(ctx context.Context, params *commandParams) error {
				Usage(ctx, params.logger)
				return nil
			},
		},
	)
}

type RootFlags struct {
	Verbose   bool
	Count     int
	Benchtime string
	Bench     string
	Pkgs      string
	NotesRef  string
	Force     bool
}

func setupFlags(ctx context.Context, logger *slog.Logger) (*flag.FlagSet, *RootFlags) {
	ctx = context.WithValue(ctx, rawHandler, true)
	fs := flag.NewFlagSet(cmdSync, flag.ContinueOnError)
	fs.SetOutput(NewLoggerWriter(ctx, logger.ErrorContext))
	fs.Usage = func() { Usage(ctx, logger) }
	gv := runtime.Version()
	notesRef := fmt.Sprintf("refs/notes/benches/%s-%s-%s", runtime.GOOS, runtime.GOARCH, gv)
	cfg := &RootFlags{}
	fs.BoolVar(&cfg.Verbose, "v", false, "verbose output")
	fs.IntVar(&cfg.Count, "count", 10, "benchmark count")
	fs.StringVar(&cfg.Benchtime, "benchtime", "", "benchtime duration (e.g. 2s)")
	fs.StringVar(&cfg.Bench, "bench", ".", "benchmark regex")
	fs.StringVar(&cfg.Pkgs, "pkgs", "./...", "comma-separated package list")
	fs.StringVar(&cfg.NotesRef, "notes-ref", notesRef, "override notes ref (default derived from env)")
	fs.BoolVar(&cfg.Force, "force", false, "allow cross-env comparisons")
	return fs, cfg
}

// requireArgs accepts pointers to n strings and fills them with the
// positional arguments. It returns an error if the number of positional
// arguments does not match n.
func requireArgs(pos []string, args ...*string) error {
	if len(pos) != len(args) {
		names := make([]string, len(args))
		for i := range args {
			names[i] = fmt.Sprintf("ARG%d", i+1)
		}
		return fmt.Errorf("expected %d args (%s), got %d", len(args), names, len(pos))
	}
	for i := range args {
		*args[i] = pos[i]
	}
	return nil
}

// optionalArgs accepts pointers to n strings and fills them with the
// positional arguments. It returns true if all n arguments were provided,
// false if fewer were provided (in which case the remaining args are not set).
func optionalArgs(pos []string, args ...*string) bool {
	for i := range args {
		if i >= len(pos) {
			return false
		}
		*args[i] = pos[i]
	}
	return true
}

func Usage(ctx context.Context, logger *slog.Logger) {
	ctx = context.WithValue(ctx, rawHandler, true)
	name, ok := ctx.Value(cmdName).(string)
	if !ok || version == "dev" {
		name = fmt.Sprintf("%s_dev", tool)
	}
	w := tabwriter.NewWriter(NewLoggerWriter(ctx, logger.ErrorContext), 0, 8, 2, ' ', 0)
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

type contextValues struct{}

var (
	cmdName    = contextValues{}
	rawHandler = contextValues{}
)

// Run dispatches based on argv[1]. It then parses the rest of the arguments
// and calls the appropriate command function.
func Run(ctx context.Context, argv []string, logger *slog.Logger, logLeveler *slog.LevelVar) error {
	err := run(ctx, argv, logger, logLeveler)
	if err != nil {
		logger.ErrorContext(ctx, err.Error())
	}
	return err
}

func run(ctx context.Context, argv []string, logger *slog.Logger, logLeveler *slog.LevelVar) error {
	prog := programName(argv)
	ctx = context.WithValue(ctx, cmdName, prog)

	if len(argv) == 1 {
		return defaultCMD.execute(ctx, logger, logLeveler, argv[1:])
	}

	targetCMD := argv[1]

	for _, cmd := range cmds {
		if targetCMD != cmd.name && !slices.Contains(cmd.aliases, targetCMD) {
			continue
		}
		if err := cmd.execute(ctx, logger, logLeveler, argv[2:]); err != nil {
			return fmt.Errorf("%s: %w", cmd.name, err)
		}
		return nil
	}

	return defaultCMD.execute(ctx, logger, logLeveler, argv[1:])
}

func programName(argv []string) string {
	if len(argv) > 0 && argv[0] != "" {
		return argv[0]
	}
	return tool
}
