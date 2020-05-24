package cmd

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/phelpsw/camera-trigger-bt-cli/boards"
	"github.com/spf13/cobra"
)

var cfgLightsCmd = &cobra.Command{
	Use:   "cfglights",
	Short: "Configure Lights",
	Long:  "Configure Lights",
	Run:   configLights,
}

var trgLightsCmd = &cobra.Command{
	Use:   "triggerlights",
	Short: "Trigger Lights",
	Long:  "Trigger Lights",
	Run:   triggerLights,
}

var (
	lightsDone = make(chan struct{})

	level         float32
	levelUpdate   bool = false
	delay         float32
	delayUpdate   bool = false
	attack        float32
	attackUpdate  bool = false
	sustain       float32
	sustainUpdate bool = false
	release       float32
	releaseUpdate bool = false

	triggered bool = false
)

func init() {
	rootCmd.AddCommand(cfgLightsCmd)

	cfgLightsCmd.Flags().Float32VarP(&level, "level", "l", 0, "Maximum lighting level 0.0 - 1.0")
	cfgLightsCmd.Flags().Float32VarP(&delay, "delay", "e", 0, "Delay in seconds before enabling lights")
	cfgLightsCmd.Flags().Float32VarP(&attack, "attack", "a", 0, "Light ramp up period in seconds")
	cfgLightsCmd.Flags().Float32VarP(&sustain, "sustain", "s", 0, "Light on period in seconds")
	cfgLightsCmd.Flags().Float32VarP(&release, "release", "r", 0, "Light ramp down period in seconds")

	rootCmd.AddCommand(trgLightsCmd)
}

func configLightsHandler(b interface{}) error {
	switch b.(type) {
	case *boards.Light:
		m := b.(*boards.Light)

		fmt.Println(m.Level(), m.Delay(), m.Attack(), m.Sustain(), m.Release())

		if levelUpdate && m.Level() != level {
			err := m.SetLevel(level, false)
			if err != nil {
				return err
			}
		}

		if delayUpdate && m.Delay() != delay {
			err := m.SetDelay(delay, false)
			if err != nil {
				return err
			}
		}

		if attackUpdate && m.Attack() != attack {
			err := m.SetAttack(attack, false)
			if err != nil {
				return err
			}
		}

		if sustainUpdate && m.Sustain() != sustain {
			err := m.SetSustain(sustain, false)
			if err != nil {
				return err
			}
		}

		if releaseUpdate && m.Release() != release {
			err := m.SetRelease(release, false)
			if err != nil {
				return err
			}
		}

		err := m.Sync()
		if err != nil {
			return err
		}

		if m.IsSynced() {
			close(lightsDone)
		}

	default:
		return fmt.Errorf("unknown type %+v", reflect.TypeOf(b))
	}
	return nil
}

func triggerLightsHandler(b interface{}) error {
	switch b.(type) {
	case *boards.Light:
		if !triggered {
			m := b.(*boards.Light)

			err := m.Trigger(0)
			triggered = true
			if err != nil {
				return err
			}
		} else {
			time.Sleep(1)
			close(lightsDone)
		}

	default:
		return fmt.Errorf("unknown type %+v", reflect.TypeOf(b))
	}

	return nil
}

func configLights(cmd *cobra.Command, args []string) {
	levelUpdate = cmd.Flags().Changed("level")
	delayUpdate = cmd.Flags().Changed("delay")
	attackUpdate = cmd.Flags().Changed("attack")
	sustainUpdate = cmd.Flags().Changed("sustain")
	releaseUpdate = cmd.Flags().Changed("release")

	m := boards.Light{}

	err := m.Init(deviceID, debug)
	if err != nil {
		log.Panicln(err)
		return
	}

	m.SetUpdateCallback(configLightsHandler)

	<-lightsDone
	log.Println("Done")
}

func triggerLights(cmd *cobra.Command, args []string) {
	m := boards.Light{}

	err := m.Init(deviceID, debug)
	if err != nil {
		log.Panicln(err)
		return
	}

	m.SetUpdateCallback(triggerLightsHandler)

	<-lightsDone
	log.Println("Done")
}
