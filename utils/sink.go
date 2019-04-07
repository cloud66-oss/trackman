package utils

// Sink is where workflow / spinners send their output
type Sink struct {
	Notifier EventNotifier
}
