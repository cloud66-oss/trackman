package utils

import (
	"bytes"
	"fmt"

	"github.com/sirupsen/logrus"
)

// SpinFormatter is a logrus custom formatter to display output of a specific Spinner
type SpinFormatter struct {
}

// Format implements logrus Formatter
func (f *SpinFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer

	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	if entry.Data["process"] != nil {
		_, _ = fmt.Fprintf(b, "[%s]", entry.Data["process"])
	}
	b.WriteString(" --> ")
	if entry.Level == logrus.ErrorLevel {
		_, _ = fmt.Fprintf(b, "\x1b[31m%s\x1b[0m", entry.Message)
	} else {
		b.WriteString(entry.Message)
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}
