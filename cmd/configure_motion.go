package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/phelpsw/camera-trigger-bt-cli/connection"
	"github.com/phelpsw/camera-trigger-bt-cli/messages"
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
	Run:   configMotion,
}

var configMotionDone = make(chan bool)
var commanded bool = false

func configMotionHandler(msg interface{}) error {
	// TODO: Validate a motion status type message was handled
	fmt.Printf("%+v\n", msg)

	if connection.IsConnected() && !commanded {
		msg := messages.NewMotionSensorConfigMessage(1000, 100.0)
		err := connection.WriteMessage(msg)
		if err != nil {
			return err
		}
		commanded = true
	}

	// TODO: Validate received config matches expected, if so
	close(configMotionDone)
	return nil
}

func configMotion(cmd *cobra.Command, args []string) {
	fmt.Println("Echo: " + strings.Join(args, " "))
	connection.Init(deviceID, configMotionHandler, debug)

	<-configMotionDone
	log.Println("Done")
}
