package main

import (
	"context"
	"os"
)

// Version is set at build time with -ldflags "-X main.version=v0.1.0".
var (
	version = "dev"
	tool    = "gb"
)

func main() {
	// Pass everything into Run for testability.
	if err := Run(context.Background(), os.Args, os.Stdout, os.Stderr); err != nil {
		// Best-effort printing; Run already wrote user-friendly messages.
		os.Exit(1)
	}
}
