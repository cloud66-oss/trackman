package utils

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var fileRegister []*os.File
var fileRegisterSignal *sync.Mutex

func init() {
	fileRegisterSignal = &sync.Mutex{}
	fileRegister = make([]*os.File, 0)
}

// LogDefinition is used to define where a logger should log to
type LogDefinition struct {
	Level       string `yaml:"level" json:"level"`
	Type        string `yaml:"type" json:"type"`
	Format      string `yaml:"format" json:"format"`
	Destination string `yaml:"destination" json:"destination"`
}

// LoggingContext is a structure that holds workflow and step
// and is used to determine log configuration
type LoggingContext struct {
	Workflow *Workflow
	Step     *Step
}

func addFile(file *os.File) {
	fileRegisterSignal.Lock()
	defer fileRegisterSignal.Unlock()

	fileRegister = append(fileRegister, file)
}

// CloseAllFiles closes all files in fileRegister
func CloseAllFiles() {
	fileRegisterSignal.Lock()
	defer fileRegisterSignal.Unlock()

	for _, file := range fileRegister {
		file.Close()
	}
}

// NewLoggingContext creates a new logging context. Da
func NewLoggingContext(workflow *Workflow, step *Step) *LoggingContext {
	return &LoggingContext{
		Workflow: workflow,
		Step:     step,
	}
}

// Parse parses the given value within this context.
func (l *LoggingContext) parse(value string) (string, error) {
	temp, err := template.New("filename").Parse(value)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	err = temp.Execute(&tpl, l)
	if err != nil {
		return "", err
	}

	return tpl.String(), nil
}

// DefaultLogDefinition returns a LogDefintion based on the given base
// if the base is nil it creates a new one
// if not nil, it fills the empty values with the defaults
func DefaultLogDefinition(baseDefinition *LogDefinition) *LogDefinition {
	var definition *LogDefinition

	if baseDefinition == nil {
		definition = &LogDefinition{}
	} else {
		definition = baseDefinition
	}

	if definition.Type == "" {
		definition.Type = viper.GetString("log-type")
	}
	if definition.Destination == "" {
		definition.Destination = viper.GetString("log-file")
	}
	if definition.Format == "" {
		definition.Format = viper.GetString("log-format")
	}
	if definition.Level == "" {
		definition.Level = viper.GetString("log-level")
	}

	return definition
}

// NewLogger creates a new logger instance and sets the right log level based
func NewLogger(baseDefinition *LogDefinition, loggingContext *LoggingContext) (*logrus.Logger, error) {
	ctx := context.Background()
	definition := DefaultLogDefinition(baseDefinition)

	logger := logrus.New()

	if definition.Type == "stdout" {
		logger.SetOutput(os.Stdout)
	} else if definition.Type == "stderr" {
		logger.SetOutput(os.Stderr)
	} else if definition.Type == "discard" {
		logger.SetOutput(ioutil.Discard)
	} else if definition.Type == "file" {
		var filename string
		var err error
		if filename, err = loggingContext.parse(definition.Destination); err != nil {
			return nil, err
		}
		if filename, err = ExpandEnvVars(ctx, filename); err != nil {
			return nil, err
		}

		file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
		if err != nil {
			return nil, err
		}

		addFile(file)
		logger.SetOutput(file)
	} else {
		return nil, fmt.Errorf("invalid log type %s", definition.Type)
	}

	level, err := logrus.ParseLevel(definition.Level)
	if err != nil {
		return nil, err
	}

	logger.SetLevel(level)

	if definition.Format == "text" {
		logger.Formatter = &logrus.TextFormatter{}
	} else if definition.Format == "json" {
		logger.Formatter = &logrus.JSONFormatter{}
	} else {
		return nil, fmt.Errorf("invalid log format %s", definition.Format)
	}

	return logger, nil
}
