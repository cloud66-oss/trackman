package utils

// Step is a single running step
type Step struct {
	Metadata   map[string]string `yaml:"metadata"`
	Name       string            `yaml:"name"`
	Command    string            `yaml:"command"`
	Args       []string          `yaml:"args"`
	StopOnFail bool              `yaml:"stop_on_fail"`
}

// GetMetaData returns metadata value of the key from this Step.
// this is useful in event notifiers. It will return "" if there is
// no metadata with the given key
func (s *Step) GetMetaData(key string) string {
	if s.Metadata == nil {
		s.Metadata = make(map[string]string)
	}

	return s.Metadata[key]
}
