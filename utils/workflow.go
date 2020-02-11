package utils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/thanhpk/randstr"
	"golang.org/x/sync/semaphore"
	"gopkg.in/yaml.v2"
)

// WorkflowOptions provides options for a workflow
type WorkflowOptions struct {
	Notifier    func(ctx context.Context, logger *logrus.Logger, event *Event) error
	Concurrency int
	Timeout     time.Duration
}

// Workflow is the internal object to hold a workflow file
type Workflow struct {
	Version  string            `yaml:"version" json:"version"`
	Metadata map[string]string `yaml:"metadata" json:"metadata"`
	Steps    []*Step           `yaml:"steps" json:"steps"`
	Logger   *LogDefinition    `yaml:"logger" json:"logger"`

	options    *WorkflowOptions
	logger     *logrus.Logger
	gatekeeper *semaphore.Weighted
	signal     *sync.Mutex
	stopFlag   bool
	sessionID  string
}

// LoadWorkflowFromBytes loads a workflow from bytes
func LoadWorkflowFromBytes(ctx context.Context, options *WorkflowOptions, buff []byte) (*Workflow, error) {
	var workflow *Workflow
	err := yaml.Unmarshal(buff, &workflow)
	if err != nil {
		return nil, err
	}
	if options == nil {
		panic("no options")
	}
	if options.Notifier == nil {
		panic("no notifier")
	}

	if workflow.Version != "1" {
		return nil, errors.New("invalid workflow version")
	}

	workflow.sessionID = randstr.String(8)
	workflow.gatekeeper = semaphore.NewWeighted(int64(options.Concurrency))
	workflow.options = options
	workflow.stopFlag = false
	workflow.signal = &sync.Mutex{}

	logger, err := NewLogger(workflow.Logger, NewLoggingContext(workflow, nil))
	if err != nil {
		return nil, err
	}
	workflow.logger = logger

	// validate depends on and link them to the step
	// TODO: check for circular dependencies
	for idx, step := range workflow.Steps {
		workflow.Steps[idx].workflow = workflow
		for _, priorStepName := range step.DependsOn {
			priorStep := workflow.findStepByName(priorStepName)
			if priorStep == nil {
				return nil, fmt.Errorf("invalid step name in depends_on for step %s (%s)", step.Name, priorStepName)
			}

			workflow.Steps[idx].dependsOn = append(workflow.Steps[idx].dependsOn, priorStep)
		}

		// setup logging for this step
		if step.Logger == nil {
			logger, err = NewLogger(workflow.Logger, NewLoggingContext(workflow, step))
			if err != nil {
				return nil, err
			}
		} else {
			logger, err = NewLogger(step.Logger, NewLoggingContext(workflow, step))
			if err != nil {
				return nil, err
			}
		}
		step.logger = logger
	}

	if err = workflow.EnrichWorkflow(ctx); err != nil {
		return workflow, err
	}

	return workflow, nil
}

// LoadWorkflowFromReader loads a workflow from an io reader
func LoadWorkflowFromReader(ctx context.Context, options *WorkflowOptions, reader io.Reader) (*Workflow, error) {
	buff, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return LoadWorkflowFromBytes(ctx, options, buff)
}

// SessionID returns the session id of this run for the workflow
func (w *Workflow) SessionID() string {
	return w.sessionID
}

func (w *Workflow) preflights(ctx context.Context) (preflights []*Preflight) {
	for kdx, step := range w.Steps {
		for idx := range step.Preflights {
			step.Preflights[idx].step = w.Steps[kdx]
			preflights = append(preflights, &step.Preflights[idx])
		}
	}

	return preflights
}

func (w *Workflow) preflightChecks(ctx context.Context) error {
	for _, preflight := range w.preflights(ctx) {
		err := preflight.Run(ctx)
		if err != nil {
			if preflight.Message != "" {
				// dump the message
				w.logger.WithField(FldStep, fmt.Sprintf("%s.preflight", preflight.step.Name)).Error(preflight.Message)
			}
			return err
		}
	}

	return nil
}

