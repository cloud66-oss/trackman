package notifiers

import (
	"context"
	"fmt"

	"github.com/cloud66/trackman/utils"
)

// ConsoleNotifier dumps events onto the console
type ConsoleNotifier struct{}

// Notify implements the interface
func (c *ConsoleNotifier) Notify(ctx context.Context, event utils.Event) error {
	switch event.Name {
	case utils.EventRunRequested:
		fmt.Printf("[%s] Starting\n", event.Payload.Step.Name)
	case utils.EventRunStarted:
		fmt.Printf("[%s] Running\n", event.Payload.Step.Name)
	case utils.EventRunSuccess:
		fmt.Printf("[%s] Successfully finished\n", event.Payload.Step.Name)
	case utils.EventRunError:
		fmt.Printf("[%s] Failed to run\n", event.Payload.Step.Name)
	case utils.EventRunFail:
		fmt.Printf("[%s] Finished with error %v\n", event.Payload.Step.Name, event.Payload.Extras)
	case utils.EventRunTimeout:
		fmt.Printf("[%s] Timedout\n", event.Payload.Step.Name)
	case utils.EventRunWaitError:
		fmt.Printf("[%s] Error during wait\n", event.Payload.Step.Name)
	}

	return nil
}

// NewConsoleNotifier creates a new console notifier
func NewConsoleNotifier(ctx context.Context) (utils.EventNotifier, error) {
	return &ConsoleNotifier{}, nil
}
