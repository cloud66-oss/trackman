package utils

import "github.com/google/uuid"

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
	// EventRunningProbe announces probing
	EventRunningProbe = "run.probing"
	// EventParsingDynamicContext announces dymamic context has run and the results are being parsed
	EventParsingDynamicContext = "parse.context"
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
			EventUUID: uuid.New().String(),
			Spinner:   spinner,
			Step:      spinner.step,
			Extras:    extras,
		},
	}
}
