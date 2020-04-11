package boards

import (
	"fmt"

	"github.com/phelpsw/camera-trigger-bt-cli/connection"
	"github.com/phelpsw/camera-trigger-bt-cli/messages"
)

type Motion struct {
	name     string
	last     messages.MotionSensorStatusMessage
	desired  messages.MotionSensorConfigMessage
	callback func(interface{}) error
}

var (
	threshPending  bool = false
	luxLowPending  bool = false
	luxHighPending bool = false
)

func (m *Motion) handleBytes(b []byte) error {
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
		m.last = msg.(messages.MotionSensorStatusMessage)
	default:
		fmt.Println("Unknown")
		return fmt.Errorf("unexpected message type %+v", msg)
	}

	if threshPending && m.last.MotionThreshold == m.desired.MotionThreshold {
		threshPending = false
	}
	if luxLowPending && m.last.LuxLowThreshold == m.desired.LuxLowThreshold {
		luxLowPending = false
	}
	if luxHighPending && m.last.LuxHighThreshold == m.desired.LuxHighThreshold {
		luxHighPending = false
	}

	if m.callback != nil {
		err = m.callback(m)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Motion) Init(name string, debug bool) error {
	m.name = name

	err := connection.Init(name, m.handleBytes, debug)
	if err != nil {
		return err
	}

	return nil
}

func (m *Motion) SetUpdateCallback(callback func(interface{}) error) {
	m.callback = callback
}

func (m *Motion) Temperature() float32 {
	return m.last.Temperature
}

func (m *Motion) Voltage() float32 {
	return m.last.Voltage
}

func (m *Motion) Motion() uint16 {
	return m.last.Motion
}

func (m *Motion) MotionThreshold() uint16 {
	return m.last.MotionThreshold
}

func (m *Motion) Lux() float32 {
	return m.last.Lux
}

func (m *Motion) LuxLowThreshold() float32 {
	return m.last.LuxLowThreshold
}

func (m *Motion) LuxHighThreshold() float32 {
	return m.last.LuxHighThreshold
}

func (m *Motion) Cooldown() float32 {
	return m.last.Cooldown
}

// TODO: Enumerate this properly
func (m *Motion) MotionSensorType() uint8 {
	return m.last.MotionSensorType
}

// TODO: Enumerate this properly
func (m *Motion) LedModes() uint8 {
	return m.last.LedModes
}

func (m *Motion) LogEntries() uint16 {
	return m.last.LogEntries
}

func (m *Motion) SetMotionThreshold(thresh uint16, sync bool) error {
	m.desired.MotionThreshold = thresh
	threshPending = true

	if sync {
		m.Sync()
	}

	return nil
}

func (m *Motion) SetLuxLowThreshold(thresh float32, sync bool) error {
	m.desired.LuxLowThreshold = thresh
	luxLowPending = true

	if sync {
		m.Sync()
	}

	return nil
}

func (m *Motion) SetLuxHighThreshold(thresh float32, sync bool) error {
	m.desired.LuxHighThreshold = thresh
	luxHighPending = true

	if sync {
		m.Sync()
	}

	return nil
}

func (m *Motion) IsSynced() bool {
	if !threshPending && !luxLowPending && !luxHighPending {
		return true
	}
	return false
}

func (m *Motion) Sync() error {
	if !threshPending && !luxLowPending && !luxHighPending {
		return nil
	}
	if !threshPending {
		m.desired.MotionThreshold = m.last.MotionThreshold
	}
	if !luxLowPending {
		m.desired.LuxLowThreshold = m.last.LuxLowThreshold
	}
	if !luxHighPending {
		m.desired.LuxHighThreshold = m.last.LuxHighThreshold
	}

	msg := messages.NewMotionSensorConfigMessage(m.desired.MotionThreshold,
		m.desired.LuxLowThreshold,
		m.desired.LuxHighThreshold)

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
