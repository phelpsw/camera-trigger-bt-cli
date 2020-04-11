package cmd

import (
	"github.com/spf13/cobra"
)

var (
	// Used for flags.
	cfgFile     string
	userLicense string
	deviceID    string
	debug       bool

	rootCmd = &cobra.Command{
		Use:   "bluetooth-test",
		Short: "A generator for Cobra based Applications",
		Long: `Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&deviceID, "device", "d", "", "Bluetooth device ID")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Set flag for debug messages")
}

func initConfig() {
	return
}
