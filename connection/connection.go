package connection

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"strings"

	"github.com/backwardn/gatt"
	"github.com/phelpsw/camera-trigger-bt-cli/messages"
)

var uartServiceID = gatt.MustParseUUID("49535343-fe7d-4ae5-8fa9-9fafd205e455")
var uartServiceRXCharID = gatt.MustParseUUID("49535343-8841-43f4-a8d4-ecbe34729bb3")
var uartServiceTXCharID = gatt.MustParseUUID("49535343-1e4d-4bd9-ba61-23c647249616")
var device gatt.Device
var devicePeripheral gatt.Peripheral
var receiveCharacteristic *gatt.Characteristic
var remoteName string
var debug bool

type callbackType func(interface{}) error

var callback callbackType
var connected bool

// Init a connection to the a bluetooth device with the specified name.
func Init(_device string, _callback callbackType, _debug bool) {
	remoteName = _device
	debug = _debug
	callback = _callback
	connected = false

	var DefaultClientOptions = []gatt.Option{
		gatt.LnxMaxConnections(1),
		gatt.LnxDeviceID(-1, false),
	}

	d, err := gatt.NewDevice(DefaultClientOptions...)
	if err != nil {
		log.Fatalf("Failed to open device, err: %s\n", err)
		return
	}

	d.Handle(
		gatt.PeripheralDiscovered(onPeriphDiscovered),
		gatt.PeripheralConnected(onPeriphConnected),
		gatt.PeripheralDisconnected(onPeriphDisconnected),
	)

	d.Init(onStateChanged)
	device = d
}

// Stop the connection
func Stop() {
	if devicePeripheral != nil {
		device.CancelConnection(devicePeripheral)
	}
	//device.Stop() // Seems to hang on hci.Close()
}

// IsConnected indicates whether a connection is present
func IsConnected() bool {
	return connected
}

// WriteMessage writes to the connected bluetooth device
func WriteMessage(msg interface{}) error {
	if !connected {
		return fmt.Errorf("attempting write when not connected")
	}
	if receiveCharacteristic == nil {
		return fmt.Errorf("attempting write when characteristic unknown")
	}
	if devicePeripheral == nil {
		return fmt.Errorf("attempting write when device unknown")
	}

	var bufArray []byte = make([]byte, 0, 512)
	var buf = bytes.NewBuffer(bufArray)
	err := binary.Write(buf, binary.BigEndian, msg)
	if err != nil {
		return err
	}

	if debug {
		fmt.Printf("(%d bytes) ", len(bufArray))
		for i := 0; i < len(bufArray); i++ {
			fmt.Printf("0x%.2x, ", bufArray[i])
		}
		fmt.Printf("\n")
	}

	err = devicePeripheral.WriteCharacteristic(receiveCharacteristic,
		bufArray,
		true)
	if err != nil {
		return err
	}

	return nil
}

func onStateChanged(d gatt.Device, s gatt.State) {
	log.Printf("State: %v", s)
	switch s {
	case gatt.StatePoweredOn:
		log.Println("scanning...")
		d.Scan([]gatt.UUID{}, false)
		return
	default:
		d.StopScanning()
	}
}

func onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	if strings.HasPrefix(a.LocalName, "camera-trigger-") {
		log.Printf("Discovered %s RSSI %d\n", a.LocalName, rssi)

		if a.LocalName == remoteName {
			p.Device().StopScanning()
			p.Device().Connect(p)
		}
	}
	return
}

func onPeriphConnected(p gatt.Peripheral, err error) {
	services, err := p.DiscoverServices(nil)
	if err != nil {
		log.Fatalf("Failed to discover services, err: %s\n", err)
		return
	}

	rxCfg := false
	txCfg := false
	for _, service := range services {
		if service.UUID().Equal(uartServiceID) {
			//log.Printf("Expected service found: %s (%s)\n",
			//	service.Name(),
			//	service.UUID())

			cs, _ := p.DiscoverCharacteristics(nil, service)
			for _, c := range cs {
				//log.Printf("Characteristic %s\n", c.UUID())
				if c.UUID().Equal(uartServiceTXCharID) {
					//log.Println("TX Characteristic Found")
					p.DiscoverDescriptors(nil, c)
					p.SetNotifyValue(c, handlePacket)
					rxCfg = true
				} else if c.UUID().Equal(uartServiceRXCharID) {
					//log.Println("RX Characteristic Found")
					devicePeripheral = p
					receiveCharacteristic = c
					txCfg = true
				}
			}
		}
	}

	if rxCfg && txCfg {
		connected = true
	}

	return
}

func handlePacket(c *gatt.Characteristic, b []byte, e error) {
	if debug {
		fmt.Printf("(%d bytes) ", len(b))
		for i := 0; i < len(b); i++ {
			fmt.Printf("0x%.2x, ", b[i])
		}
		fmt.Printf("\n")
	}

	msg, err := messages.ReadMessage(b)
	if err != nil {
		log.Printf("Message handling error: %s\n", err)
	}

	// TODO: Check if the message is a type with a Calendar internally
	// If so, confirm the Calendar is within 10 minutes of the current
	// time.  If not, immediately calculate a time correction and
	// command.
	//
	// If time correction needs to occur, don't show this status message

	if callback != nil {
		err := callback(msg)
		if err != nil {
			log.Printf("Callback handling error: %s\n", err)
		}
	}

	return
}

func onPeriphDisconnected(p gatt.Peripheral, err error) {
	log.Println("Disconnected")
	connected = false
}
