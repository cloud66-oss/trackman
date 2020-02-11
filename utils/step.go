package utils

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	stepPending = 1
	stepRunning = 2
	stepDone    = 3
)

// StepOptions provides options for a Step
type StepOptions struct {
	Notifier func(ctx context.Context, logger *logrus.Logger, event *Event) error
}

// Step is a single running Step
type Step struct {
	Metadata       map[string]string `yaml:"metadata" json:"metadata"`
	Name           string            `yaml:"name" json:"name"`
	Command        string            `yaml:"command" json:"command"`
	ContinueOnFail bool              `yaml:"continue_on_fail" json:"continue_on_fail"`
	Timeout        *time.Duration    `yaml:"timeout" json:"timeout"`
	Workdir        string            `yaml:"workdir" json:"workdir"`
	Env            []string          `yaml:"env" json:"env"`
	Probe          *Probe            `yaml:"probe" json:"probe"`
	DependsOn      []string          `yaml:"depends_on" json:"depends_on"`
	Preflights     []Preflight       `yaml:"preflights" json:"preflights"`
	AskToProceed   bool              `yaml:"ask_to_proceed" json:"ask_to_proceed"`
	ShowCommand    bool              `yaml:"show_command" json:"show_command"`
	Disabled       bool              `yaml:"disabled" json:"disabled"`
	Logger         *LogDefinition    `yaml:"logger" json:"logger"`

	options   *StepOptions
	workflow  *Workflow
	logger    *logrus.Logger
	status    int
	dependsOn []*Step
}

// String overrides string
func (s *Step) String() string {
	str := fmt.Sprintf("%s: %d", s.Name, s.status)
	var deps []string
	for _, step := range s.dependsOn {
		deps = append(deps, step.String())
	}

	return str + "\n" + strings.Join(deps, ",")
}

// MergedMetadata merges step and workflow metadata
func (s *Step) MergedMetadata() map[string]string {
	if s.Metadata == nil {
		return s.workflow.Metadata
	}

	result := make(map[string]string, len(s.workflow.Metadata))
	for k, v := range s.workflow.Metadata {
		result[k] = v
	}

	for k, v := range s.Metadata {
		result[k] = v
	}

	return result
}

// shouldRun returns a step that can be run, hasn't started, isn't done and isn't marked to be done
func (s *Step) shouldRun() bool {
	// has this run or marked to run?
	status := s.status != stepRunning && s.status != stepDone && s.status != stepPending
	if !status {
		// if it has, then it shouldn't run
		return false
	}

	// this can run but how about the dependencies?
	for _, step := range s.dependsOn {
		if !step.isDone() {
			// there is a dependency that is not done
			return false
		}
	}

	// all good, we can run this
	return true
}

func (s *Step) isDone() bool {
	return s.status == stepDone
}

// MarkAsPending marks the step as pending meaning it's waiting to run
func (s *Step) MarkAsPending() {
	s.status = stepPending
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

// Run runs a Step and its probe
func (s *Step) Run(ctx context.Context) error {
	s.status = stepRunning
	defer func() { s.status = stepDone }()

	if s.Disabled {
		s.logger.WithField(FldStep, s.Name).Info("Disabled step. Skipping")
		return nil
	}

	err := s.EnrichStep(ctx)
	if err != nil {
		return err
	}

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

		s.logger.WithField(FldStep, spinner.Name).Error(err)
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

			s.logger.WithField(FldStep, probeSpinner.Name).Error(err)
		}
	}

	return nil
}

