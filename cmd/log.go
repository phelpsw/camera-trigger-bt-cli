package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/phelpsw/camera-trigger-bt-cli/boards"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(dumpLogCmd)
	rootCmd.AddCommand(resetLogCmd)
}

var dumpLogCmd = &cobra.Command{
	Use:   "logdump",
	Short: "Pretty Print all log messages from the device",
	Long:  "Pretty Print all log messages from the device",
	Run:   dumpLog,
}

var resetLogCmd = &cobra.Command{
	Use:   "logreset",
	Short: "Reset device log",
	Long:  "Reset device log",
	Run:   resetLog,
}

var (
	logDone = make(chan struct{})

	logIndex uint16 = 0
	//triggered bool   = false // TODO: Dont use this variable
)

func dumpLogHandler(b *boards.Basic) error {
	var logCount = b.LogEntries()

	if logIndex >= logCount {
		close(logDone)
	}

	fmt.Println(logIndex)

	// TODO: Check index of last received log message before incrementing
	go b.GetLog(logIndex)
	logIndex++

	return nil
}

func dumpLog(cmd *cobra.Command, args []string) {
	m := boards.Basic{}

	err := m.Init(deviceID, debug)
	if err != nil {
		log.Panicln(err)
		return
	}

	for !m.IsConnected() {
	}

	m.SetLogCallback(dumpLogHandler)

	// TODO: This won't be correct unless this receives status messages
	// Maybe make these log messages generic and then let each specific board wrap the generics
	if m.LogEntries() > 0 {
		m.GetLog(logIndex)
		logIndex++
	}

	<-logDone
	log.Println("Done")
}

func resetLog(cmd *cobra.Command, args []string) {
	m := boards.Basic{}

	err := m.Init(deviceID, debug)
	if err != nil {
		log.Panicln(err)
		return
	}

	for !m.IsConnected() {
	}

	err = m.ResetLog()
	if err != nil {
		return
	}

	time.Sleep(5 * time.Second)

	log.Println("Done")
}
