package messages

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

/*
 * Motion Sensor Types
 */
const (
	motionSensor10m  uint8 = 1
	motionSensorSpot uint8 = 2
)

/*
 * LED Configuration States
 */
const (
	ledOff   uint8 = 0
	ledRed   uint8 = 1
	ledGreen uint8 = 2
)

type Message interface{}

/*
 * Bluetooth Message Types
 */
const (
	logRequest                uint8 = 0x01
	logResponse               uint8 = 0x02
	logReset                  uint8 = 0x03
	setTime                   uint8 = 0x04
	motionSensorConfiguration uint8 = 0x10
	motionSensorStatus        uint8 = 0x11
	motionSensorTrigger       uint8 = 0x12
	lightConfiguration        uint8 = 0x20
	lightStatus               uint8 = 0x21
	cameraConfiguration       uint8 = 0x30
	cameraStatus              uint8 = 0x31
)

type Calendar struct {
	Seconds    uint8
	Minutes    uint8
	Hours      uint8
	DayOfWeek  uint8
	DayOfMonth uint8
	Month      uint8
	Year       uint16
}

type BasicMessage struct {
	Type   uint8
	Length uint8
}

type LogRequestMessage struct {
	Type   uint8
	Length uint8
	Index  uint16
}

func NewLogRequestMessage(
	index uint16) Message {
	return LogRequestMessage{
		Type:   logRequest,
		Length: uint8(binary.Size(LogRequestMessage{})),
		Index:  index,
	}
}

type LogResponseMessage struct {
	Type      uint8
	Length    uint8
	Index     uint16
	Timestamp Calendar
	LogType   uint8
	Payload   [13]byte
}

type LogResetMessage struct {
	Type   uint8
	Length uint8
}

func NewLogResetMessage() Message {
	return LogResetMessage{
		Type:   logReset,
		Length: uint8(binary.Size(LogResetMessage{})),
	}
}

type SetTimeMessage struct {
	Type      uint8
	Length    uint8
	Timestamp Calendar
}

func NewSetTimeMessage(
	timestamp Calendar) Message {
	return SetTimeMessage{
		Type:      setTime,
		Length:    uint8(binary.Size(SetTimeMessage{})),
		Timestamp: timestamp,
	}
}

type LightConfigMessage struct {
	Type    uint8
	Length  uint8
	Level   float32
	Delay   float32
	Attack  float32
	Sustain float32
	Release float32
}

func NewLightConfigMessage(
	level float32,
	delay float32,
	attack float32,
	sustain float32,
	release float32) Message {
	return LightConfigMessage{
		Type:    lightConfiguration,
		Length:  uint8(binary.Size(LightConfigMessage{})),
		Level:   level,
		Delay:   delay,
		Attack:  attack,
		Sustain: sustain,
		Release: release,
	}
}

type MotionSensorConfigMessage struct {
	Type             uint8
	Length           uint8
	MotionThreshold  uint16
	LuxLowThreshold  float32
	LuxHighThreshold float32
}

// NewMotionSensorConfigMessage generates a message of this type
func NewMotionSensorConfigMessage(
	motionThresh uint16,
	luxLowThresh float32,
	luxHighThresh float32) Message {
	return MotionSensorConfigMessage{
		Type:             motionSensorConfiguration,
		Length:           uint8(binary.Size(MotionSensorConfigMessage{})),
		MotionThreshold:  motionThresh,
		LuxLowThreshold:  luxLowThresh,
		LuxHighThreshold: luxHighThresh,
	}
}

type MotionSensorTriggerMessage struct {
	Type      uint8
	Length    uint8
	Timestamp Calendar
	Motion    uint16
	Lux       float32
}

// NewMotionSensorTriggerMessage generates a message of this type
func NewMotionSensorTriggerMessage(
	motion uint16,
	lux float32) Message {
	return MotionSensorTriggerMessage{
		Type:      motionSensorTrigger,
		Length:    uint8(binary.Size(MotionSensorTriggerMessage{})),
		Timestamp: Calendar{0, 0, 0, 0, 0, 0, 0},
		Motion:    motion,
		Lux:       lux,
	}
}

type MotionSensorStatusMessage struct {
	Type             uint8
	Length           uint8
	Timestamp        Calendar
	Temperature      float32
	Voltage          float32
	Motion           uint16
	MotionThreshold  uint16
	Lux              float32
	LuxLowThreshold  float32
	LuxHighThreshold float32
	Cooldown         float32
	MotionSensorType uint8
	LedModes         uint8
	LogEntries       uint16
}

type LightStatus struct {
	Temperature float32
	Voltage     float32
	Level       float32
	Delay       float32
	Attack      float32
	Sustain     float32
	Release     float32
	//Current     float32
	//Voltage     float32
	LedModes   uint8
	LogEntries uint16
}

type LightStatusMessage struct {
	BasicMessage

	Timestamp Calendar
	Payload   LightStatus
}

var rxBufArray []byte = make([]byte, 0, 512)
var rxBuf = bytes.NewBuffer(rxBufArray)

// ReadMessage parses a slice of bytes and returns a message if found
func ReadMessage(b []byte) (interface{}, error) {
	_, err := rxBuf.Write(b)
	if err != nil {
		return nil, err
	}

	//log.Printf("rxBuf.len %d\n", rxBuf.Len())

	// Buffer what we have and wait for more data
	if rxBuf.Len() < 2 {
		return nil, nil
	}

	header := BasicMessage{rxBuf.Bytes()[0], rxBuf.Bytes()[1]}
	//log.Printf("Message Type 0x%x Length %d %d\n", header.Type, header.Length, rxBuf.Len())

	// Validate message length
	if int(header.Length) != getMessageTypeLength(header.Type) {
		rxBuf.Reset()
		return nil, fmt.Errorf("Type %d Unexpected Length %d", header.Type, header.Length)
	}

	// Ensure entire message is in buffer, if not just wait for more
	if rxBuf.Len() < int(header.Length) {
		return nil, nil
	}

	switch header.Type {
	case logResponse:
		msg := LogResponseMessage{}
		err = binary.Read(rxBuf, binary.BigEndian, &msg)
		if err != nil {
			return nil, err
		}

		return msg, nil
	case motionSensorStatus:
		msg := MotionSensorStatusMessage{}
		err = binary.Read(rxBuf, binary.BigEndian, &msg)
		if err != nil {
			return nil, err
		}

		return msg, nil
	case lightStatus:
		msg := LightStatusMessage{}
		err = binary.Read(rxBuf, binary.BigEndian, &msg)
		if err != nil {
			return nil, err
		}
		return msg, nil
	default:
		rxBuf.Reset()
		return nil, fmt.Errorf("Unknown message type 0x%x, flushing buffer", header.Type)
	}
}

// WriteMessage serializes the message to a Buffer
func WriteMessage(msg interface{}) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)

	err := binary.Write(buf, binary.BigEndian, msg)
	if err != nil {
		return buf, err
	}
	return buf, nil
}

func getMessageTypeLength(_type uint8) int {
	switch _type {
	case logResponse:
		return binary.Size(LogResponseMessage{})
	case motionSensorStatus:
		return binary.Size(MotionSensorStatusMessage{})
	case lightStatus:
		return binary.Size(LightStatusMessage{})
	default:
		return -1
	}
}
