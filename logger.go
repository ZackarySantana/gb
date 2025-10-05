package main

import (
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/lmittmann/tint"
)

type splitHandler struct {
	stdoutLogger slog.Handler
	stderrLogger slog.Handler
	stdout       io.Writer
	stderr       io.Writer
}

func newSplitLogger(stdout, stderr io.Writer, level *slog.LevelVar) *slog.Logger {
	// stderrLogger := slog.NewTextHandler(stdout, &slog.HandlerOptions{Level: level})
	stdoutLogger := tint.NewHandler(stdout, &tint.Options{
		Level:      level,
		TimeFormat: time.Kitchen,
	})
	stderrLogger := tint.NewHandler(stderr, &tint.Options{
		Level:      level,
		TimeFormat: time.Kitchen,
	})
	return slog.New(&splitHandler{
		stdoutLogger: stdoutLogger,
		stderrLogger: stderrLogger,
		stdout:       stdout,
		stderr:       stderr,
	})
}

func (h *splitHandler) Enabled(ctx context.Context, level slog.Level) bool {
	// Delegate to underlying handler’s level filtering
	return h.stdoutLogger.Enabled(ctx, level)
}

func (h *splitHandler) Handle(ctx context.Context, r slog.Record) error {
	raw, ok := ctx.Value(rawHandler).(bool)
	if ok && raw {
		if r.Level >= slog.LevelError {
			_, err := h.stderr.Write([]byte(r.Message))
			return err
		}
		_, err := h.stdout.Write([]byte(r.Message))
		return err
	}
	if r.Level >= slog.LevelError {
		return h.stderrLogger.Handle(ctx, r)
	}
	return h.stdoutLogger.Handle(ctx, r)
}

func (h *splitHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &splitHandler{
		stdoutLogger: h.stdoutLogger.WithAttrs(attrs),
		stderrLogger: h.stderrLogger.WithAttrs(attrs),
	}
}

func (h *splitHandler) WithGroup(name string) slog.Handler {
	return &splitHandler{
		stdoutLogger: h.stdoutLogger.WithGroup(name),
		stderrLogger: h.stderrLogger.WithGroup(name),
	}
}

type loggerWriter struct {
	ctx context.Context
	log func(context.Context, string, ...any)
}

func (w *loggerWriter) Write(p []byte) (int, error) {
	w.log(w.ctx, string(p))
	return len(p), nil
}

func NewLoggerWriter(ctx context.Context, log func(context.Context, string, ...any)) io.Writer {
	return &loggerWriter{ctx: ctx, log: log}
}
