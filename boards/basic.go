package boards

import (
	"fmt"
	"math"
	"reflect"
	"time"

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
	name              string
	conn              *connection.Connection
	observedType      interface{}
	logCount          uint16
	statusCallback    func(interface{}) error
	getUint16Callback func(interface{}) error
	setUint16Callback func(interface{}) error
	getFloatCallback  func(interface{}) error
	setFloatCallback  func(interface{}) error
	logCallback       func(*Basic) error

	logMessages []messages.LogResponseMessage
}

var eps float32 = 0.000001

func floatEquals(a, b float32) bool {
	if float32(math.Abs(float64(a)-float64(b))) < eps {
		return true
	}
	return false
}

func (m *Basic) Scan() (map[string]connection.Device, error) {
	if m.conn == nil {
		m.conn = &connection.Connection{}
	}
	_, err := m.conn.Scan(1 * time.Second)
	time.Sleep(5 * time.Second)
	m.conn.StopScan()
	return m.conn.ListDevices(), err
}

func (m *Basic) Init(name string, debug bool) error {
	m.name = name
	m.observedType = nil

	if m.conn == nil {
		m.conn = &connection.Connection{}
	}
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
	case messages.LogResponseMessage:
		fmt.Printf("%+v\n", msg.(messages.LogResponseMessage))
		m.logMessages = append(m.logMessages, msg.(messages.LogResponseMessage))

		if m.logCallback != nil {
			err = m.logCallback(m)
			if err != nil {
				return err
			}
		}
	case messages.GetUint16Response:
		if m.getUint16Callback != nil {
			err = m.getUint16Callback(msg)
			if err != nil {
				return err
			}
		}
	case messages.SetUint16Response:
		if m.setUint16Callback != nil {
			err = m.setUint16Callback(msg)
			if err != nil {
				return err
			}
		}
	case messages.GetFloatResponse:
		if m.getFloatCallback != nil {
			err = m.getFloatCallback(msg)
			if err != nil {
				return err
			}
		}
	case messages.SetFloatResponse:
		if m.setFloatCallback != nil {
			err = m.setFloatCallback(msg)
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

func (m *Basic) GetUint16(id uint16, persist uint8) (messages.GetUint16Response, error) {
	var message messages.GetUint16Response
	msg := messages.NewGetUint16Request(id, persist)

	if !m.conn.IsConnected() {
		return message, fmt.Errorf("not connected")
	}

	buf, err := messages.WriteMessage(msg)
	if err != nil {
		return message, err
	}

	response := false
	m.getUint16Callback = func(b interface{}) error {
		message = b.(messages.GetUint16Response)
		response = true
		return nil
	}

	err = m.conn.WriteBytes(buf)
	if err != nil {
		m.getUint16Callback = nil
		return message, err
	}

	attempts := 0
	for !response {
		attempts++
		if attempts > 500 {
			m.getUint16Callback = nil
			return message, fmt.Errorf("GetUint16: timeout")
		}
		time.Sleep(10 * time.Millisecond)
	}

	m.getUint16Callback = nil
	return message, nil
}

func (m *Basic) SetUint16(id uint16, persist uint8, value uint16) (messages.SetUint16Response, error) {

	var message messages.SetUint16Response
	msg := messages.NewSetUint16Request(id, persist, value)

	if !m.conn.IsConnected() {
		return message, fmt.Errorf("not connected")
	}

	buf, err := messages.WriteMessage(msg)
	if err != nil {
		return message, err
	}

	response := false
	m.setUint16Callback = func(b interface{}) error {
		message = b.(messages.SetUint16Response)
		response = true
		return nil
	}

	err = m.conn.WriteBytes(buf)
	if err != nil {
		m.setUint16Callback = nil
		return message, err
	}

	attempts := 0
	for !response {
		attempts++
		if attempts > 500 {
			m.setUint16Callback = nil
			return message, fmt.Errorf("SetUint16: timeout")
		}
		time.Sleep(10 * time.Millisecond)
	}

	m.setUint16Callback = nil
	return message, nil
}

func (m *Basic) GetFloat(id uint16, persist uint8) (messages.GetFloatResponse, error) {

	var message messages.GetFloatResponse
	msg := messages.NewGetFloatRequest(id, persist)

	if !m.conn.IsConnected() {
		return message, fmt.Errorf("not connected")
	}

	buf, err := messages.WriteMessage(msg)
	if err != nil {
		return message, err
	}

	response := false
	m.getFloatCallback = func(b interface{}) error {
		message = b.(messages.GetFloatResponse)
		response = true
		return nil
	}

	err = m.conn.WriteBytes(buf)
	if err != nil {
		m.getFloatCallback = nil
		return message, err
	}

	attempts := 0
	for !response {
		attempts++
		if attempts > 500 {
			m.getFloatCallback = nil
			return message, fmt.Errorf("GetFloat: timeout")
		}
		time.Sleep(10 * time.Millisecond)
	}

	m.getFloatCallback = nil
	return message, nil
}

func (m *Basic) SetFloat(id uint16, persist uint8, value float32) (messages.SetFloatResponse, error) {

	var message messages.SetFloatResponse
	msg := messages.NewSetFloatRequest(id, persist, value)

	if !m.conn.IsConnected() {
		return message, fmt.Errorf("not connected")
	}

	buf, err := messages.WriteMessage(msg)
	if err != nil {
		return message, err
	}

	response := false
	m.setFloatCallback = func(b interface{}) error {
		message = b.(messages.SetFloatResponse)
		response = true
		return nil
	}

	err = m.conn.WriteBytes(buf)
	if err != nil {
		m.setFloatCallback = nil
		return message, err
	}

	attempts := 0
	for !response {
		attempts++
		if attempts > 500 {
			m.setFloatCallback = nil
			return message, fmt.Errorf("SetFloat: timeout")
		}
		time.Sleep(10 * time.Millisecond)
	}

	m.setFloatCallback = nil
	return message, nil
}

func (m *Basic) Trigger(lux float32) error {
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
