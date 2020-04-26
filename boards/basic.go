package boards

import (
	"fmt"

	"github.com/phelpsw/camera-trigger-bt-cli/connection"
	"github.com/phelpsw/camera-trigger-bt-cli/messages"
)

type Basic struct {
	name string
}

func (m *Basic) handleBytes(b []byte) error {
	msg, err := messages.ReadMessage(b)
	if err != nil {
		return err
	}

	// A full message was not found
	if msg == nil {
		return nil
	}

	switch msg.(type) {
	case messages.MotionSensorStatusMessage:
		fmt.Println("MotionSensorStatusMessage")
		fmt.Printf("%+v\n", msg.(messages.MotionSensorStatusMessage))
	case messages.LightStatusMessage:
		fmt.Printf("%+v\n", msg.(messages.LightStatusMessage))
	default:
		fmt.Println("Unknown")
		return fmt.Errorf("unexpected message type %+v", msg)
	}

	return nil
}

func (m *Basic) Init(name string, debug bool) error {
	m.name = name

	err := connection.Init(name, m.handleBytes, debug)
	if err != nil {
		return err
	}

	return nil
}

func (m *Basic) SetUpdateCallback(callback func(interface{}) error) {
}
