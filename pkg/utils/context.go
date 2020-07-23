package utils

import (
	"context"
	"go.uber.org/zap"
)

type contextLoggerKey struct{}

func GetContextLogger(ctx context.Context) (logger *zap.SugaredLogger) {
	logger, ok := ctx.Value(contextLoggerKey{}).(*zap.SugaredLogger)
	if !ok {
		panic("expected logger")
	}

	return
}

func WithContextLogger(ctx context.Context, logger *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, contextLoggerKey{}, logger)
}
