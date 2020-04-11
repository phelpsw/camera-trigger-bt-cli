package cmd

import (
	"fmt"
	"log"

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

// TODO: Add more values to BT messages

// TODO: Investigate weird clock rolling issue, bcd issue?

// TODO: Make this handle the generic board type rather than something specific
// to the motion sensor
func monitorHandler(m interface{}) error {
	switch m.(type) {
	case boards.Motion:
		b := m.(boards.Motion)
		fmt.Printf("Motion %d / %d\n", b.Motion(), b.MotionThreshold())
	}

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
