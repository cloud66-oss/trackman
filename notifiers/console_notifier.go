package notifiers

import (
	"context"

	"github.com/cloud66/trackman/utils"
	"github.com/sirupsen/logrus"
)

// ConsoleNotify writes notifications to console
func ConsoleNotify(ctx context.Context, event *utils.Event) error {
	var logger *logrus.Logger
	if ctx.Value(utils.CtxLogger) != nil {
		logger = ctx.Value(utils.CtxLogger).(*logrus.Logger)
	} else {
		logger = logrus.New()
	}

	switch event.Name {
	case utils.EventRunRequested:
		logger.WithField("process", event.Payload.Step.Name).Info("Starting")
	case utils.EventRunStarted:
		logger.WithField("process", event.Payload.Step.Name).Info("Running")
	case utils.EventRunSuccess:
		logger.WithField("process", event.Payload.Step.Name).Info("Successfully finished")
	case utils.EventRunError:
		logger.WithField("process", event.Payload.Step.Name).Error("Failed to run")
	case utils.EventRunFail:
		logger.WithField("process", event.Payload.Step.Name).Errorf("Finished with error %v", event.Payload.Extras)
	case utils.EventRunTimeout:
		logger.WithField("process", event.Payload.Step.Name).Error("Timed out")
	case utils.EventRunWaitError:
		logger.WithField("process", event.Payload.Step.Name).Error("Error during wait")
	}

	return nil
}
