package cmd

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/phelpsw/camera-trigger-bt-cli/boards"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(promptCmd)
}

var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Command prompt",
	Long:  "Command prompt",
	Run:   promptFunc,
}

type Member struct {
	Name        string
	Description string
}

var uint16_persist = []Member{
	{"motion_thresh", "Motion sensor motion threshold value 0-65535"},
	{"sony_sleep_mode", "Sony camera sleep mode, 0 - Off, 1 - Idle"},
}

var uint16_temp = []Member{
	{"major_version", "Major version number"},
	{"minor_version", "Minor version number"},
	{"patch_version", "Patch version number"},
	{"device_type", "Device type"},
}

var float_persist = []Member{
	{"motion_cooldown", "Minimum seconds between motion sensor retrigger"},
	{"runcam_video_duration", "Length of runcam recording in seconds"},
	{"sony_video_duration", "Length of sony recording in seconds"},
	{"light_level2_thresh", "Lux level for light brightness level 2"},
	{"light_level3_thresh", "Lux level for light brightness level 3"},
	{"light_delay", "Light delay in seconds before fade up"},
	{"light_attack", "Light fade up in seconds"},
	{"light_sustain", "Light on period in seconds"},
	{"light_release", "Light fade out in seconds"},
}

var float_temp = []Member{
	{"cpu_temperature", "Major version number"},
	{"battery_voltage", "Minor version number"},
	{"uptime", "System uptime in seconds"},
}

var m boards.Basic

func getIndex(input string, members []Member) (int, error) {
	for indx, member := range members {
		if input == member.Name {
			return indx, nil
		}
	}

	return 0, fmt.Errorf("not found")
}

func getUint16Index(input string) (uint16, uint8, error) {
	indx, err := getIndex(input, uint16_persist)
	if err == nil {
		return uint16(indx), 1, nil
	}

	indx, err = getIndex(input, uint16_temp)
	if err == nil {
		return uint16(indx), 0, nil
	}

	return 0, 0, fmt.Errorf("not found")
}

func getFloatIndex(input string) (uint16, uint8, error) {
	indx, err := getIndex(input, float_persist)
	if err == nil {
		return uint16(indx), 1, nil
	}

	indx, err = getIndex(input, float_temp)
	if err == nil {
		return uint16(indx), 0, nil
	}

	return 0, 0, fmt.Errorf("not found")
}
func executorFunc(in string) {
	in = strings.TrimSpace(in)
	//fmt.Println("Your input: " + in)

	var value_uint16 uint16
	var value_float float32
	var command, variable string
	blocks := strings.Split(in, " ")
	switch blocks[0] {
	case "exit":
		fmt.Println("Bye!")
		os.Exit(0)
	case "gi":
		command = "gi"
		if len(blocks) != 2 {
			fmt.Println("variable required")
			return
		}
		variable = blocks[1]
	case "si":
		command = "si"
		if len(blocks) != 3 {
			fmt.Println("variable and value required")
			return
		}
		variable = blocks[1]
		tmp, err := strconv.ParseUint(blocks[2], 10, 16)
		value_uint16 = uint16(tmp)
		if err != nil {
			fmt.Println("cannot convert value to uint16")
			return
		}
	case "gf":
		command = "gf"
		if len(blocks) != 2 {
			fmt.Println("variable required")
			return
		}
		variable = blocks[1]
	case "sf":
		command = "sf"
		if len(blocks) != 3 {
			fmt.Println("variable and value required")
			return
		}
		variable = blocks[1]
		tmp, err := strconv.ParseFloat(blocks[2], 32)
		value_float = float32(tmp)
		if err != nil {
			fmt.Println("cannot convert value to float")
			return
		}
	}

	switch command {
	case "gi":
		indx, persist, err := getUint16Index(variable)
		if err != nil {
			fmt.Println("variable is not uint16")
			return
		}

		resp, err := m.GetUint16(indx, persist)
		if err != nil {
			log.Println(err)
			return
		}

		if resp.Success == 1 {
			fmt.Printf("%s: %d\n", variable, resp.Value)
		} else {
			fmt.Printf("%s: get failed\n", variable)
		}
	case "si":
		indx, persist, err := getUint16Index(variable)
		if err != nil {
			fmt.Println("variable is not uint16")
			return
		}

		resp, err := m.SetUint16(indx, persist, value_uint16)
		if err != nil {
			log.Println(err)
			return
		}

		if resp.Success == 1 {
			fmt.Printf("%s: %d\n", variable, resp.Value)
		} else {
			fmt.Printf("%s: set failed\n", variable)
		}
	case "gf":
		indx, persist, err := getFloatIndex(variable)
		if err != nil {
			fmt.Println("variable is not float")
			return
		}

		resp, err := m.GetFloat(indx, persist)
		if err != nil {
			log.Println(err)
			return
		}

		if resp.Success == 1 {
			fmt.Printf("%s: %f\n", variable, resp.Value)
		} else {
			fmt.Printf("%s: get failed\n", variable)
		}
	case "sf":
		indx, persist, err := getFloatIndex(variable)
		if err != nil {
			fmt.Println("variable is not float")
			return
		}

		resp, err := m.SetFloat(indx, persist, value_float)
		if err != nil {
			log.Println(err)
			return
		}

		if resp.Success == 1 {
			fmt.Printf("%s: %f\n", variable, resp.Value)
		} else {
			fmt.Printf("%s: set failed\n", variable)
		}
	}
}

func completerFunc(d prompt.Document) []prompt.Suggest {

	s := []prompt.Suggest{
		//{Text: "list", Description: "List camera-trigger devices"},
		//{Text: "connect", Description: "Connect to specified device"},
		//{Text: "disconnect", Description: "Disconnect from current device"},
		//{Text: "status", Description: "Display connectivity state"},
		{Text: "gi", Description: "Get uint16"},
		{Text: "si", Description: "Set uint16"},
		{Text: "gf", Description: "Get float"},
		{Text: "sf", Description: "Set float"},

		{Text: "exit", Description: "Exit the program"},
	}

	for _, elem := range uint16_persist {
		s = append(s, prompt.Suggest{Text: elem.Name, Description: elem.Description})
	}
	for _, elem := range uint16_temp {
		s = append(s, prompt.Suggest{Text: elem.Name, Description: elem.Description})
	}
	for _, elem := range float_persist {
		s = append(s, prompt.Suggest{Text: elem.Name, Description: elem.Description})
	}
	for _, elem := range float_temp {
		s = append(s, prompt.Suggest{Text: elem.Name, Description: elem.Description})
	}

	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func promptFunc(cmd *cobra.Command, args []string) {
	m = boards.Basic{}

	err := m.Init(deviceID, debug)
	if err != nil {
		log.Panicln(err)
		return
	}

	for !m.IsConnected() {
	}

	p := prompt.New(
		executorFunc,
		completerFunc,
		prompt.OptionTitle("camera-prompt: camera-trigger configuration prompt"),
		prompt.OptionPrefix("> "),
		prompt.OptionInputTextColor(prompt.Yellow),
		prompt.OptionCompletionOnDown(),
	)

	p.Run()
}