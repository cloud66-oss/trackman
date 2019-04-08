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

// SpinnerOptions provides options for a workflow
type SpinnerOptions struct {
	Notifier func(ctx context.Context, event *Event) error
}

// Spinner is the main component that runs a process
type Spinner struct {
	options *SpinnerOptions
	cmd     string
	args    []string
	timeout time.Duration

	UUID string
	Step Step
}

// NewSpinner creates a new instance of Spinner based on the Options
func NewSpinner(ctx context.Context, step Step, options *SpinnerOptions) (*Spinner, error) {
	if options.Notifier == nil {
		options.Notifier = func(ctx context.Context, event *Event) error {
			fmt.Println(event.String())

			return nil
		}
	}

	expandedCommand := os.ExpandEnv(step.Command)

	for idx, item := range step.Args {
		step.Args[idx] = os.ExpandEnv(item)
	}

	spinner := &Spinner{
		UUID:    uuid.New().String(),
		options: options,
		Step:    step,
		cmd:     expandedCommand,
		args:    step.Args,
		timeout: viper.GetDuration("timeout"),
	}

	return spinner, nil
}

// Run runs the process required
func (s *Spinner) Run(ctx context.Context) error {
	s.push(ctx, NewEvent(s, EventRunRequested, nil))

	cmdCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if ctx.Value(CtxLogger) == nil {
		ctx = context.WithValue(ctx, CtxLogger, logrus.New())
	}

	// add this spinner to the context for the log writers
	ctx = context.WithValue(ctx, CtxSpinner, s)

	outChannel := NewLogWriter(ctx, logrus.InfoLevel)
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

			return fmt.Errorf("step %s timed out after %s", s.Step.Name, s.timeout)
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
	err := s.options.Notifier(ctx, event)
	if err != nil {
		fmt.Println(err)
	}
}
