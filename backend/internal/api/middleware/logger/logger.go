package logger

import (
	"context"
	"log/slog"
)

type contextKey struct {
	name string
}

var loggerCtxKey = &contextKey{"Logger"}

func WithContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerCtxKey, l)
}

func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerCtxKey).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}
