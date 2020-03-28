package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cfgMotionCmd)
}

var cfgMotionCmd = &cobra.Command{
	Use:   "cfgmotion <motion-threshold> <light-threshold>",
	Short: "Configure Motion Sensor",
	Long:  "Configure Motion Sensor",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Echo: " + strings.Join(args, " "))
	},
}
