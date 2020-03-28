package messages

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
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

/*
 * Bluetooth Message Types
 */
const (
	motionSensorConfiguration uint8 = 1
	motionSensorMotion        uint8 = 2 // Include violation bit
	motionSensorStatus        uint8 = 3
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

type MotionSensorAlertMessage struct {
	Type   uint8
	Length uint8
}

type MotionSensorMotionMessage struct {
	Type   uint8
	Length uint8
	Motion uint16
}

type MotionSensorConfigurationMessage struct {
	Type          uint8
	Length        uint8
	Threshold     uint16
	Cooldown      float32
	LedOnMotion   uint8
	LedOnTransmit uint8
	LedOnReceive  uint8
	BtSleepDelay  float32
}

type MotionSensorStatusMessage struct {
	Type             uint8
	Length           uint8
	Timestamp        Calendar
	Lux              float32
	LuxThreshold     float32
	Temperature      float32
	Motion           uint16
	MotionThreshold  uint16
	Cooldown         float32
	MotionSensorType uint8
	LedModes         uint8
	LogEntries       uint16
	BtSleepDelay     float32
}

type LightStatusMessage struct {
	Type        uint8
	Length      uint8
	Timestamp   Calendar
	Delay       float32
	Attack      float32
	Sustain     float32
	Release     float32
	Temperature float32
}

type CameraStatusMessage struct {
	Type        uint8
	Length      uint8
	Timestamp   Calendar
	Exposure    float32
	Temperature float32
}

var rxBufArray []byte = make([]byte, 0, 512)
var rxBuf = bytes.NewBuffer(rxBufArray)

func MarshalBinary(msg interface{}) (data []byte, err error) {
	var buf bytes.Buffer
	err = binary.Write(&buf, binary.BigEndian, msg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func ReadMessage(b []byte) (interface{}, error) {
	_, err := rxBuf.Write(b)
	if err != nil {
		return nil, err
	}

	log.Printf("rxBuf.len %d\n", rxBuf.Len())

	// Buffer what we have and wait for more data
	if rxBuf.Len() < 2 {
		return nil, nil
	}

	header := BasicMessage{rxBuf.Bytes()[0], rxBuf.Bytes()[1]}
	log.Printf("Message Type 0x%x Length %d %d\n", header.Type, header.Length, rxBuf.Len())

	switch header.Type {
	case motionSensorMotion:
		msg := MotionSensorMotionMessage{}

		// Validate message length
		if int(header.Length) != binary.Size(msg) {
			rxBuf.Reset()
			return nil, fmt.Errorf("%T Unexpected Length %d", msg, header.Length)
		}

		// Ensure entire message is in buffer, if not just wait for more
		if rxBuf.Len() < int(header.Length) {
			return nil, nil
		}

		err = binary.Read(rxBuf, binary.BigEndian, &msg)
		if err != nil {
			return nil, err
		}

		return msg, nil
	case motionSensorStatus:
		msg := MotionSensorStatusMessage{}

		// Validate message length
		if int(header.Length) != binary.Size(msg) {
			rxBuf.Reset()
			return nil, fmt.Errorf("%T Unexpected Length %d", msg, header.Length)
		}

		// Ensure entire message is in buffer, if not just wait for more
		if rxBuf.Len() < int(header.Length) {
			return nil, nil
		}

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

/*
func getMessageType(_type uint8) interface{} {
	switch _type {
	case motionSensorMotion:
		return MotionSensorMotionMessage{}
	case motionSensorStatus:
		return MotionSensorStatusMessage{}
	default:
		return nil
	}
}
*/

/*
func SendMessage(conn io.Writer, m encoding.BinaryMarshaler) {
	if bs, err := m.MarshalBinary(); err != nil {
		log.Fatalln(err)
	}

	if _, err := conn.Write(bs); err != nil {
		log.Fatalln(err)
	}
}
*/
