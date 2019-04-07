package utils

import "io"

// Sink is where workflow / spinners send their output
type Sink struct {
	Notifier   EventNotifier
	OutChannel io.Writer
	ErrChannel io.Writer
}
