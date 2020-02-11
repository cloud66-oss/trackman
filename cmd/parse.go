package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cloud66-oss/trackman/notifiers"
	"github.com/cloud66-oss/trackman/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

var parseCmd = &cobra.Command{
	Use:   "parse",
	Short: "Parse the workflow and print the output",
	Run:   parseExec,
}

var (
	parsingWorkflowFile string
)

func init() {
	parseCmd.Flags().StringVarP(&parsingWorkflowFile, "file", "f", "", "workflow file to parse")
	parseCmd.Flags().DurationP("timeout", "", 10*time.Second, "global timeout unless overwritten by a step")

	_ = viper.BindPFlag("timeout", parseCmd.Flags().Lookup("timeout"))

	rootCmd.AddCommand(parseCmd)
}

func parseExec(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	options := &utils.WorkflowOptions{
		Notifier:    notifiers.ConsoleNotify,
		Concurrency: 1,
		Timeout:     viper.GetDuration("timeout"),
	}

	workflow, err := loadWorkflow(ctx, args, options, cmd)
	if err != nil {
		utils.PrintError(err.Error())
		os.Exit(1)
	}

	for _, step := range workflow.Steps {
		if err = step.EnrichStep(ctx); err != nil {
			utils.PrintError(err.Error())
			os.Exit(1)
		}
	}

	buff, err := yaml.Marshal(&workflow)
	if err != nil {
		utils.PrintError(err.Error())
		os.Exit(1)
	}

	fmt.Println(string(buff))
}
