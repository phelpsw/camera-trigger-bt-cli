package cmd

import (
	"time"

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

func list(cmd *cobra.Command, args []string) {
	var conn connection.Connection
	conn.Scan(5 * time.Second)
}
