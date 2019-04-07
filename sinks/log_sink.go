package sinks

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/cloud66/trackman/utils"
)

// NewLogSink creates a new LogSink
func NewLogSink(ctx context.Context, notifier utils.EventNotifier, spinner *utils.Spinner) (*utils.Sink, error) {
	outWriter := utils.NewLogWriter(ctx, logrus.InfoLevel, spinner)
	errWriter := utils.NewLogWriter(ctx, logrus.ErrorLevel, spinner)

	return &utils.Sink{
		Notifier:   notifier,
		ErrChannel: errWriter,
		OutChannel: outWriter,
	}, nil
}
