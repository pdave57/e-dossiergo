// Package logger provides a structured slog-based logger for e-Dossier.
package logger

import (
	"context"
	"log/slog"
	"os"
)

type contextKey struct{}

// New creates a JSON slog logger writing to stdout.
func New(level slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	}))
}

// WithContext embeds a logger into a context.
func WithContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, l)
}

// FromContext retrieves the logger embedded in the context, or returns
// the default slog logger.
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(contextKey{}).(*slog.Logger); ok && l != nil {
		return l
	}
	return slog.Default()
}
