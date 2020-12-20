package boards

import (
	"fmt"

	"github.com/phelpsw/camera-trigger-bt-cli/connection"
	"github.com/phelpsw/camera-trigger-bt-cli/messages"
)

type Camera struct {
	name     string
	conn     *connection.Connection
	last     messages.CameraStatus
	desired  messages.CameraStatus
	callback func(interface{}) error
}

var (
	durationPending bool = false
)

func (m *Camera) handleBytes(b []byte) error {
	msg, err := messages.ReadMessage(b)
	if err != nil {
		return err
	}

	// A full message was not found
	if msg == nil {
		return nil
	}

	switch msg.(type) {
	case messages.CameraStatusMessage:
		m.last = msg.(messages.CameraStatusMessage).Payload
	default:
		fmt.Println("Unknown")
		return fmt.Errorf("unexpected message type %+v", msg)
	}

	if durationPending && floatEquals(m.last.Duration, m.desired.Duration) {
		durationPending = false
	}

	if m.callback != nil {
		err = m.callback(m)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Camera) Init(name string, debug bool) error {
	m.name = name

	m.conn = &connection.Connection{}
	err := m.conn.Init(name, m.handleBytes, debug)
	if err != nil {
		return err
	}

	return nil
}

func (m *Camera) InitFromBasic(b *Basic) error {
	m.conn = b.GetConnection()
	m.conn.Callback(m.handleBytes)

	return nil
}

func (m *Camera) SetUpdateCallback(callback func(interface{}) error) {
	m.callback = callback
}

func (m *Camera) Temperature() float32 {
	return m.last.Temperature
}

func (m *Camera) Voltage() float32 {
	return m.last.Voltage
}

func (m *Camera) Duration() float32 {
	return m.last.Duration
}

func (m *Camera) LogEntries() uint16 {
	return m.last.LogEntries
}

func (m *Camera) SetDuration(val float32, sync bool) error {
	m.desired.Duration = val
	durationPending = true

	if sync {
		m.Sync()
	}

	return nil
}

func (m *Camera) IsSynced() bool {
	if !durationPending {
		return true
	}
	return false
}

func (m *Camera) Sync() error {
	if m.IsSynced() {
		return nil
	}

	if !durationPending {
		m.desired.Duration = m.last.Duration
	}

	msg := messages.NewCameraConfigMessage(m.desired.Duration)

	if !m.conn.IsConnected() {
		return fmt.Errorf("not connected")
	}

	buf, err := messages.WriteMessage(msg)
	if err != nil {
		return err
	}

	err = m.conn.WriteBytes(buf)
	if err != nil {
		return err
	}

	return nil
}

func (m *Camera) Trigger(lux float32) error {
	if !m.IsSynced() {
		return fmt.Errorf("not synced")
	}

	msg := messages.NewMotionSensorTriggerMessage(lux)

	if !m.conn.IsConnected() {
		return fmt.Errorf("not connected")
	}

	buf, err := messages.WriteMessage(msg)
	if err != nil {
		return err
	}

	err = m.conn.WriteBytes(buf)
	if err != nil {
		return err
	}

	return nil
}
