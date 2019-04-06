package sinks

import (
	"context"
	"os"

	"github.com/cloud66/trackman/utils"
)

func NewConsoleSink(ctx context.Context, notifier utils.EventNotifier) (*utils.Sink, error) {
	return &utils.Sink{
		StdOut:   os.Stdout,
		StdErr:   os.Stderr,
		Notifier: notifier,
	}, nil
}
