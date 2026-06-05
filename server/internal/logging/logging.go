// Package logging provides the structured JSON logger and request-scoped
// correlation-id helpers used across the server. Log records carry the fields
// required by ADR-0001 D8.3: severity, component, and (when available)
// session_id, tick, message_id, and correlation_id.
package logging

import (
	"context"
	"io"
	"log/slog"
	"os"
)

type ctxKey int

const correlationKey ctxKey = iota

// New returns a structured JSON logger writing to stdout. The "level" key is
// renamed to "severity" and "msg" to "message" to match the logging schema.
func New(env string) *slog.Logger {
	return NewTo(os.Stdout, env)
}

// NewTo is like New but writes to an arbitrary writer (used by tests).
func NewTo(w io.Writer, env string) *slog.Logger {
	level := slog.LevelInfo
	if env == "local" {
		level = slog.LevelDebug
	}
	h := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.LevelKey:
				a.Key = "severity"
			case slog.MessageKey:
				a.Key = "message"
			case slog.TimeKey:
				a.Key = "ts"
			}
			return a
		},
	})
	return slog.New(h)
}

// Component returns a child logger tagged with the given component name.
func Component(l *slog.Logger, name string) *slog.Logger {
	return l.With("component", name)
}

// ContextWithCorrelation stores a correlation id on the context.
func ContextWithCorrelation(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, correlationKey, id)
}

// CorrelationFromContext retrieves the correlation id, if any.
func CorrelationFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(correlationKey).(string)
	return id, ok && id != ""
}
