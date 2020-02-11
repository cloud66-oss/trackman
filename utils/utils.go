package utils

import (
	"context"
	"os"

	"github.com/fatih/color"
)

// ExpandEnvVars replaces any reference to environment variables with the OS envs
func ExpandEnvVars(ctx context.Context, value string) (string, error) {
	if value == "" {
		return "", nil
	}

	expandedCommand := os.ExpandEnv(value)
	return expandedCommand, nil
}

// PrintError prints an error to the console in red
func PrintError(format string, a ...interface{}) {
	color.Red(format, a...)
}
