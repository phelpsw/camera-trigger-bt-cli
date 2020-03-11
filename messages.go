package main

import (
	"bytes"
	"encoding/binary"
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
	motionSensorMotion        uint8 = 2
	motionSensorStatus        uint8 = 3
	motionSensorAlert         uint8 = 4
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
	Temperature      float32
	Threshold        uint16
	Cooldown         float32
	MotionSensorType uint8
	LedOnMotion      uint8
	LedOnTransmit    uint8
	LedOnReceive     uint8
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

func MarshalBinary(msg interface{}) (data []byte, err error) {
	var buf bytes.Buffer
	err = binary.Write(&buf, binary.BigEndian, msg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func ReadMessage(b []byte) {
	if len(b) < 2 {
		log.Println("Cannot parse message less than 2 bytes")
	}
	header := BasicMessage{}
	buf := bytes.NewBuffer(b)
	err := binary.Read(buf, binary.BigEndian, &header)
	if err != nil {
		log.Fatal("binary.Read failed", err)
	}

	switch header.Type {
	case motionSensorAlert:
		msg := MotionSensorAlertMessage{}
		if int(header.Length) != binary.Size(msg) {
			log.Fatalln("Unexpected MotionSensorAlertMessage Length")
		}

		log.Println("MotionSensorAlertMessage")
		err := binary.Read(buf, binary.BigEndian, &msg)
		if err != nil {
			log.Fatal("binary.Read failed", err)
		}
	case motionSensorStatus:
		msg := MotionSensorStatusMessage{}
		if int(header.Length) != binary.Size(msg) {
			log.Fatalln("Unexpected MotionSensorStatusMessage Length")
		}

		log.Println("MotionSensorStatusMessage")
		err := binary.Read(buf, binary.BigEndian, &msg)
		if err != nil {
			log.Fatal("binary.Read failed", err)
		}
	case lightStatus:
		msg := LightStatusMessage{}
		if int(header.Length) != binary.Size(msg) {
			log.Fatalln("Unexpected LightStatusMessage Length")
		}

		log.Println("LightStatusMessage")
		err := binary.Read(buf, binary.BigEndian, &msg)
		if err != nil {
			log.Fatal("binary.Read failed", err)
		}
	case cameraStatus:
		msg := CameraStatusMessage{}
		if int(header.Length) != binary.Size(msg) {
			log.Fatalln("Unexpected CameraStatusMessage Length")
		}

		log.Println("CameraStatusMessage")
		err := binary.Read(buf, binary.BigEndian, &msg)
		if err != nil {
			log.Fatal("binary.Read failed", err)
		}
	}

	buf.Reset()
}

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
