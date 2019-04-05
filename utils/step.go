package utils

// Step is a single running step
type Step struct {
	Metadata map[string]string
	Name     string
	Command  string
}
