package cmd

import (
	"log"

	"github.com/phelpsw/camera-trigger-bt-cli/connection"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List Devices",
	Long:  "List all valid device ids for command and control.",
	Run:   list,
}

func listHandler(msg interface{}) error {
	return nil
}

var listDone = make(chan struct{})

func list(cmd *cobra.Command, args []string) {

	connection.Init("", nil, debug)

	<-listDone
	log.Println("Done")
}
