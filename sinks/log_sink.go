package sinks

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/cloud66/trackman/utils"
)

// NewLogSink creates a new LogSink
func NewLogSink(ctx context.Context, notifier utils.EventNotifier) (*utils.SinkOptions, error) {
	outWriter := utils.NewLogWriter(ctx, logrus.InfoLevel)
	errWriter := utils.NewLogWriter(ctx, logrus.ErrorLevel)

	return &utils.SinkOptions{
		Notifier:   notifier,
		ErrChannel: errWriter,
		OutChannel: outWriter,
	}, nil
}
