package cmd

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/phelpsw/camera-trigger-bt-cli/boards"
	"github.com/spf13/cobra"
)

var cfgCameraCmd = &cobra.Command{
	Use:   "cfgcamera",
	Short: "Configure Camera",
	Long:  "Configure Camera",
	Run:   configCamera,
}

var trgCameraCmd = &cobra.Command{
	Use:   "triggercamera",
	Short: "Trigger Camera",
	Long:  "Trigger Camera",
	Run:   triggerCamera,
}

var (
	cameraDone = make(chan struct{})

	duration       float32
	durationUpdate bool = false

	triggeredCamera bool = false
)

func init() {
	cfgCameraCmd.Flags().Float32VarP(&duration, "duration", "p", 0, "Duration of camera video capture")

	rootCmd.AddCommand(cfgCameraCmd)
	rootCmd.AddCommand(trgCameraCmd)
}

func configCameraHandler(b interface{}) error {
	switch b.(type) {
	case *boards.Camera:
		m := b.(*boards.Camera)

		fmt.Println(m.Duration())

		if durationUpdate && m.Duration() != duration {
			err := m.SetDuration(duration, false)
			if err != nil {
				return err
			}
		}

		err := m.Sync()
		if err != nil {
			return err
		}

		if m.IsSynced() {
			close(cameraDone)
		}

	default:
		return fmt.Errorf("unknown type %+v", reflect.TypeOf(b))
	}
	return nil
}

func triggerCameraHandler(b interface{}) error {
	switch b.(type) {
	case *boards.Camera:
		if !triggeredCamera {
			m := b.(*boards.Camera)

			err := m.Trigger(0)
			triggeredCamera = true
			if err != nil {
				return err
			}
		} else {
			time.Sleep(1)
			close(cameraDone)
		}

	default:
		return fmt.Errorf("unknown type %+v", reflect.TypeOf(b))
	}

	return nil
}

func configCamera(cmd *cobra.Command, args []string) {
	durationUpdate = cmd.Flags().Changed("duration")
	m := boards.Camera{}

	err := m.Init(deviceID, debug)
	if err != nil {
		log.Panicln(err)
		return
	}

	m.SetUpdateCallback(configCameraHandler)

	<-cameraDone
	log.Println("Done")
}

func triggerCamera(cmd *cobra.Command, args []string) {
	m := boards.Camera{}

	err := m.Init(deviceID, debug)
	if err != nil {
		log.Panicln(err)
		return
	}

	m.SetUpdateCallback(triggerCameraHandler)

	<-cameraDone
	log.Println("Done")
}
