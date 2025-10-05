package main

import (
	"context"
	"log/slog"
	"os"
)

type cmd struct {
	logger *slog.Logger
}

func main() {
	ctx := context.Background()
	logLevel := &slog.LevelVar{}

	cmd := &cmd{
		logger: newSplitLogger(os.Stdout, os.Stderr, logLevel),
	}

	rootCMD := cmd.Root(logLevel)

	if err := rootCMD.Run(ctx, os.Args); err != nil {
		cmd.logger.Error("command failed", "error", err)
		os.Exit(1)
	}
}

type contextValues struct{}

var (
	rawHandler = contextValues{}
)
