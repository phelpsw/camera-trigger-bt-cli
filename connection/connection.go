package connection

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/JuulLabs-OSS/ble"
	"github.com/pkg/errors"
)

var uartServiceID = ble.MustParse("49535343fe7d4ae58fa99fafd205e455")
var uartServiceRXCharID = ble.MustParse("49535343884143f4a8d4ecbe34729bb3")
var uartServiceTXCharID = ble.MustParse("495353431e4d4bd9ba6123c647249616")
var client ble.Client
var receiveCharacteristic *ble.Characteristic
var remoteName string
var debug bool

type readBytesCallbackType func(b []byte) error

var callback readBytesCallbackType
var connected bool

// NewDevice ...
func NewDevice(impl string, opts ...ble.Option) (d ble.Device, err error) {
	return DefaultDevice(opts...)
}

// Init a connection to the a bluetooth device with the specified name.
func Init(_device string, _callback readBytesCallbackType, _debug bool) error {
	remoteName = _device
	debug = _debug
	callback = _callback
	connected = false

	deviceName := "default"
	d, err := NewDevice(deviceName)
	if err != nil {
		return fmt.Errorf("failed to open device, err: %s", err)
	}
	ble.SetDefaultDevice(d)

	// Default to search device with name specified by user
	filter := func(a ble.Advertisement) bool {
		return strings.ToUpper(a.LocalName()) == strings.ToUpper(_device)
	}

	// Scan duration in seconds, 0 scans forever
	var scanDuration time.Duration = 10 * time.Second
	// Scan for specified durantion, or until interrupted by user.
	fmt.Printf("Scanning for %s sec...\n", scanDuration)
	ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), scanDuration))
	cln, err := ble.Connect(ctx, filter)
	if err != nil {
		log.Fatalf("can't connect : %s", err)
	}

	// Make sure we had the chance to print out the message.
	done := make(chan struct{})
	// Normally, the connection is disconnected by us after our exploration.
	// However, it can be asynchronously disconnected by the remote peripheral.
	// So we wait(detect) the disconnection in the go routine.
	go func() {
		<-cln.Disconnected()
		connected = false
		fmt.Printf("[ %s ] is disconnected \n", cln.Addr())
		close(done)
	}()

	fmt.Printf("Discovering profile...\n")
	p, err := cln.DiscoverProfile(true)
	if err != nil {
		log.Fatalf("can't discover profile: %s", err)
	}

	for _, s := range p.Services {
		fmt.Printf("    Service: %s %s, Handle (0x%02X)\n", s.UUID, ble.Name(s.UUID), s.Handle)
		for _, c := range s.Characteristics {
			fmt.Printf("      Characteristic: %s %s, Property: 0x%02X (%s), Handle(0x%02X), VHandle(0x%02X)\n",
				c.UUID, ble.Name(c.UUID), c.Property, propString(c.Property), c.Handle, c.ValueHandle)
		}

		if s.UUID.Equal(uartServiceID) { // Not sure if byte[] have .Equal operator
			for _, c := range s.Characteristics {
				// Validate this property supports notifies
				if (c.Property & ble.CharNotify) != 0 {
					if c.UUID.Equal(uartServiceTXCharID) {
						if err := cln.Subscribe(c, false, readBytes); err != nil {
							log.Fatalf("subscribe failed: %s", err)
						}
						//rxCfg = true
					}
				}

				if c.UUID.Equal(uartServiceRXCharID) {
					receiveCharacteristic = c
					//txCfg = true
				}
			}
		}
	}

	/*
		d.Handle(
			gatt.PeripheralDiscovered(onPeriphDiscovered),
			gatt.PeripheralConnected(onPeriphConnected),
			gatt.PeripheralDisconnected(onPeriphDisconnected),
		)

		d.Init(onStateChanged)
	*/
	client = cln
	connected = true

	return nil
}

// Stop the connection
func Stop() {
	if client != nil {
		client.CancelConnection()
	}
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

	err := client.WriteCharacteristic(receiveCharacteristic, b.Bytes(), true)
	return errors.Wrap(err, "can't write characteristic")
}

/*
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
*/
func readBytes(b []byte) {
	/*
		if e != nil {
			log.Printf("Connection packet handling error\n")
			log.Printf("%+v", e)
			return
		}
	*/
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

/*
func onPeriphDisconnected(p gatt.Peripheral, err error) {
	log.Println("Disconnected")
	connected = false
}
*/

func propString(p ble.Property) string {
	var s string
	for k, v := range map[ble.Property]string{
		ble.CharBroadcast:   "B",
		ble.CharRead:        "R",
		ble.CharWriteNR:     "w",
		ble.CharWrite:       "W",
		ble.CharNotify:      "N",
		ble.CharIndicate:    "I",
		ble.CharSignedWrite: "S",
		ble.CharExtended:    "E",
	} {
		if p&k != 0 {
			s += v
		}
	}
	return s
}
