package utils

// Probe defines a checker for a Step's health
type Probe struct {
	Command string `yaml:"command" json:"command"`
	Workdir string `yaml:"workdir" json:"workdir"`

	cmd  string
	args []string
}
