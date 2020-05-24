package cmd

import (
	"fmt"
	"log"
	"reflect"

	"github.com/phelpsw/camera-trigger-bt-cli/boards"
	"github.com/spf13/cobra"
)

var (
	cfgMotionCmd = &cobra.Command{
		Use:   "cfgmotion",
		Short: "Configure Motion Sensor",
		Long:  "Configure Motion Sensor",
		Run:   configMotion,
	}

	motionDone = make(chan struct{})

	thresh         float32
	threshUpdate   bool = false
	luxLow         float32
	luxLowUpdate   bool = false
	luxHigh        float32
	luxHighUpdate  bool = false
	cooldown       float32
	cooldownUpdate bool = false
)

func init() {
	cfgMotionCmd.Flags().Float32VarP(&thresh, "motion", "m", 0, "set the motion trigger threshold")
	cfgMotionCmd.Flags().Float32VarP(&luxLow, "luxlow", "l", 0, "set the lux low threshold")
	cfgMotionCmd.Flags().Float32VarP(&luxHigh, "luxhigh", "t", 0, "set the lux high threshold")
	cfgMotionCmd.Flags().Float32VarP(&cooldown, "cooldown", "c", 0, "set cooldown period")

	rootCmd.AddCommand(cfgMotionCmd)
}

func configMotionHandler(b interface{}) error {
	switch b.(type) {
	case *boards.Motion:
		m := b.(*boards.Motion)

		if threshUpdate && m.MotionThreshold() != thresh {
			err := m.SetMotionThreshold(thresh, false)
			if err != nil {
				return err
			}
		}

		if luxLowUpdate && m.LuxLowThreshold() != luxLow {
			err := m.SetLuxLowThreshold(luxLow, false)
			if err != nil {
				return err
			}
		}

		if luxHighUpdate && m.LuxHighThreshold() != luxHigh {
			err := m.SetLuxHighThreshold(luxHigh, false)
			if err != nil {
				return err
			}
		}

		if cooldownUpdate && m.Cooldown() != cooldown {
			err := m.SetCooldown(cooldown, false)
			if err != nil {
				return err
			}
		}

		err := m.Sync()
		if err != nil {
			return err
		}

		if m.IsSynced() {
			close(motionDone)
		}
	default:
		return fmt.Errorf("unknown type %+v", reflect.TypeOf(b))
	}
	return nil
}

func configMotion(cmd *cobra.Command, args []string) {
	threshUpdate = cmd.Flags().Changed("motion")
	luxLowUpdate = cmd.Flags().Changed("luxlow")
	luxHighUpdate = cmd.Flags().Changed("luxhigh")
	cooldownUpdate = cmd.Flags().Changed("cooldown")

	m := boards.Motion{}

	err := m.Init(deviceID, debug)
	if err != nil {
		log.Panicln(err)
		return
	}

	m.SetUpdateCallback(configMotionHandler)

	<-motionDone
	log.Println("Done")
}
