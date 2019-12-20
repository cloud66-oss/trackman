package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/kballard/go-shellquote"
	"github.com/sirupsen/logrus"
)

// Spinner is the main component that runs a process
type Spinner struct {
	UUID string
	Name string

	cmd     string
	args    []string
	env     []string
	timeout time.Duration
	workdir string
	step    Step
}

// NewSpinnerForStep creates a new instance of Spinner based on the Options
func NewSpinnerForStep(ctx context.Context, step Step) (*Spinner, error) {
	spinner, err := newSpinnerForStep(ctx, step)
	if err != nil {
		return nil, err
	}

	spinner.validate(ctx)

	return spinner, nil
}

// NewSpinnerForPreflight creates a new instance of Spinner based on the Options
func NewSpinnerForPreflight(ctx context.Context, preflight *Preflight) (*Spinner, error) {
	spinner, err := newSpinnerForPreflight(ctx, preflight)
	if err != nil {
		return nil, err
	}

	spinner.validate(ctx)

	return spinner, nil
}

// NewSpinnerForProbe creates a new instance of Spinner based on the Options
func NewSpinnerForProbe(ctx context.Context, step Step) (*Spinner, error) {
	spinner, err := newSpinnerForProbe(ctx, step)
	if err != nil {
		return nil, err
	}

	spinner.validate(ctx)

	return spinner, nil
}

func newSpinnerForStep(ctx context.Context, step Step) (*Spinner, error) {
	if step.options == nil {
		step.options = &StepOptions{
			Notifier: step.workflow.options.Notifier,
		}
	}

	parts, err := shellquote.Split(step.Command)
	if err != nil {
		return nil, err
	}

	return &Spinner{
		UUID:    uuid.New().String(),
		Name:    step.Name,
		cmd:     parts[0],
		args:    parts[1:],
		step:    step,
		env:     step.Env,
		workdir: step.Workdir,
	}, nil
}

func newSpinnerForPreflight(ctx context.Context, preflight *Preflight) (*Spinner, error) {
	if preflight.step.options == nil {
		preflight.step.options = &StepOptions{
			Notifier: preflight.step.workflow.options.Notifier,
		}
	}

	parts, err := shellquote.Split(preflight.Command)
	if err != nil {
		return nil, err
	}

	var timeout time.Duration
	if preflight.Timeout != nil {
		timeout = *preflight.Timeout
	}

	return &Spinner{
		UUID:    uuid.New().String(),
		Name:    fmt.Sprintf("%s.preflight", preflight.step.Name),
		cmd:     parts[0],
		args:    parts[1:],
		step:    *preflight.step,
		workdir: preflight.Workdir,
		env:     preflight.step.Env,
		timeout: timeout,
	}, nil
}

func newSpinnerForProbe(ctx context.Context, step Step) (*Spinner, error) {
	if step.options == nil {
		step.options = &StepOptions{
			Notifier: step.workflow.options.Notifier,
		}
	}

	parts, err := shellquote.Split(step.Probe.Command)
	if err != nil {
		return nil, err
	}

	return &Spinner{
		UUID:    uuid.New().String(),
		Name:    fmt.Sprintf("%s.probe", step.Name),
		cmd:     parts[0],
		args:    parts[1:],
		step:    step,
		env:     step.Env,
		workdir: step.Workdir,
	}, nil
}

func (s *Spinner) validate(ctx context.Context) {
	if s.step.workflow == nil {
		panic("no workflow")
	}

	if s.step.workflow.options == nil {
		panic("no workflow option")
	}

	if s.step.Timeout != nil {
		s.timeout = *s.step.Timeout
	} else {
		s.timeout = s.step.workflow.options.Timeout
	}
}

// Run runs the process required
func (s *Spinner) Run(ctx context.Context) error {
	s.push(ctx, NewEvent(s, EventRunRequested, nil))

	cmdCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	logger := s.step.logger

	// add this spinner to the context for the log writers
	ctx = context.WithValue(ctx, CtxSpinner, s)

	outChannel := NewLogWriter(ctx, logger, logrus.DebugLevel)
	errChannel := NewLogWriter(ctx, logger, logrus.ErrorLevel)

	logger.WithField(FldStep, s.Name).Tracef("Running %s with %s", s.cmd, s.args)

	cmd := exec.CommandContext(cmdCtx, s.cmd, s.args...)
	cmd.Stderr = errChannel
	cmd.Stdout = outChannel
	envs := os.Environ()
	for _, env := range s.env {
		envs = append(envs, env)
	}

	cmd.Env = envs
	cmd.Dir = s.workdir

	err := cmd.Start()
	if err != nil {
		s.push(ctx, NewEvent(s, EventRunError, nil))

		return err
	}

	s.push(ctx, NewEvent(s, EventRunStarted, nil))

	if err := cmd.Wait(); err != nil {
		if cmdCtx.Err() == context.DeadlineExceeded {
			s.push(ctx, NewEvent(s, EventRunTimeout, nil))

			return fmt.Errorf("Timed out after %s", s.timeout)
		}

		if exitErr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				s.push(ctx, NewEvent(s, EventRunFail, status))
				return exitErr
			}
		} else {
			// wait error
			s.push(ctx, NewEvent(s, EventRunWaitError, s))

			return exitErr
		}
	}

	s.push(ctx, NewEvent(s, EventRunSuccess, nil))

	return nil
}

func (s *Spinner) push(ctx context.Context, event *Event) {
	err := s.step.options.Notifier(ctx, s.step.logger, event)
	if err != nil {
		fmt.Println(err)
	}
}
