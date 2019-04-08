package utils

import (
	"context"
	"time"
)

// StepOptions provides options for a step
type StepOptions struct {
	Notifier func(ctx context.Context, event *Event) error
}

// Step is a single running step
type Step struct {
	Metadata       map[string]string `yaml:"metadata"`
	Name           string            `yaml:"name"`
	Command        string            `yaml:"command"`
	Args           []string          `yaml:"args"`
	ContinueOnFail bool              `yaml:"continue_on_fail"`
	Timeout        *time.Duration    `yaml:"timeout"`
	Workdir        string            `yaml:"workdir"`
	Probe          *Probe            `yaml:"probe"`

	options *StepOptions
}

// GetMetaData returns metadata value of the key from this Step.
// this is useful in event notifiers. It will return "" if there is
// no metadata with the given key
func (s *Step) GetMetaData(key string) string {
	if s.Metadata == nil {
		s.Metadata = make(map[string]string)
	}

	return s.Metadata[key]
}
