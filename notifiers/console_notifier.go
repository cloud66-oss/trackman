package notifiers

import (
	"context"

	"github.com/cloud66/trackman/utils"
)

// ConsoleNotify writes notifications to console
func ConsoleNotify(ctx context.Context, event *utils.Event) error {
	logger := utils.GetLogger(ctx)

	switch event.Name {
	case utils.EventRunRequested:
		logger.WithField(utils.FldStep, event.Payload.Step.Name).Info("Starting")
	case utils.EventRunStarted:
		logger.WithField(utils.FldStep, event.Payload.Step.Name).Info("Running")
	case utils.EventRunSuccess:
		logger.WithField(utils.FldStep, event.Payload.Step.Name).Info("Successfully finished")
	case utils.EventRunError:
		logger.WithField(utils.FldStep, event.Payload.Step.Name).Error("Failed to run")
	case utils.EventRunFail:
		logger.WithField(utils.FldStep, event.Payload.Step.Name).Errorf("Finished with error %v", event.Payload.Extras)
	case utils.EventRunTimeout:
		logger.WithField(utils.FldStep, event.Payload.Step.Name).Error("Timed out")
	case utils.EventRunWaitError:
		logger.WithField(utils.FldStep, event.Payload.Step.Name).Error("Error during wait")
	}

	return nil
}
