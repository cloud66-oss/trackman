package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"gopkg.in/yaml.v2"
)

// DynamicContextBuilder runs an external process, fetches the results and parses them as an object
// which then can be used by the workflow or steps to render the workflow dynamically
type DynamicContextBuilder struct {
	Command string         `yaml:"command" json:"command"`
	Workdir string         `yaml:"workdir" json:"workdir"`
	Format  string         `yaml:"format" json:"format"`
	Timeout *time.Duration `yaml:"timeout" json:"timeout"`
	Env     []string       `yaml:"env" json:"env"`

	spinner  *Spinner
	workflow *Workflow
}

// Run runs the spinner for this dcb, fetches the results and parses them based on the format and returns them as a hash
func (dcb *DynamicContextBuilder) Run(ctx context.Context) (map[string]interface{}, error) {
	if dcb.spinner == nil {
		panic("no spinner defined for DynamicContextBuilder")
	}

	var buf bytes.Buffer
	ctx = context.WithValue(ctx, CtxOutWriter, &buf)
	err := dcb.spinner.Run(ctx)
	if err != nil {
		return nil, err
	}
	dcb.spinner.push(ctx, NewEvent(dcb.spinner, EventParsingDynamicContext, nil))
	result := make(map[string]interface{})
	if dcb.Format == "json" {
		err = json.Unmarshal(buf.Bytes(), &result)
		if err != nil {
			dcb.workflow.logger.WithField(FldContextBuilder, "ContextBuilder").Error(err)
			return nil, err
		}
	} else if dcb.Format == "yaml" {
		err = yaml.Unmarshal(buf.Bytes(), &result)
		if err != nil {
			dcb.workflow.logger.WithField(FldContextBuilder, "ContextBuilder").Error(err)
			return nil, err
		}
	} else {
		// should never happen unless validate is not called
		return nil, errors.New("invalid context builder format")
	}

	if dcb.workflow.logger.Level == logrus.DebugLevel {
		buff, err := yaml.Marshal(&result)
		if err != nil {
			return nil, err
		}
		dcb.workflow.logger.WithField(FldContextBuilder, "ContextBuilder").Debug(string(buff))
	}

	return result, nil
}

// Validate validates the DCB's input values
func (dcb *DynamicContextBuilder) Validate(ctx context.Context) {
	if dcb.workflow == nil {
		panic("no workflow")
	}

	if dcb.Format != "" && dcb.Format != "json" && dcb.Format != "yaml" && dcb.Format != "yml" {
		PrintError("only json and yaml are accepted as Context Builder Format (default is json)")
		os.Exit(1)
	}

	if dcb.Format == "" {
		dcb.Format = "json"
	}

	if dcb.Format == "yml" {
		dcb.Format = "yaml"
	}
}
