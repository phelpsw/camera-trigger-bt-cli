package boards

import (
	"fmt"

	"github.com/phelpsw/camera-trigger-bt-cli/connection"
	"github.com/phelpsw/camera-trigger-bt-cli/messages"
)

type Generic interface {
	Init(name string, debug bool) error

	IsConnected() bool

	LogEntries() uint16

	GetLog(index uint16) error
	ResetLog() error
	SetTime() error
}

type Basic struct {
	name           string
	logCount       uint16
	statusCallback func(interface{}) error
	logCallback    func(*Basic) error
	connected      bool

	logMessages []messages.LogResponseMessage
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
		fmt.Printf("%+v\n", msg.(messages.MotionSensorStatusMessage))
		val := msg.(messages.MotionSensorStatusMessage)
		m.logCount = val.LogEntries

		if m.statusCallback != nil {
			err = m.statusCallback(m)
			if err != nil {
				return err
			}
		}

		m.connected = true
	case messages.LightStatusMessage:
		fmt.Printf("%+v\n", msg.(messages.LightStatusMessage))
		val := msg.(messages.LightStatusMessage)
		m.logCount = val.Payload.LogEntries

		if m.statusCallback != nil {
			err = m.statusCallback(m)
			if err != nil {
				return err
			}
		}

		m.connected = true
	case messages.LogResponseMessage:
		fmt.Printf("%+v\n", msg.(messages.LogResponseMessage))
		m.logMessages = append(m.logMessages, msg.(messages.LogResponseMessage))

		if m.logCallback != nil {
			err = m.logCallback(m)
			if err != nil {
				return err
			}
		}
	default:
		fmt.Println("Unknown")
		return fmt.Errorf("unexpected message type %+v", msg)
	}

	if m.statusCallback != nil {
		err = m.statusCallback(m)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Basic) Init(name string, debug bool) error {
	m.connected = false
	m.name = name

	err := connection.Init(name, m.handleBytes, debug)
	if err != nil {
		return err
	}

	return nil
}

func (m *Basic) IsConnected() bool {
	return m.connected
}

func (m *Basic) LogEntries() uint16 {
	return m.logCount
}

func (m *Basic) Log() []messages.LogResponseMessage {
	return m.logMessages
}

func (m *Basic) GetLog(index uint16) error {
	msg := messages.NewLogRequestMessage(index)

	if !connection.IsConnected() {
		return fmt.Errorf("not connected")
	}

	buf, err := messages.WriteMessage(msg)
	if err != nil {
		return err
	}

	err = connection.WriteBytes(buf)
	if err != nil {
		return err
	}

	return nil
}

func (m *Basic) ResetLog() error {
	msg := messages.NewLogResetMessage()

	if !connection.IsConnected() {
		return fmt.Errorf("not connected")
	}

	buf, err := messages.WriteMessage(msg)
	if err != nil {
		return err
	}

	err = connection.WriteBytes(buf)
	if err != nil {
		return err
	}

	return nil
}

func (m *Basic) SetTime(cal messages.Calendar) error {
	msg := messages.NewSetTimeMessage(cal)

	if !connection.IsConnected() {
		return fmt.Errorf("not connected")
	}

	buf, err := messages.WriteMessage(msg)
	if err != nil {
		return err
	}

	err = connection.WriteBytes(buf)
	if err != nil {
		return err
	}

	return nil
}

func (m *Basic) SetUpdateCallback(callback func(interface{}) error) {
	m.statusCallback = callback
}

func (m *Basic) SetLogCallback(callback func(*Basic) error) {
	m.logCallback = callback
}
