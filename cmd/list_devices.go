package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

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

func list(cmd *cobra.Command, args []string) {
	connection.Init("", nil, debug)

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()

	<-done
	connection.Stop()
	log.Println("Done")

}
