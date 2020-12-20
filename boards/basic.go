package boards

import (
	"fmt"
	"math"
	"reflect"

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

	PrintStatus() error
}

type Basic struct {
	name           string
	conn           *connection.Connection
	observedType   interface{}
	logCount       uint16
	statusCallback func(interface{}) error
	logCallback    func(*Basic) error

	logMessages []messages.LogResponseMessage
}

var eps float32 = 0.000001

func floatEquals(a, b float32) bool {
	if float32(math.Abs(float64(a)-float64(b))) < eps {
		return true
	}
	return false
}

func (m *Basic) Init(name string, debug bool) error {
	m.name = name
	m.observedType = nil

	m.conn = &connection.Connection{}
	err := m.conn.Init(name, m.handleBytes, debug)
	if err != nil {
		return err
	}

	return nil
}

func (m *Basic) GetType() interface{} {
	return m.observedType
}

func (m *Basic) GetConnection() *connection.Connection {
	return m.conn
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
		m.observedType = reflect.TypeOf(Motion{})
	case messages.LightStatusMessage:
		m.observedType = reflect.TypeOf(Light{})
	case messages.CameraStatusMessage:
		m.observedType = reflect.TypeOf(Camera{})
	case messages.LogResponseMessage:
		fmt.Printf("%+v\n", msg.(messages.LogResponseMessage))
		m.logMessages = append(m.logMessages, msg.(messages.LogResponseMessage))

		if m.logCallback != nil {
			err = m.logCallback(m)
			if err != nil {
				return err
			}
		}
	}

	if m.statusCallback != nil {
		err = m.statusCallback(m)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Basic) IsConnected() bool {
	return m.conn.IsConnected()
}

func (m *Basic) LogEntries() uint16 {
	return m.logCount
}

func (m *Basic) Log() []messages.LogResponseMessage {
	return m.logMessages
}

func (m *Basic) GetLog(index uint16) error {
	msg := messages.NewLogRequestMessage(index)

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

func (m *Basic) ResetLog() error {
	msg := messages.NewLogResetMessage()

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

func (m *Basic) SetTime(cal messages.Calendar) error {
	msg := messages.NewSetTimeMessage(cal)

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

func (m *Basic) SetUpdateCallback(callback func(interface{}) error) {
	m.statusCallback = callback
}

func (m *Basic) SetLogCallback(callback func(*Basic) error) {
	m.logCallback = callback
}
