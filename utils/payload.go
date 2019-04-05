package utils

// Payload is what's sent over to a notifier
type Payload struct {
	EventUUID   string
	SpinnerUUID string
	Step        Step
	Extras      interface{}
}
