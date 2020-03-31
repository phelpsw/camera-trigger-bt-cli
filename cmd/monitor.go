package cmd

import (
	"log"

	"github.com/phelpsw/camera-trigger-bt-cli/connection"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(monitorCmd)
}

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Pretty Print all status messages from the device",
	Long:  "Pretty Print all status messages from the device",
	Run:   monitor,
}

var monitorDone = make(chan struct{})

func monitorHandler(msg interface{}) error {
	return nil
}

func monitor(cmd *cobra.Command, args []string) {

	connection.Init(deviceID, monitorHandler, debug)

	<-monitorDone
	log.Println("Done")
}
