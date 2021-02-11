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
	getFloatRequest           uint8 = 0x40
	getFloatResponse          uint8 = 0x41
	setFloatRequest           uint8 = 0x42
	setFloatResponse          uint8 = 0x43
	getUint16Request          uint8 = 0x44
	getUint16Response         uint8 = 0x45
	setUint16Request          uint8 = 0x46
	setUint16Response         uint8 = 0x47
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
	MotionThreshold  float32
	LuxLowThreshold  float32
	LuxHighThreshold float32
	Cooldown         float32
}

// NewMotionSensorConfigMessage generates a message of this type
func NewMotionSensorConfigMessage(
	motionThresh float32,
	luxLowThresh float32,
	luxHighThresh float32,
	cooldown float32) Message {
	return MotionSensorConfigMessage{
		Type:             motionSensorConfiguration,
		Length:           uint8(binary.Size(MotionSensorConfigMessage{})),
		MotionThreshold:  motionThresh,
		LuxLowThreshold:  luxLowThresh,
		LuxHighThreshold: luxHighThresh,
		Cooldown:         cooldown,
	}
}

type MotionSensorTriggerMessage struct {
	Type      uint8
	Length    uint8
	Timestamp Calendar
	Lux       float32
}

// NewMotionSensorTriggerMessage generates a message of this type
func NewMotionSensorTriggerMessage(
	lux float32) Message {
	return MotionSensorTriggerMessage{
		Type:      motionSensorTrigger,
		Length:    uint8(binary.Size(MotionSensorTriggerMessage{})),
		Timestamp: Calendar{0, 0, 0, 0, 0, 0, 0},
		Lux:       lux,
	}
}

type MotionSensorStatusMessage struct {
	Type             uint8
	Length           uint8
	Timestamp        Calendar
	Temperature      float32
	Voltage          float32
	Motion           float32
	MotionThreshold  float32
	Lux              float32
	LuxLowThreshold  float32
	LuxHighThreshold float32
	Cooldown         float32
	MotionSensorType uint8
	LedModes         uint8
	LogEntries       uint16
}

type LightStatus struct {
	Temperature      float32
	Voltage          float32
	Level            float32
	Delay            float32
	Attack           float32
	Sustain          float32
	Release          float32
	LightTemperature float32
	Current          float32
	LedModes         uint8
	LogEntries       uint16
}

type LightStatusMessage struct {
	BasicMessage

	Timestamp Calendar
	Payload   LightStatus
}

type GetUint16Request struct {
	BasicMessage

	Id      uint16
	Persist uint8
}

func NewGetUint16Request(
	id uint16, persist uint8) Message {
	header := BasicMessage{Type: getUint16Request,
		Length: uint8(binary.Size(GetUint16Request{}))}
	return GetUint16Request{
		BasicMessage: header,
		Id:           id,
		Persist:      persist,
	}
}

type GetUint16Response struct {
	BasicMessage

	Success uint8
	Id      uint16
	Persist uint8
	Value   uint16
}

type SetUint16Request struct {
	BasicMessage

	Id      uint16
	Persist uint8
	Value   uint16
}

func NewSetUint16Request(
	id uint16, persist uint8, value uint16) Message {
	header := BasicMessage{Type: setUint16Request,
		Length: uint8(binary.Size(SetUint16Request{}))}
	return SetUint16Request{
		BasicMessage: header,
		Id:           id,
		Persist:      persist,
		Value:        value,
	}
}

type SetUint16Response struct {
	BasicMessage

	Success uint8
	Id      uint16
	Persist uint8
	Value   uint16
}

type GetFloatRequest struct {
	BasicMessage

	Id      uint16
	Persist uint8
}

func NewGetFloatRequest(
	id uint16, persist uint8) Message {
	header := BasicMessage{Type: getFloatRequest,
		Length: uint8(binary.Size(GetFloatRequest{}))}
	return GetFloatRequest{
		BasicMessage: header,
		Id:           id,
		Persist:      persist,
	}
}

type GetFloatResponse struct {
	BasicMessage

	Success uint8
	Id      uint16
	Persist uint8
	Value   float32
}

type SetFloatRequest struct {
	BasicMessage

	Id      uint16
	Persist uint8
	Value   float32
}

func NewSetFloatRequest(
	id uint16, persist uint8, value float32) Message {
	header := BasicMessage{Type: setFloatRequest,
		Length: uint8(binary.Size(SetFloatRequest{}))}
	return SetFloatRequest{
		BasicMessage: header,
		Id:           id,
		Persist:      persist,
		Value:        value,
	}
}

type SetFloatResponse struct {
	BasicMessage

	Success uint8
	Id      uint16
	Persist uint8
	Value   float32
}

var rxBufArray []byte = make([]byte, 0, 512)
var rxBuf = bytes.NewBuffer(rxBufArray)

// ReadMessage parses a slice of bytes and returns a message if found
func ReadMessage(b []byte) (interface{}, error) {
	_, err := rxBuf.Write(b)
	if err != nil {
		return nil, err
	}

	// Buffer what we have and wait for more data
	if rxBuf.Len() < 2 {
		return nil, nil
	}

	for rxBuf.Len() > 2 {
		if int(rxBuf.Bytes()[1]) != getMessageTypeLength(rxBuf.Bytes()[0]) {
			_, _ = rxBuf.ReadByte()

			// Not strictly an error
			if rxBuf.Len() <= 2 {
				return nil, nil
			}
		} else {
			// A message has potentially been identified
			break
		}
	}

	// TODO: Maybe add a checksum and validate that here

	header := BasicMessage{rxBuf.Bytes()[0], rxBuf.Bytes()[1]}
	//log.Printf("Message Type 0x%x Length %d %d\n", header.Type, header.Length, rxBuf.Len())

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
	case getUint16Response:
		msg := GetUint16Response{}
		err = binary.Read(rxBuf, binary.BigEndian, &msg)
		if err != nil {
			return nil, err
		}
		return msg, nil
	case setUint16Response:
		msg := SetUint16Response{}
		err = binary.Read(rxBuf, binary.BigEndian, &msg)
		if err != nil {
			return nil, err
		}
		return msg, nil
	case getFloatResponse:
		msg := GetFloatResponse{}
		err = binary.Read(rxBuf, binary.BigEndian, &msg)
		if err != nil {
			return nil, err
		}
		return msg, nil
	case setFloatResponse:
		msg := SetFloatResponse{}
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
	case getUint16Response:
		return binary.Size(GetUint16Response{})
	case setUint16Response:
		return binary.Size(SetUint16Response{})
	case getFloatResponse:
		return binary.Size(GetFloatResponse{})
	case setFloatResponse:
		return binary.Size(SetFloatResponse{})
	default:
		return -1
	}
}
