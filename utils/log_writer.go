package utils

import (
	"context"
	"strings"

	"github.com/sirupsen/logrus"
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
			l.entry.WithField("process", l.spinner.step.Name).Log(l.level, line)
		} else {
			l.entry.Log(l.level, line)
		}
	}

	return n, nil
}

// NewLogWriter creates a new LogWriter
func NewLogWriter(ctx context.Context, level logrus.Level) *LogWriter {
	logger := logrus.New()
	logger.SetFormatter(&SpinFormatter{})
	lw := &LogWriter{
		entry: logger.WithContext(ctx),
		level: level,
	}

	if ctx.Value(ctxSpinner) != nil {
		lw.spinner = ctx.Value(ctxSpinner).(*Spinner)
	}

	return lw
}
