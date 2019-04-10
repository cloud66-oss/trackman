package utils

// Probe defines a checker for a Step's health
type Probe struct {
	Command string   `yaml:"command"`
	Args    []string `yaml:"args"`
	Workdir string   `yaml:"workdir"`
}
