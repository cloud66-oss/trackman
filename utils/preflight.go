package utils

import (
	"context"
	"fmt"
	"time"
)

// Preflight is a check that runs at the beginning of the workflow
type Preflight struct {
	Command string         `yaml:"command" json:"command"`
	Message string         `yaml:"message" json:"message"`
	Workdir string         `yaml:"workdir" json:"workdir"`
	Timeout *time.Duration `yaml:"timeout" json:"timeout"`

	step *Step
}

// Run runs the preflight
func (p *Preflight) Run(ctx context.Context) error {
	logger := p.step.logger
	logger.WithField(FldStep, fmt.Sprintf("%s.preflight", p.step.Name)).Tracef("Running preflight")
	spinner, err := NewSpinnerForPreflight(ctx, p)
	if err != nil {
		return err
	}

	return spinner.Run(ctx)
}
