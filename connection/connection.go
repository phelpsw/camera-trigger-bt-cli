package connection

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/JuulLabs-OSS/ble"
	"github.com/JuulLabs-OSS/ble/examples/lib/dev"
	"github.com/pkg/errors"
)

var uartServiceID = ble.MustParse("49535343fe7d4ae58fa99fafd205e455")
var uartServiceRXCharID = ble.MustParse("49535343884143f4a8d4ecbe34729bb3")
var uartServiceTXCharID = ble.MustParse("495353431e4d4bd9ba6123c647249616")

type readBytesCallbackType func(b []byte) error

var curr struct {
	device                ble.Device
	client                ble.Client
	receiveCharacteristic *ble.Characteristic
	debug                 bool
	callback              readBytesCallbackType
	connected             bool
}

func setup() error {
	if curr.device != nil {
		return nil
	}
	fmt.Printf("Initializing device ...\n")
	d, err := dev.NewDevice("default")
	if err != nil {
		return errors.Wrap(err, "can't init new device")
	}
	ble.SetDefaultDevice(d)
	curr.device = d
	return nil
}

func advHandler(a ble.Advertisement) {
	//curr.addr = a.Addr()
	if a.Connectable() {
		fmt.Printf("[%s]     Connectable %3d:", a.Addr(), a.RSSI())
	} else {
		fmt.Printf("[%s] Not Connectable %3d:", a.Addr(), a.RSSI())
	}
	comma := ""
	if len(a.LocalName()) > 0 {
		fmt.Printf(" Name: %s", a.LocalName())
		comma = ","
	}
	if len(a.Services()) > 0 {
		fmt.Printf("%s Svcs: %v", comma, a.Services())
		comma = ","
	}
	if len(a.ManufacturerData()) > 0 {
		fmt.Printf("%s MD: %X", comma, a.ManufacturerData())
	}

	fmt.Printf("\n")
	for _, s := range a.Services() {

		fmt.Println(s.String())
	}

	for _, s := range a.ServiceData() {
		fmt.Println(s.UUID)
		fmt.Println(s.Data)
	}
}

func advFilter() ble.AdvFilter {
	return func(a ble.Advertisement) bool {
		return true
	}
	/*
		return func(a ble.Advertisement) bool {
			for _, s := range a.Services() {
				if s.Equal(uartServiceID) {
					return true
				}
			}
			return false
		}
	*/
}

func chkErr(err error) error {
	switch errors.Cause(err) {
	case context.DeadlineExceeded:
		// Sepcified duration passed, which is the expected case.
		return nil
	case context.Canceled:
		fmt.Printf("\n(Canceled)\n")
		return nil
	}
	return err
}

func Scan() error {
	err := setup()
	if err != nil {
		return err
	}

	var allowDuplicates bool = false
	ctx := ble.WithSigHandler(context.WithCancel(context.Background()))
	return chkErr(ble.Scan(ctx, allowDuplicates, advHandler, advFilter()))
}

// Init a connection to the a bluetooth device with the specified name.
func Init(_device string, _callback readBytesCallbackType, _debug bool) error {
	curr.debug = _debug
	curr.callback = _callback
	curr.connected = false

	d, err := DefaultDevice()
	if err != nil {
		return fmt.Errorf("failed to open device, err: %s", err)
	}
	ble.SetDefaultDevice(d)

	// Default to search device with name specified by user
	filter := func(a ble.Advertisement) bool {
		return strings.ToUpper(a.LocalName()) == strings.ToUpper(_device)
	}

	ctx := ble.WithSigHandler(context.WithCancel(context.Background()))
	cln, err := ble.Connect(ctx, filter)
	if err != nil {
		log.Fatalf("cannot connect : %s", err)
	}

	// Make sure we had the chance to print out the message.
	done := make(chan struct{})
	// Normally, the connection is disconnected by us after our exploration.
	// However, it can be asynchronously disconnected by the remote peripheral.
	// So we wait(detect) the disconnection in the go routine.
	go func() {
		<-cln.Disconnected()
		curr.connected = false
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

		if s.UUID.Equal(uartServiceID) {
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
					curr.receiveCharacteristic = c
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
	curr.client = cln
	curr.connected = true

	return nil
}

// Stop the connection
func Stop() {
	if curr.client != nil {
		curr.client.CancelConnection()
	}
}

// IsConnected indicates whether a connection is present
func IsConnected() bool {
	return curr.connected
}

func WriteBytes(b *bytes.Buffer) error {
	if curr.debug {
		fmt.Printf("TX %d bytes: ", b.Len())

		for i := 0; i < b.Len(); i++ {
			fmt.Printf("0x%.2x, ", b.Bytes()[i])
		}
		fmt.Printf("\n")
	}

	err := curr.client.WriteCharacteristic(curr.receiveCharacteristic, b.Bytes(), true)
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
	if curr.debug {
		fmt.Printf("RX %d bytes: ", len(b))
		for i := 0; i < len(b); i++ {
			fmt.Printf("0x%.2x, ", b[i])
		}
		fmt.Printf("\n")
	}

	if curr.callback != nil {
		err := curr.callback(b)
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
