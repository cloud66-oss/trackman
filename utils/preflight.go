package utils

import (
	"context"
	"fmt"
	"time"
)

// Preflight is a check that runs at the beginning of the workflow
type Preflight struct {
	Command string         `yaml:"command"`
	Message string         `yaml:"message"`
	Workdir string         `yaml:"workdir"`
	Timeout *time.Duration `yaml:"timeout"`

	step *Step
}

// Run runs the preflight
func (p *Preflight) Run(ctx context.Context) error {
	logger, _ := LoggerContext(ctx)
	logger.WithField(FldStep, fmt.Sprintf("%s.preflight", p.step.Name)).Tracef("Running preflight")
	spinner, err := NewSpinnerForPreflight(ctx, p)
	if err != nil {
		return err
	}

	return spinner.Run(ctx)
}