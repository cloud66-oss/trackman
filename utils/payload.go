package utils

// Payload is what's sent over to a notifier
type Payload struct {
	EventUUID string
	Spinner   *Spinner
	Step      Step
	Extras    interface{}
}
