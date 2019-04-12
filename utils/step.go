package utils

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"strings"
	"time"
)

const (
	stepPending = 1
	stepRunning = 2
	stepDone    = 3
)

// StepOptions provides options for a Step
type StepOptions struct {
	Notifier func(ctx context.Context, event *Event) error
}

// Step is a single running Step
type Step struct {
	Metadata       map[string]string `yaml:"metadata"`
	Name           string            `yaml:"name"`
	Command        string            `yaml:"command"`
	ContinueOnFail bool              `yaml:"continue_on_fail"`
	Timeout        *time.Duration    `yaml:"timeout"`
	Workdir        string            `yaml:"workdir"`
	Probe          *Probe            `yaml:"probe"`
	DependsOn      []string          `yaml:"depends_on"`
	Preflights     []Preflight       `yaml:"preflights"`

	options   *StepOptions
	workflow  *Workflow
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

func (s *Step) parseCommand(ctx context.Context) error {
	buf := &bytes.Buffer{}
	tmpl, err := template.New("t1").Parse(s.Command)
	if err != nil {
		return err
	}

	err = tmpl.Execute(buf, s)
	if err != nil {
		return err
	}

	s.Command = buf.String()

	return nil
}

func (s *Step) expandEnvVars(ctx context.Context) {
	expandedCommand := os.ExpandEnv(s.Command)
	s.Command = expandedCommand

	if s.Workdir != "" {
		s.Workdir = os.ExpandEnv(s.Workdir)
	}
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

	logger, ctx := LoggerContext(ctx)

	err := s.parseCommand(ctx)
	if err != nil {
		// a failure here is down to workflow errors so
		// continue on failure doesn't apply
		return err
	}
	s.expandEnvVars(ctx)

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
