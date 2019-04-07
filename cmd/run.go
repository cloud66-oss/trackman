package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"

	"github.com/cloud66/trackman/sinks"

	"github.com/spf13/viper"

	"github.com/cloud66/trackman/notifiers"
	"github.com/cloud66/trackman/utils"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the given workflow",
	Run:   runExec,
}

var (
	workflowFile        string
	notificationManager *utils.NotificationManager
)

func init() {
	runCmd.Flags().StringVarP(&workflowFile, "file", "f", "", "workflow file to run")
	runCmd.Flags().DurationP("timeout", "", 10*time.Second, "global timeout unless overwritten by a step")
	runCmd.Flags().IntP("queue-size", "", 100, "usage notification queue size")

	if err := viper.BindPFlag("timeout", runCmd.Flags().Lookup("timeout")); err != nil {
		panic("cannot bind timeout")
	}
	if err := viper.BindPFlag("queue-size", runCmd.Flags().Lookup("queue-size")); err != nil {
		panic("cannot bind queue-size")
	}

	rootCmd.AddCommand(runCmd)
}

func runExec(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	notifier, err := notifiers.NewConsoleNotifier(ctx)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	sink, err := sinks.NewLogSink(ctx, notifier)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	notificationManager = utils.NewNotificationManager(ctx, notifier)
	defer notificationManager.Close(ctx)

	options := &utils.WorkflowOptions{
		NotificationManager: notificationManager,
		Sink:                sink,
	}

	workflow, err := loadWorkflow(cmd, args, options)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		fmt.Println("\nReceived an interrupt, stopping services...")
		workflow.Stop(ctx)
	}()

	err = workflow.Run(ctx)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Done")
}

func loadWorkflow(cmd *cobra.Command, args []string, options *utils.WorkflowOptions) (*utils.Workflow, error) {
	// are we sending in stream or file?
	file, err := cmd.Flags().GetString("file")
	if err != nil {
		return nil, err
	}

	var reader io.Reader
	if file == "-" {
		reader = os.Stdin
	} else {
		reader, err = os.Open(file)
		if err != nil {
			return nil, err
		}
	}

	return utils.LoadWorkflowFromReader(reader, options)
}
