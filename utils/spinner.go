package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Spinner is the main component that runs a process
type Spinner struct {
	cmd     string
	args    []string
	timeout time.Duration
	workdir string
	step    Step

	UUID     string
	StepName string // StepName is used for logging
}

// NewSpinnerForStep creates a new instance of Spinner based on the Options
func NewSpinnerForStep(ctx context.Context, step Step) (*Spinner, error) {
	spinner := newSpinnerForStep(ctx, step)
	spinner.expandEnvVars(ctx)

	return spinner, nil
}

// NewSpinnerForProbe creates a new instance of Spinner based on the Options
func NewSpinnerForProbe(ctx context.Context, step Step) (*Spinner, error) {
	spinner := newSpinnerForProbe(ctx, step)
	spinner.expandEnvVars(ctx)

	return spinner, nil
}

func newSpinnerForStep(ctx context.Context, step Step) *Spinner {
	if step.options == nil {
		panic("no options")
	}

	return &Spinner{
		UUID:     uuid.New().String(),
		StepName: step.Name,
		cmd:      step.Command,
		args:     step.Args,
		step:     step,
	}
}

func newSpinnerForProbe(ctx context.Context, step Step) *Spinner {
	if step.options == nil {
		panic("no options")
	}

	return &Spinner{
		UUID:     uuid.New().String(),
		StepName: step.Name,
		cmd:      step.Probe.Command,
		args:     step.Probe.Args,
		step:     step,
	}
}

func (s *Spinner) expandEnvVars(ctx context.Context) {
	expandedCommand := os.ExpandEnv(s.step.Command)
	s.cmd = expandedCommand

	for idx, item := range s.args {
		s.args[idx] = os.ExpandEnv(item)
	}

	if s.step.Timeout != nil {
		s.timeout = *s.step.Timeout
	} else {
		s.timeout = viper.GetDuration("timeout")
	}

	if s.workdir != "" {
		s.workdir = os.ExpandEnv(s.workdir)
	}
}

// Run runs the process required
func (s *Spinner) Run(ctx context.Context) error {
	s.push(ctx, NewEvent(s, EventRunRequested, nil))

	cmdCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	logger := GetLogger(ctx)
	ctx = context.WithValue(ctx, CtxLogger, logger)

	// add this spinner to the context for the log writers
	ctx = context.WithValue(ctx, CtxSpinner, s)

	outChannel := NewLogWriter(ctx, logrus.DebugLevel)
	errChannel := NewLogWriter(ctx, logrus.ErrorLevel)

	cmd := exec.CommandContext(cmdCtx, s.cmd, s.args...)
	cmd.Stderr = errChannel
	cmd.Stdout = outChannel

	err := cmd.Start()
	if err != nil {
		s.push(ctx, NewEvent(s, EventRunError, nil))

		return err
	}

	s.push(ctx, NewEvent(s, EventRunStarted, nil))

	if err := cmd.Wait(); err != nil {
		if cmdCtx.Err() == context.DeadlineExceeded {
			s.push(ctx, NewEvent(s, EventRunTimeout, nil))

			return fmt.Errorf("step %s timed out after %s", s.step.Name, s.timeout)
		}

		if exitErr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				s.push(ctx, NewEvent(s, EventRunFail, status))
				return nil
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
	err := s.step.options.Notifier(ctx, event)
	if err != nil {
		fmt.Println(err)
	}
}
