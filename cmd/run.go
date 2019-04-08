package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

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
	workflowFile string
)

func init() {
	runCmd.Flags().StringVarP(&workflowFile, "file", "f", "", "workflow file to run")
	runCmd.Flags().DurationP("timeout", "", 10*time.Second, "global timeout unless overwritten by a step")

	if err := viper.BindPFlag("timeout", runCmd.Flags().Lookup("timeout")); err != nil {
		panic("cannot bind timeout")
	}

	rootCmd.AddCommand(runCmd)
}

func runExec(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	logger := utils.GetLogger(ctx)
	ctx = context.WithValue(ctx, utils.CtxLogger, logger)

	options := &utils.WorkflowOptions{
		Notifier: notifiers.ConsoleNotify,
	}

	workflow, err := loadWorkflow(cmd, args, options)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = workflow.Run(ctx)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	logger.Info("Done")
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
