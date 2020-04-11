package messages

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
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
	motionSensorConfiguration uint8 = 1
	motionSensorStatus        uint8 = 2
	lightConfiguration        uint8 = 10
	lightStatus               uint8 = 11
	lightTrigger              uint8 = 12
	cameraConfiguration       uint8 = 20
	cameraStatus              uint8 = 21
	cameraTrigger             uint8 = 22
)

type Calendar struct {
	Seconds uint8
	Minutes uint8
	Hours   uint8
	Month   uint8
	Year    uint16
}

type BasicMessage struct {
	Type   uint8
	Length uint8
}

type CameraConfigurationMessage struct {
	Type     uint8
	Length   uint8
	Exposure float32
}

type LightConfigurationMessage struct {
	Type    uint8
	Length  uint8
	Delay   float32
	Attack  float32
	Sustain float32
	Release float32
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
	Delay       float32
	Attack      float32
	Sustain     float32
	Release     float32
	Temperature float32
}

type LightStatusMessage struct {
	BasicMessage

	Timestamp Calendar
	Payload   LightStatus
}

/*
type CameraStatusMessage struct {
	Type        uint8
	Length      uint8
	Timestamp   Calendar
	Exposure    float32
	Temperature float32
}
*/

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

	switch msg.(type) {
	case MotionSensorConfigMessage:
		err := binary.Write(buf, binary.BigEndian, msg.(MotionSensorConfigMessage))
		if err != nil {
			return buf, err
		}
		return buf, nil
	default:
		return nil, fmt.Errorf("unknown type %+v", reflect.TypeOf(msg))
	}
}

func getMessageTypeLength(_type uint8) int {
	switch _type {
	case motionSensorStatus:
		return binary.Size(MotionSensorStatusMessage{})
	case lightStatus:
		return binary.Size(LightStatusMessage{})
	default:
		return -1
	}
}
