package messages

import (
	"reflect"
	"testing"
)

func TestReadMessage(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		args    args
		wantMsg interface{}
		wantErr bool
	}{
		// TODO: Add test cases.

		{"Basic Parse",
			args{[]byte{0x03, 0x24,
				0x0b, 0x00, 0x00, 0x01, 0x00, 0x12,
				0x42, 0x8e, 0x99, 0x9a,
				0x42, 0xc8, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
				0x08, 0x05,
				0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
				0x00,
				0x00,
				0x00, 0x0a,
				0x00, 0x00, 0x00, 0x00}},
			MotionSensorStatusMessage{
				Type:   3,
				Length: 36,
				Timestamp: Calendar{
					Seconds: 11,
					Minutes: 0,
					Hours:   0,
					Month:   1,
					Year:    18},
				Lux:              71.3,
				LuxThreshold:     100.0,
				Temperature:      0.0,
				Motion:           2053,
				MotionThreshold:  0,
				Cooldown:         0.0,
				MotionSensorType: 0,
				LedModes:         0,
				LogEntries:       10,
				BtSleepDelay:     0.0,
			},
			false},
		{"Partial Parse Step 1",
			args{[]byte{0x03, 0x24,
				0x0b, 0x00, 0x00, 0x01, 0x00, 0x12,
				0x42, 0x8e, 0x99, 0x9a,
				0x42, 0xc8, 0x00, 0x00}},
			nil,
			false},
		{"Partial Parse Step 2",
			args{[]byte{
				0x00, 0x00, 0x00, 0x00,
				0x08, 0x05,
				0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
				0x00,
				0x00,
				0x00, 0x0a,
				0x00, 0x00, 0x00, 0x00}},
			MotionSensorStatusMessage{
				Type:   3,
				Length: 36,
				Timestamp: Calendar{
					Seconds: 11,
					Minutes: 0,
					Hours:   0,
					Month:   1,
					Year:    18},
				Lux:              71.3,
				LuxThreshold:     100.0,
				Temperature:      0.0,
				Motion:           2053,
				MotionThreshold:  0,
				Cooldown:         0.0,
				MotionSensorType: 0,
				LedModes:         0,
				LogEntries:       10,
				BtSleepDelay:     0.0,
			},
			false},
		{"Unknown Type",
			args{[]byte{0xFF, 0x02}},
			nil,
			true},
		{"Stub with invalid type part 1",
			args{[]byte{0xFF}},
			nil,
			false},
		{"Stub with invalid type part 2",
			args{[]byte{0xFF}},
			nil,
			true},
		{"Stub with valid type but invalid length",
			args{[]byte{0x03, 0xff}},
			nil,
			true},
		{"Light Status Message",
			args{[]byte{
				0x0b, 0x1c,
				0x0b, 0x00, 0x00, 0x01, 0x00, 0x12,
				0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00}},
			LightStatusMessage{
				BasicMessage: BasicMessage{
					Type:   11,
					Length: 28},
				Timestamp: Calendar{
					Seconds: 11,
					Minutes: 0,
					Hours:   0,
					Month:   1,
					Year:    18},
				Payload: LightStatus{
					Delay:       0.0,
					Attack:      0.0,
					Sustain:     0.0,
					Release:     0.0,
					Temperature: 0.0}},
			false},
		{"Motion Sensor Motion",
			args{[]byte{
				0x02, 0x04,
				0x04, 0x00}},
			MotionSensorMotionMessage{
				Type:   2,
				Length: 4,
				Motion: 1024},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMsg, err := ReadMessage(tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotMsg, tt.wantMsg) {
				t.Errorf("ReadMessage() = %v, want %v", gotMsg, tt.wantMsg)
			}
		})
	}
}
