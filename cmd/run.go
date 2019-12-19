package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/cloud66-oss/trackman/notifiers"
	"github.com/cloud66-oss/trackman/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	runCmd.Flags().IntP("concurrency", "", runtime.NumCPU()-1, "maximum number of concurrent steps to run")
	runCmd.Flags().BoolP("yes", "y", false, "Answer Yes to all confirmation questions")

	_ = viper.BindPFlag("timeout", runCmd.Flags().Lookup("timeout"))
	_ = viper.BindPFlag("concurrency", runCmd.Flags().Lookup("concurrency"))
	_ = viper.BindPFlag("confirm.yes", runCmd.Flags().Lookup("yes"))

	rootCmd.AddCommand(runCmd)
}

func runExec(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	options := &utils.WorkflowOptions{
		Notifier:    notifiers.ConsoleNotify,
		Concurrency: viper.GetInt("concurrency"),
		Timeout:     viper.GetDuration("timeout"),
	}

	workflow, err := loadWorkflow(ctx, args, options, cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	logger, err := utils.NewLogger(workflow.Logger, utils.NewLoggingContext(workflow, nil))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err, stepErrors := workflow.Run(ctx)
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}

	if stepErrors != nil {
		// this is already logged, just get out
		logger.Error("Done with errors")
		os.Exit(1)
	} else {
		logger.Info("Done")
	}
}

func loadWorkflow(ctx context.Context, args []string, options *utils.WorkflowOptions, cmd *cobra.Command) (*utils.Workflow, error) {
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

	return utils.LoadWorkflowFromReader(ctx, options, reader)
}
