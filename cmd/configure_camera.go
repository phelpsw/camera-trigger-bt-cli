package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cameraCmd)
}

var cameraCmd = &cobra.Command{
	Use:   "triggercamera",
	Short: "Trigger Camera",
	Long:  "Trigger Camera",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hugo Static Site Generator v0.9 -- HEAD")
	},
}
