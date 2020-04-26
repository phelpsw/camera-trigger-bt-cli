package boards

import (
	"fmt"
	"math"

	"github.com/phelpsw/camera-trigger-bt-cli/connection"
	"github.com/phelpsw/camera-trigger-bt-cli/messages"
)

type Light struct {
	name     string
	last     messages.LightStatus
	desired  messages.LightStatus
	callback func(interface{}) error
}

var (
	levelPending   bool = false
	delayPending   bool = false
	attackPending  bool = false
	sustainPending bool = false
	releasePending bool = false
)

var eps float32 = 0.000001

func floatEquals(a, b float32) bool {
	if float32(math.Abs(float64(a)-float64(b))) < eps {
		return true
	}
	return false
}

func (m *Light) handleBytes(b []byte) error {
	msg, err := messages.ReadMessage(b)
	if err != nil {
		return err
	}

	// A full message was not found
	if msg == nil {
		return nil
	}

	switch msg.(type) {
	case messages.LightStatusMessage:
		m.last = msg.(messages.LightStatusMessage).Payload
	default:
		fmt.Println("Unknown")
		return fmt.Errorf("unexpected message type %+v", msg)
	}

	if levelPending && floatEquals(m.last.Level, m.desired.Level) {
		levelPending = false
	}
	if delayPending && floatEquals(m.last.Delay, m.desired.Delay) {
		delayPending = false
	}
	if attackPending && floatEquals(m.last.Attack, m.desired.Attack) {
		attackPending = false
	}
	if sustainPending && floatEquals(m.last.Sustain, m.desired.Sustain) {
		sustainPending = false
	}
	if releasePending && floatEquals(m.last.Release, m.desired.Release) {
		releasePending = false
	}

	if m.callback != nil {
		err = m.callback(m)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Light) Init(name string, debug bool) error {
	m.name = name

	err := connection.Init(name, m.handleBytes, debug)
	if err != nil {
		return err
	}

	return nil
}

func (m *Light) SetUpdateCallback(callback func(interface{}) error) {
	m.callback = callback
}

func (m *Light) Temperature() float32 {
	return m.last.Temperature
}

func (m *Light) Voltage() float32 {
	return m.last.Voltage
}

func (m *Light) Level() float32 {
	return m.last.Level
}

func (m *Light) Delay() float32 {
	return m.last.Delay
}

func (m *Light) Attack() float32 {
	return m.last.Attack
}

func (m *Light) Sustain() float32 {
	return m.last.Sustain
}

func (m *Light) Release() float32 {
	return m.last.Release
}

// TODO: Enumerate this properly
func (m *Light) LedModes() uint8 {
	return m.last.LedModes
}

func (m *Light) LogEntries() uint16 {
	return m.last.LogEntries
}

func (m *Light) SetLevel(val float32, sync bool) error {
	m.desired.Level = val
	levelPending = true

	if sync {
		m.Sync()
	}

	return nil
}

func (m *Light) SetDelay(val float32, sync bool) error {
	m.desired.Delay = val
	delayPending = true

	if sync {
		m.Sync()
	}

	return nil
}

func (m *Light) SetAttack(val float32, sync bool) error {
	m.desired.Attack = val
	attackPending = true

	if sync {
		m.Sync()
	}

	return nil
}

func (m *Light) SetSustain(val float32, sync bool) error {
	m.desired.Sustain = val
	sustainPending = true

	if sync {
		m.Sync()
	}

	return nil
}

func (m *Light) SetRelease(val float32, sync bool) error {
	m.desired.Release = val
	releasePending = true

	if sync {
		m.Sync()
	}

	return nil
}

func (m *Light) IsSynced() bool {
	if !levelPending &&
		!delayPending &&
		!attackPending &&
		!sustainPending &&
		!releasePending {
		return true
	}
	return false
}

func (m *Light) Sync() error {
	if m.IsSynced() {
		return nil
	}

	if !levelPending {
		m.desired.Level = m.last.Level
	}

	if !delayPending {
		m.desired.Delay = m.last.Delay
	}

	if !attackPending {
		m.desired.Attack = m.last.Attack
	}

	if !sustainPending {
		m.desired.Sustain = m.last.Sustain
	}

	if !releasePending {
		m.desired.Release = m.last.Release
	}

	msg := messages.NewLightConfigMessage(m.desired.Level,
		m.desired.Delay,
		m.desired.Attack,
		m.desired.Sustain,
		m.desired.Release)

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

func (m *Light) Trigger(motion uint16, lux float32) error {
	if !m.IsSynced() {
		return fmt.Errorf("not synced")
	}

	msg := messages.NewMotionSensorTriggerMessage(motion, lux)

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
