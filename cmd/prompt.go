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
	{"device_id", "Device ID"},
	{"device_group", "Device Group ID"},
	{"sony_sleep_mode", "Sony camera sleep mode, 0 - Off, 1 - Idle"},
	{"led_on_record", "1 - red led, 2 - green led, 0 - disabled"},
	{"motion_blink_on_detect", "1 - red led, 2 - green led, 0 - disabled"},
	{"motion_transmit_on_detect", "1 - transmit enable, 0 - disabled"},
}

var uint16_temp = []Member{
	{"version_major", "Major version number"},
	{"version_minor", "Minor version number"},
	{"version_patch", "Patch version number"},
	{"version_dirty", "Bit indicating whether local mods have been made"},
	{"version_hash1", "Top 16 bits of git version hash"},
	{"version_hash2", "Lower 16 bits of git version hash"},
	{"part_number", ""},
	{"serial_number", ""},
	{"manufacture_year", ""},
	{"manufacture_doy", ""},
	{"device_type", "Device type"},
	{"led_red_state", "Red LED State, 0 - Off, 1 - On, 2 - Blink Once, 3 - Blink Continuous"},
	{"led_green_state", "Green LED State, 0 - Off, 1 - On, 2 - Blink Once, 3 - Blink Continuous"},
	{"motion_state", "Motion sensor state machine state"},
	{"motion_trigger_count", "Motion sensor trigger count since boot"},
	{"trigger_state", "State of device trigger, 0 available, 1 - cooldown"},
	{"runcam_control_state", "Runcam controller state"},
	{"runcam_state", "Runcam button push state machine"},
	{"sony_control_state", ""},
	{"sony_version_major", ""},
	{"sony_version_minor", ""},
	{"sony_version_patch", ""},
	{"sony_version_dirty", ""},
	{"sony_version_reg1", ""},
	{"sony_version_reg2", ""},
	{"sony_version_reg3", ""},
	{"sony_version_reg4", ""},
	{"sony_type", ""},
	{"sony_mode", ""},
	{"sony_status", ""},
	{"sony_led", ""},
}

var float_persist = []Member{
	{"motion_gain", "Motion sensor gain (0.0 - 1.0)"},
	{"motion_threshold", "Motion sensor trigger threshold (0.0 - 1.0)"},
	{"motion_cooldown", "Minimum seconds between motion sensor retrigger"},
	{"lux_interval", "Lux measurement interval"},
	{"video_duration", "Length of video recording trigger event in seconds"},
	{"trigger_max_duration", "Cumulative consecutive length of trigger events in seconds"},
	{"trigger_max_duration_cooldown", "Cooldown period following trigger max duration"},
	{"led_on_period", "On time of LED blink"},
	{"led_off_period", "Seconds between continuous LED blinks"},
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
	{"motion_value", "Motion sensor value"},
	{"lux_value", "Lux measurement"},
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
		if err != nil {
			fmt.Println("cannot convert value to uint16")
			return
		}
		value_uint16 = uint16(tmp)
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
		if err != nil {
			fmt.Println("cannot convert value to float")
			return
		}
		value_float = float32(tmp)
	case "t":
		command = "t"
		if len(blocks) == 1 {
			value_float = 0.0
		} else if len(blocks) == 2 {
			tmp, err := strconv.ParseFloat(blocks[1], 32)
			if err != nil {
				fmt.Println("cannot convert value to float")
				return
			}
			value_float = float32(tmp)
		} else if len(blocks) > 2 {
			fmt.Println("syntax error: t <lux>")
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
	case "t":
		err := m.Trigger(value_float)
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Printf("trigger sent (lux %f)\n", value_float)
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
		{Text: "t", Description: "Trigger"},

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
