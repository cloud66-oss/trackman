package notifiers

import (
	"context"

	"github.com/cloud66-oss/trackman/utils"
)

// ConsoleNotify writes notifications to console
func ConsoleNotify(ctx context.Context, event *utils.Event) error {
	logger, _ := utils.LoggerContext(ctx)

	switch event.Name {
	case utils.EventRunRequested:
		logger.WithField(utils.FldStep, event.Payload.Spinner.Name).Info("Starting")
	case utils.EventRunStarted:
		logger.WithField(utils.FldStep, event.Payload.Spinner.Name).Debug("Running")
	case utils.EventRunSuccess:
		logger.WithField(utils.FldStep, event.Payload.Spinner.Name).Info("Successfully finished")
	case utils.EventRunError:
		logger.WithField(utils.FldStep, event.Payload.Spinner.Name).Error("Failed to run")
	case utils.EventRunFail:
		logger.WithField(utils.FldStep, event.Payload.Spinner.Name).Errorf("Finished with error %v", event.Payload.Extras)
	case utils.EventRunTimeout:
		logger.WithField(utils.FldStep, event.Payload.Spinner.Name).Error("Timed out")
	case utils.EventRunWaitError:
		logger.WithField(utils.FldStep, event.Payload.Spinner.Name).Error("Error during wait")
	case utils.EventRunningProbe:
		logger.WithField(utils.FldStep, event.Payload.Spinner.Name).Debug("Running a probe")
	}

	return nil
}
