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

	options  *StepOptions
	workflow Workflow
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

// Run runs a step and its probe
func (s *Step) Run(ctx context.Context) error {
	logger, ctx := LoggerContext(ctx)

	spinner, err := NewSpinnerForStep(ctx, *s)
	if err != nil {
		return err
	}

	err = spinner.Run(ctx)
	if err != nil {
		if !s.ContinueOnFail {
			// main spinner failed and we need to get out
			return err
		}

		logger.WithField(FldStep, spinner.Name).Error(err)
	}

	// main spinner is done. we should use the probe to check if
	// it was successful

	if s.Probe != nil {
		probeSpinner, err := NewSpinnerForProbe(ctx, *s)
		if err != nil {
			return err
		}

		probeSpinner.push(ctx, NewEvent(probeSpinner, EventRunningProbe, nil))

		err = probeSpinner.Run(ctx)
		if err != nil {
			// probe failed
			if !s.ContinueOnFail {
				return err
			}

			logger.WithField(FldStep, probeSpinner.Name).Error(err)
		}
	}

	return nil
}
