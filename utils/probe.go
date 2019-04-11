package utils

// Probe defines a checker for a Step's health
type Probe struct {
	Command string `yaml:"command"`
	Workdir string `yaml:"workdir"`

	cmd  string
	args []string
}
