package utils

import (
	"context"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	// FldStep is a logger field
	FldStep = "Step"
)

// LogWriter implements io.Writer so it can be used to dump a process output
// but links it to logrus
type LogWriter struct {
	entry   *logrus.Entry
	level   logrus.Level
	spinner *Spinner
}

// Write implements io.Writer
func (l *LogWriter) Write(b []byte) (int, error) {
	n := len(b)
	if n > 0 && b[n-1] == '\n' {
		b = b[:n-1]
	}

	// we want each line to show on its own
	for _, line := range strings.Split(string(b), "\n") {
		if l.spinner != nil {
			l.entry.WithField(FldStep, l.spinner.Name).Log(l.level, line)
		} else {
			l.entry.Log(l.level, line)
		}
	}

	return n, nil
}

// NewLogWriter creates a new LogWriter
func NewLogWriter(ctx context.Context, logger *logrus.Logger, level logrus.Level) *LogWriter {
	lw := &LogWriter{
		entry: logger.WithContext(ctx),
		level: level,
	}

	if ctx.Value(CtxSpinner) != nil {
		lw.spinner = ctx.Value(CtxSpinner).(*Spinner)
	}

	return lw
}
