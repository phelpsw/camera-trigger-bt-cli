package cmd

import (
	"fmt"
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

// TODO: Setup message specific handler deal like
// the gatt Device type

// TODO: Increase BT message freq

// TODO: Add more values to BT messages

// TODO: Investigate weird clock rolling issue, bcd issue?

// TODO: Double check motion u16 and lux float are being encoded properly

func monitorHandler(msg interface{}) error {
	if msg != nil {
		fmt.Printf("%+v\n", msg)
	}
	return nil
}

func monitor(cmd *cobra.Command, args []string) {

	connection.Init(deviceID, monitorHandler, debug)

	<-monitorDone
	log.Println("Done")
}
