package logx

import (
	"context"
)

type LoggerContextKey string

const (
	LoggerContext LoggerContextKey = "LOGGER_CONTEXT"
)

func WithCtx(ctx context.Context, log Logger) context.Context {
	return context.WithValue(ctx, LoggerContext, log)
}

func FromCtx(ctx context.Context) Logger {
	log, ok := ctx.Value(LoggerContext).(Logger)
	if !ok {
		panic("logger not exist in ctx")
	}
	return log
}

func NewCtx(log Logger) context.Context {
	return WithCtx(context.Background(), log)
}
