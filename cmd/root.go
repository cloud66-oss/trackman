package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/cloud66/trackman/utils"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:              "trackman",
	Short:            "Trackman is a tool to run commands in a sequence",
	PersistentPreRun: checkForUpdates,
}

var (
	cfgFile string
	// UpdateDone makes sure background updater is done before the app is closed
	UpdateDone *sync.WaitGroup
)

func init() {
	UpdateDone = &sync.WaitGroup{}
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cobra.yaml)")
	rootCmd.PersistentFlags().String("log-level", "info", "log level. Use debug to see process output")
	rootCmd.PersistentFlags().Bool("no-update", false, "turn off auto update")

	_ = viper.BindPFlag("log-level", rootCmd.PersistentFlags().Lookup("log-level"))
	_ = viper.BindPFlag("no-update", rootCmd.PersistentFlags().Lookup("no-update"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(".")
		viper.AddConfigPath(filepath.Join(home, ".trackman"))
		viper.AddConfigPath("/etc/trackman")
		viper.SetConfigName("config")
	}

	_ = viper.ReadInConfig()
}

func checkForUpdates(cmd *cobra.Command, args []string) {
	if utils.Channel != "dev" && cmd.Name() != "update" && cmd.Name() != "version" && !viper.GetBool("no-update") {
		go func() {
			UpdateDone.Add(1)
			defer UpdateDone.Done()

			update(true)
		}()
	}
}

// Execute main cobra entry point
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
