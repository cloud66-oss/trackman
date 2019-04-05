package utils

import "context"

// EventNotifier is an interface to notify an external entity about
// Spinner events
type EventNotifier interface {
	Notify(ctx context.Context, event Event) error
}
