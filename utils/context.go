package utils

import (
	"context"

	"github.com/sirupsen/logrus"
)

// CtxKey is a context key
type CtxKey struct{ int }

var (
	// CtxSpinner is the key to a spinner on the context
	CtxSpinner = CtxKey{1}
	// CtxLogger is the key to a logger on the context
	CtxLogger = CtxKey{2}
	// CtxLogLevel holds the desired log level on the context
	CtxLogLevel = CtxKey{3}
)

// GetLogger returns a new or existing logger from the context
func getLogger(ctx context.Context) *logrus.Logger {
	var logger *logrus.Logger
	var logLevel logrus.Level
	if ctx.Value(CtxLogLevel) == nil {
		logLevel = logrus.DebugLevel
	} else {
		logLevel = ctx.Value(CtxLogLevel).(logrus.Level)
	}
	if ctx.Value(CtxLogger) == nil {
		logger = logrus.New()
		logger.SetLevel(logLevel)
	} else {
		logger = ctx.Value(CtxLogger).(*logrus.Logger)
	}

	return logger
}

// LoggerContext checks the context for a logger and creates a new one and puts it
// on the context if not there
func LoggerContext(ctx context.Context) (*logrus.Logger, context.Context) {
	logger := getLogger(ctx)
	ctx = context.WithValue(ctx, CtxLogger, logger)

	return logger, ctx
}
