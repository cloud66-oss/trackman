package utils

import "io"

// SpinnerSink is where Spinners send their output
type SpinnerSink struct {
	StdOut   io.Writer
	StdErr   io.Writer
	Notifier EventNotifier
}
