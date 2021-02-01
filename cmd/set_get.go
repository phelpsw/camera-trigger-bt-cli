package cmd

import (
	"fmt"
	"log"
	"strconv"

	"github.com/phelpsw/camera-trigger-bt-cli/boards"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(getUint16Cmd)
}

var getUint16Cmd = &cobra.Command{
	Use:   "gi [id] [persist]",
	Short: "Get Uint16",
	Long:  "Get Uint16",
	Args:  cobra.MinimumNArgs(2),
	Run:   getUint16,
}

func getUint16(cmd *cobra.Command, args []string) {
	m := boards.Basic{}

	id_tmp, _ := strconv.ParseUint(args[0], 10, 16)
	persist_tmp, _ := strconv.ParseBool(args[1])

	id := uint16(id_tmp)
	var persist uint8 = 0
	if persist_tmp {
		persist = 1
	}

	err := m.Init(deviceID, debug)
	if err != nil {
		log.Panicln(err)
		return
	}

	for !m.IsConnected() {
	}

	resp, err := m.GetUint16(id, persist)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("%+v\n", resp)
}