// EnrichStep resolves environment variables and parses the command for the step
// on all applicable attributes
func (s *Step) EnrichStep(ctx context.Context) error {
	var err error

	// parse for meta data
	if s.Metadata != nil {
		for idx, metadata := range s.Metadata {
			if s.Metadata[idx], err = s.parseAttribute(ctx, metadata); err != nil {
				return err
			}
		}
	}
	if s.Command, err = s.parseAttribute(ctx, s.Command); err != nil {
		return err
	}
	if s.Name, err = s.parseAttribute(ctx, s.Name); err != nil {
		return err
	}
	if s.Workdir, err = s.parseAttribute(ctx, s.Workdir); err != nil {
		return err
	}
	if s.Probe != nil {
		if s.Probe.Command, err = s.parseAttribute(ctx, s.Probe.Command); err != nil {
			return err
		}
		if s.Probe.Workdir, err = s.parseAttribute(ctx, s.Probe.Workdir); err != nil {
			return err
		}
	}
	if s.Logger != nil {
		if s.Logger.Destination, err = s.parseAttribute(ctx, s.Logger.Destination); err != nil {
			return err
		}
		if s.Logger.Format, err = s.parseAttribute(ctx, s.Logger.Format); err != nil {
			return err
		}
		if s.Logger.Level, err = s.parseAttribute(ctx, s.Logger.Level); err != nil {
			return err
		}
		if s.Logger.Type, err = s.parseAttribute(ctx, s.Logger.Type); err != nil {
			return err
		}
	}
	if s.Preflights != nil {
		for idx, preFlight := range s.Preflights {
			if s.Preflights[idx].Command, err = s.parseAttribute(ctx, preFlight.Command); err != nil {
				return err
			}
			if s.Preflights[idx].Workdir, err = s.parseAttribute(ctx, preFlight.Workdir); err != nil {
				return err
			}
			if s.Preflights[idx].Message, err = s.parseAttribute(ctx, preFlight.Message); err != nil {
				return err
			}
		}
	}

	// expand env var
	if s.Metadata != nil {
		for idx, metadata := range s.Metadata {
			if s.Metadata[idx], err = ExpandEnvVars(ctx, metadata); err != nil {
				return err
			}
		}
	}
	if s.Command, err = ExpandEnvVars(ctx, s.Command); err != nil {
		return err
	}
	if s.Workdir, err = ExpandEnvVars(ctx, s.Workdir); err != nil {
		return err
	}
	if s.Command, err = ExpandEnvVars(ctx, s.Command); err != nil {
		return err
	}
	if s.Name, err = ExpandEnvVars(ctx, s.Name); err != nil {
		return err
	}
	if s.Workdir, err = ExpandEnvVars(ctx, s.Workdir); err != nil {
		return err
	}
	if s.Probe != nil {
		if s.Probe.Command, err = ExpandEnvVars(ctx, s.Probe.Command); err != nil {
			return err
		}
		if s.Probe.Workdir, err = ExpandEnvVars(ctx, s.Probe.Workdir); err != nil {
			return err
		}
	}
	if s.Logger != nil {
		if s.Logger.Destination, err = ExpandEnvVars(ctx, s.Logger.Destination); err != nil {
			return err
		}
		if s.Logger.Format, err = ExpandEnvVars(ctx, s.Logger.Format); err != nil {
			return err
		}
		if s.Logger.Level, err = ExpandEnvVars(ctx, s.Logger.Level); err != nil {
			return err
		}
		if s.Logger.Type, err = ExpandEnvVars(ctx, s.Logger.Type); err != nil {
			return err
		}
	}
	if s.Preflights != nil {
		for idx, preFlight := range s.Preflights {
			if s.Preflights[idx].Command, err = ExpandEnvVars(ctx, preFlight.Command); err != nil {
				return err
			}
			if s.Preflights[idx].Workdir, err = ExpandEnvVars(ctx, preFlight.Workdir); err != nil {
				return err
			}
			if s.Preflights[idx].Message, err = ExpandEnvVars(ctx, preFlight.Message); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Step) parseAttribute(ctx context.Context, value string) (string, error) {
	if value == "" {
		return "", nil
	}

	buf := &bytes.Buffer{}
	tmpl, err := template.New("step").Parse(value)
	if err != nil {
		return "", err
	}

	err = tmpl.Execute(buf, s)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
