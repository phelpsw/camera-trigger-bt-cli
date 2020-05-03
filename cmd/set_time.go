package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/phelpsw/camera-trigger-bt-cli/boards"
	"github.com/phelpsw/camera-trigger-bt-cli/messages"
	"github.com/spf13/cobra"
)

func init() {
	setTimeCmd.Flags().BoolVarP(&utc, "utc", "u", false, "Set flag UTC rather than local time")
	rootCmd.AddCommand(setTimeCmd)
}

var setTimeCmd = &cobra.Command{
	Use:   "settime",
	Short: "Set the time",
	Long:  "Set the time",
	Run:   setTime,
}

var (
	timeDone      = make(chan struct{})
	utc      bool = false
)

func setTime(cmd *cobra.Command, args []string) {
	m := boards.Basic{}

	err := m.Init(deviceID, debug)
	if err != nil {
		log.Panicln(err)
		return
	}

	for !m.IsConnected() {
	}

	var ts time.Time
	if utc {
		ts = time.Now().UTC()
	} else {
		ts = time.Now()
	}

	cal := messages.Calendar{
		Seconds:    uint8(ts.Second()),
		Minutes:    uint8(ts.Minute()),
		Hours:      uint8(ts.Hour()),
		DayOfWeek:  uint8(ts.Weekday()),
		DayOfMonth: uint8(ts.Day()),
		Month:      uint8(ts.Month()),
		Year:       uint16(ts.Year()),
	}
	fmt.Printf("%+v\n", cal)

	m.SetTime(cal)

	time.Sleep(5 * time.Second)

	log.Println("Done")
}
