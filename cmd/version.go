package cmd

import (
	"fmt"

	"github.com/cloud66/trackman/utils"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Trackman",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Trackman")
		fmt.Println("(c) 2019 Cloud66 Inc.")
		fmt.Println("Trackman is a commandline and library to run multiple commands as a workflow")
		fmt.Printf("%s/%s\n", utils.Channel, utils.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
