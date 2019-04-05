package utils

// Payload is what's sent over to a notifier
type Payload struct {
	EventUUID   string
	SpinnerUUID string
	Command     string
	Extras      interface{}
}
