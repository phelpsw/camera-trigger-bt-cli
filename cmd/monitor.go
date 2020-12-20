package cmd

import (
	"fmt"
	"log"
	"reflect"

	"github.com/phelpsw/camera-trigger-bt-cli/boards"
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

var motionBoard boards.Motion
var lightBoard boards.Light
var cameraBoard boards.Camera

func monitorHandler(m interface{}) error {
	switch m.(type) {
	case *boards.Basic:
		b := m.(*boards.Basic)
		if b.GetType() == reflect.TypeOf(boards.Motion{}) {
			// Initialize motion board type from basic board
			motionBoard.InitFromBasic(b)
			motionBoard.SetUpdateCallback(monitorHandler)
		} else if b.GetType() == reflect.TypeOf(boards.Light{}) {
			// Initialize motion board type from basic board
			lightBoard.InitFromBasic(b)
			lightBoard.SetUpdateCallback(monitorHandler)
		} else if b.GetType() == reflect.TypeOf(boards.Camera{}) {
			// Initialize motion board type from basic board
			cameraBoard.InitFromBasic(b)
			cameraBoard.SetUpdateCallback(monitorHandler)
		} else {
			fmt.Printf("Unknown type, %v\n", b.GetType())
		}
	case *boards.Motion:
		b := m.(*boards.Motion)
		fmt.Printf("Motion Sensor\n")
		fmt.Printf("  Motion: %.3f Thresh %.3f\n", b.Motion(), b.MotionThreshold())
		fmt.Printf("  Light: %.2f lux\n", b.Lux())
		fmt.Printf("    Thresh Low: %.2f High %.2f\n", b.LuxLowThreshold(), b.LuxHighThreshold())
		fmt.Printf("  Transmit Cooldown %.1f sec\n", b.Cooldown())
		fmt.Printf("  CPU Temp %.2f degC\n", b.Temperature())
		fmt.Printf("  Log Count: %d\n", b.LogEntries())

	case *boards.Light:
		b := m.(*boards.Light)
		fmt.Printf("Light Controller\n")
		fmt.Printf("  Brightness Level %f\n", b.Level())
		fmt.Printf("    Delay %.2f sec", b.Delay())
		fmt.Printf("    Attack %.2f sec", b.Attack())
		fmt.Printf("    Sustain %.2f sec", b.Sustain())
		fmt.Printf("    Release %.2f sec", b.Release())
		fmt.Printf("  CPU Temp %.2f degC\n", b.Temperature())
		fmt.Printf("  Log Count: %d\n", b.LogEntries())

	case *boards.Camera:
		b := m.(*boards.Camera)
		fmt.Printf("Camera Controller\n")
		fmt.Printf("  Video Duration %.2f sec\n", b.Duration())
		fmt.Printf("  Battery %.2f V\n", b.Voltage())
		fmt.Printf("  CPU Temp %.2f degC\n", b.Temperature())
		fmt.Printf("  Log Count: %d\n", b.LogEntries())
	}

	fmt.Printf("\n")

	return nil
}

func monitor(cmd *cobra.Command, args []string) {
	var done = make(chan struct{})

	m := boards.Basic{}

	err := m.Init(deviceID, debug)
	if err != nil {
		log.Panicln(err)
		return
	}

	m.SetUpdateCallback(monitorHandler)

	<-done
	log.Println("Done")
}
