package sinks

import (
	"context"
	"os"

	"github.com/cloud66/trackman/utils"
)

// NewConsoleSink returns a console sink. This is a basic sink just to
// dump values onto the screen.
func NewConsoleSink(ctx context.Context, notifier utils.EventNotifier) (*utils.Sink, error) {
	return &utils.Sink{
		Notifier:   notifier,
		ErrChannel: os.Stderr,
		OutChannel: os.Stdout,
	}, nil
}
