package utils

import "io"

// Sink is where workflow / spinners send their output
type Sink struct {
	StdOut   io.Writer
	StdErr   io.Writer
	Notifier EventNotifier
}