// Run runs the entire workflow
func (w *Workflow) Run(ctx context.Context) (runErrors error, stepErrors error) {
	// if w.Logger is null, it's going to use the defaults which should be the same as with the app
	// since the default values from from the same place
	w.logger.Infof("Running Workflow with Session ID %s", w.sessionID)
	w.logger.Info("Running Preflight checks")
	err := w.preflightChecks(ctx)
	if err != nil {
		return err, nil
	}
	w.logger.Info("Preflight checks complete")

	joiner := sync.WaitGroup{}

	// Run all that can run
	for {
		if w.shouldStop(ctx) {
			return nil, stepErrors
		}
		if w.allDone() {
			break
		}

		step := w.nextToRun(ctx)
		if step == nil {
			continue
		}

		w.logger.WithField(FldStep, step.Name).Trace("Next to run")

		err := w.gatekeeper.Acquire(ctx, 1)
		if err != nil {
			return err, nil
		}

		if w.shouldStop(ctx) {
			return
		}

		joiner.Add(1)
		go func(toRun *Step) {
			if w.shouldStop(ctx) {
				return
			}

			defer func() {
				w.logger.WithField(FldStep, toRun.Name).Trace("Done running")
				w.gatekeeper.Release(1)
				joiner.Done()
			}()

			w.logger.WithField(FldStep, toRun.Name).Trace("Preparing to run")

			if toRun.ShowCommand {
				w.logger.WithField(FldStep, toRun.Name).Info(toRun.Command)
			}

			if !toRun.Disabled && toRun.AskToProceed && !viper.GetBool("confirm.yes") {
				// we need an interactive permission for this
				if !confirm(fmt.Sprintf("Run %s?", toRun.Name), 1) {
					w.logger.WithField(FldStep, toRun.Name).Info("Stopping execution")
					w.stop(ctx)
				}
			}

			err := toRun.Run(ctx)
			if err != nil {
				stepErrors = multierror.Append(err, stepErrors)
				// run failed in some way that the whole workflow should stop
				w.logger.WithField(FldStep, toRun.Name).Error(err)
				w.logger.WithField(FldStep, toRun.Name).Error("Calling a stop to run")
				w.stop(ctx)
			}
		}(step)
	}

	joiner.Wait()

	return nil, stepErrors
}

// nextToRun returns the next step that can run
func (w *Workflow) nextToRun(ctx context.Context) *Step {
	// using a universal lock per workflow to pick the next step to run
	w.signal.Lock()
	defer w.signal.Unlock()

	for idx, step := range w.Steps {
		if step.shouldRun() {
			w.Steps[idx].MarkAsPending()
			return w.Steps[idx]
		}
	}

	return nil
}

func (w *Workflow) allDone() bool {
	w.signal.Lock()
	defer w.signal.Unlock()

	for _, step := range w.Steps {
		if !step.isDone() {
			return false
		}
	}

	return true
}

func (w *Workflow) findStepByName(name string) *Step {
	for idx, step := range w.Steps {
		if step.Name == name {
			return w.Steps[idx]
		}
	}

	return nil
}

func (w *Workflow) stop(ctx context.Context) {
	w.signal.Lock()
	defer w.signal.Unlock()

	w.stopFlag = true
}

func (w *Workflow) shouldStop(ctx context.Context) bool {
	w.signal.Lock()
	defer w.signal.Unlock()

	return w.stopFlag
}

// EnrichWorkflow parses and replaces any placeholders in the workflow
func (w *Workflow) EnrichWorkflow(ctx context.Context) error {
	var err error

	// meta data first
	if w.Metadata != nil {
		for idx, metadata := range w.Metadata {
			if w.Metadata[idx], err = w.parseAttribute(ctx, metadata); err != nil {
				return err
			}
		}
	}

	if w.Metadata != nil {
		for idx, metadata := range w.Metadata {
			if w.Metadata[idx], err = ExpandEnvVars(ctx, metadata); err != nil {
				return err
			}
		}
	}

	return nil
}

func (w *Workflow) parseAttribute(ctx context.Context, value string) (string, error) {
	if value == "" {
		return "", nil
	}

	buf := &bytes.Buffer{}
	tmpl, err := template.New("workflow").Parse(value)
	if err != nil {
		return "", err
	}

	err = tmpl.Execute(buf, w)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
