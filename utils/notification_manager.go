package utils

import (
	"context"
	"fmt"
)

// NotificationManager is a singleton that makes sure notifications are sent
type NotificationManager struct {
	queue    chan *Event
	notifier EventNotifier
	stopChan chan struct{}
}

// NewNotificationManager creates a new NotificationManager
func NewNotificationManager(ctx context.Context, eventNotifier EventNotifier) *NotificationManager {
	return &NotificationManager{
		queue:    make(chan *Event, 100),
		notifier: eventNotifier,
		stopChan: make(chan struct{}, 1),
	}
}

// Stop stops the event cycle
func (n *NotificationManager) Stop(ctx context.Context) {
	n.stopChan <- struct{}{}
}

// Close should be called all the time at the end of a run
func (n *NotificationManager) Close(ctx context.Context) {
	close(n.queue)
}

// Start starts the cycle for the notifier
func (n *NotificationManager) Start(ctx context.Context) error {
	go func() {
		defer close(n.stopChan)
		for {
			select {
			default:
				event := <-n.queue
				err := n.notifier.Notify(ctx, *event)
				if err != nil {
					// we have an error
					// TODO: log it for now
					fmt.Println(err)
				}

			case <-n.stopChan:
				return
			}
		}
	}()

	return nil
}

// Push is a failsafe method to push any notification to the queue to be sent
// this might block if we hit the limits on the queue but that's deliberate
func (n *NotificationManager) Push(ctx context.Context, event *Event) {
	n.queue <- event
}
