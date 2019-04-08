package utils

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// CtxKey is a context key
type CtxKey struct{ int }

var (
	// CtxSpinner is the key to a spinner on the context
	CtxSpinner = CtxKey{1}
	// CtxLogger is the key to a logger on the context
	CtxLogger = CtxKey{2}
)

// GetLogger returns a new or existing logger from the context
func getLogger(ctx context.Context) *logrus.Logger {
	var logger *logrus.Logger
	if ctx.Value(CtxLogger) == nil {
		logger = logrus.New()
		level, err := logrus.ParseLevel(viper.GetString("log-level"))
		if err != nil {
			logger.Fatal("invalid log-level")
		}
		logger.SetLevel(level)
	} else {
		logger = ctx.Value(CtxLogger).(*logrus.Logger)
	}

	return logger
}

func LoggerContext(ctx context.Context) (*logrus.Logger, context.Context) {
	logger := getLogger(ctx)
	ctx = context.WithValue(ctx, CtxLogger, logger)

	return logger, ctx
}
