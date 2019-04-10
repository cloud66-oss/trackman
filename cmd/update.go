package cmd

import (
	"fmt"
	"os"

	"github.com/cloud66/trackman/utils"
	"github.com/khash/updater"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update trackman to the latest version",
	Run:   updateExec,
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func updateExec(cmd *cobra.Command, args []string) {
	update(false)

	fmt.Println("Updated")
}

func update(silent bool) {
	worker, err := updater.NewUpdater(utils.Version, &updater.Options{
		VersionURL: "https://s3.amazonaws.com/downloads.cloud66.com/trackman/VERSION",
		BinURL:     "https://s3.amazonaws.com/downloads.cloud66.com/trackman/{{OS}}_{{ARCH}}_{{VERSION}}",
		Silent:     silent,
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = worker.RunAutoUpdater()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
