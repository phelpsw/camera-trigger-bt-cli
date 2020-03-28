package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(lightsCmd)
}

var lightsCmd = &cobra.Command{
	Use:   "cmdlights",
	Short: "Trigger Lights",
	Long:  "Trigger Lights",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hugo Static Site Generator v0.9 -- HEAD")
	},
}
