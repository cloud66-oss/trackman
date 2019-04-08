package utils

// Probe defines a checker for a step's health
type Probe struct {
	Command string   `yaml:"command"`
	Args    []string `yaml:"args"`
	Workdir string   `yaml:"workdir"`
}
