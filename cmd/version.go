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
		fmt.Printf("%s:%s\n", utils.Version, utils.Revision)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
