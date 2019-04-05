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
	fmt.Println(event.String())

	return nil
}
