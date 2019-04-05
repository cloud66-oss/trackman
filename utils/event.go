package utils

import (
	"fmt"

	"github.com/google/uuid"
)

const (
	// EventRunRequested run requested
	EventRunRequested = "run.requested"
	// EventRunStarted run started
	EventRunStarted = "run.started"
	// EventRunError start failed
	EventRunError = "run.error"
	// EventRunFail ran but the command failed
	EventRunFail = "run.fail"
	// EventRunWaitError ran but failed on wait
	EventRunWaitError = "run.wait.error"
	// EventRunSuccess ran with success
	EventRunSuccess = "run.success"
	// EventRunTimeout run timed out
	EventRunTimeout = "run.timeout"
)

// Event is a simple event
type Event struct {
	Name    string
	Payload Payload
}

// NewEvent creates a new event
func NewEvent(spinner *Spinner, name string, extras interface{}) *Event {
	return &Event{
		Name: name,
		Payload: Payload{
			EventUUID:   uuid.New().String(),
			SpinnerUUID: spinner.uuid,
			Command:     spinner.cmd,
			Extras:      extras,
		},
	}
}

func (e *Event) String() string {
	return fmt.Sprintf("Event: %s, Spinner: %s, Extras: %v", e.Payload.EventUUID, e.Payload.SpinnerUUID, e.Payload.Extras)
}
