package utils

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

// WorkflowOptions provides options for a workflow
type WorkflowOptions struct {
	NotificationManager *NotificationManager
}

// Workflow is the internal object to hold a workflow file
type Workflow struct {
	Version  string
	Metadata map[string]string
	Steps    []Step

	options *WorkflowOptions
}

// LoadWorkflowFromBytes loads a workflow from bytes
func LoadWorkflowFromBytes(buff []byte, options *WorkflowOptions) (*Workflow, error) {
	var workflow *Workflow
	err := yaml.Unmarshal(buff, &workflow)
	if err != nil {
		return nil, err
	}
	workflow.options = options

	return workflow, nil
}

// LoadWorkflowFromReader loads a workflow from an io reader
func LoadWorkflowFromReader(reader io.Reader, options *WorkflowOptions) (*Workflow, error) {
	buff, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return LoadWorkflowFromBytes(buff, options)
}

// Run runs the entire workflow
func (w *Workflow) Run(ctx context.Context) error {
	options := &SpinnerOptions{
		Sink: &SpinnerSink{
			StdOut: os.Stdout,
			StdErr: os.Stderr,
		},
		NotificationManager: w.options.NotificationManager,
	}

	err := w.options.NotificationManager.Start(ctx)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, step := range w.Steps {
		spinner, err := NewSpinner(ctx, step, options)
		if err != nil {
			return err
		}

		err = spinner.Run(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// Stop stops the run
func (w *Workflow) Stop(ctx context.Context) {
	w.options.NotificationManager.Stop(ctx)
}
