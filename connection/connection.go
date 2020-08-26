package connection

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/backwardn/gatt"
)

var uartServiceID = gatt.MustParseUUID("49535343-fe7d-4ae5-8fa9-9fafd205e455")
var uartServiceRXCharID = gatt.MustParseUUID("49535343-8841-43f4-a8d4-ecbe34729bb3")
var uartServiceTXCharID = gatt.MustParseUUID("49535343-1e4d-4bd9-ba61-23c647249616")
var device gatt.Device
var devicePeripheral gatt.Peripheral
var receiveCharacteristic *gatt.Characteristic
var remoteName string
var debug bool

type readBytesCallbackType func(b []byte) error

var callback readBytesCallbackType
var connected bool

// Init a connection to the a bluetooth device with the specified name.
func Init(_device string, _callback readBytesCallbackType, _debug bool) error {
	remoteName = _device
	debug = _debug
	callback = _callback
	connected = false

	d, err := gatt.NewDevice(DefaultClientOptions...)

	if err != nil {
		return fmt.Errorf("failed to open device, err: %s", err)
	}

	d.Handle(
		gatt.PeripheralDiscovered(onPeriphDiscovered),
		gatt.PeripheralConnected(onPeriphConnected),
		gatt.PeripheralDisconnected(onPeriphDisconnected),
	)

	d.Init(onStateChanged)
	device = d

	return nil
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

func WriteBytes(b *bytes.Buffer) error {
	if debug {
		fmt.Printf("TX %d bytes: ", b.Len())

		for i := 0; i < b.Len(); i++ {
			fmt.Printf("0x%.2x, ", b.Bytes()[i])
		}
		fmt.Printf("\n")
	}

	err := devicePeripheral.WriteCharacteristic(receiveCharacteristic,
		b.Bytes(),
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
			cs, _ := p.DiscoverCharacteristics(nil, service)
			for _, c := range cs {
				if c.UUID().Equal(uartServiceTXCharID) {
					p.DiscoverDescriptors(nil, c)
					p.SetNotifyValue(c, readBytes)
					rxCfg = true
				} else if c.UUID().Equal(uartServiceRXCharID) {
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

func readBytes(c *gatt.Characteristic, b []byte, e error) {

	if e != nil {
		log.Printf("Connection packet handling error\n")
		log.Printf("%+v", e)
		return
	}

	if debug {
		fmt.Printf("RX %d bytes: ", len(b))
		for i := 0; i < len(b); i++ {
			fmt.Printf("0x%.2x, ", b[i])
		}
		fmt.Printf("\n")
	}

	if callback != nil {
		err := callback(b)
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
