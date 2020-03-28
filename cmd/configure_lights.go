package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cfgLightsCmd)
}

var cfgLightsCmd = &cobra.Command{
	Use:   "cfglights",
	Short: "Configure Lights",
	Long:  "Configure Lights",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hugo Static Site Generator v0.9 -- HEAD")
	},
}
