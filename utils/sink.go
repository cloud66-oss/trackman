package utils

import "io"

// SinkOptions is where workflow / spinners send their output
type SinkOptions struct {
	Notifier   EventNotifier
	OutChannel io.Writer
	ErrChannel io.Writer
}
