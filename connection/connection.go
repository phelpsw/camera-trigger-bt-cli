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

type Connection struct {
	device                ble.Device
	client                ble.Client
	receiveCharacteristic *ble.Characteristic
	debug                 bool
	callback              readBytesCallbackType
	connected             bool
}

func (curr *Connection) setup() error {
	if curr.device != nil {
		return nil
	}
	fmt.Printf("Initializing interface...")
	d, err := dev.NewDevice("default")
	if err != nil {
		return errors.Wrap(err, "can't init new device")
	}
	ble.SetDefaultDevice(d)
	curr.device = d
	fmt.Printf("complete\n")
	return nil
}

func advHandler(a ble.Advertisement) {
	if a.Connectable() {
		fmt.Printf("[%s]     Connectable %3d dBm:", a.Addr(), a.RSSI())
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
}

func advFilter() ble.AdvFilter {
	return func(a ble.Advertisement) bool {
		if strings.HasPrefix(a.LocalName(), "camera-trigger-") {
			return true
		}
		return false
	}
}

func chkErr(err error) error {
	switch errors.Cause(err) {
	case context.DeadlineExceeded:
		return nil
	case context.Canceled:
		fmt.Printf("\n(Canceled)\n")
		return nil
	}
	return err
}

// Scan for eligible devices and print details when they are found
func (curr *Connection) Scan() error {
	err := curr.setup()
	if err != nil {
		return err
	}

	var allowDuplicates bool = false
	ctx := ble.WithSigHandler(context.WithCancel(context.Background()))
	return chkErr(ble.Scan(ctx, allowDuplicates, advHandler, advFilter()))
}

// Set the callback to be used when receiving bytes
func (curr *Connection) Callback(_callback readBytesCallbackType) {
	curr.callback = _callback
}

// Init a connection to the a bluetooth device with the specified name.
func (curr *Connection) Init(_device string, _callback readBytesCallbackType, _debug bool) error {
	curr.debug = _debug
	curr.callback = _callback
	curr.connected = false

	err := curr.setup()
	if err != nil {
		return err
	}

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

	fmt.Printf("Discovering profile...")
	p, err := cln.DiscoverProfile(true)
	if err != nil {
		log.Fatalf("can't discover profile: %s", err)
	}
	fmt.Printf("complete\n")

	for _, s := range p.Services {
		if curr.debug {
			fmt.Printf("    Service: %s %s, Handle (0x%02X)\n", s.UUID, ble.Name(s.UUID), s.Handle)
			for _, c := range s.Characteristics {
				fmt.Printf("      Characteristic: %s %s, Property: 0x%02X (%s), Handle(0x%02X), VHandle(0x%02X)\n",
					c.UUID, ble.Name(c.UUID), c.Property, propString(c.Property), c.Handle, c.ValueHandle)
			}
		}

		if s.UUID.Equal(uartServiceID) {
			for _, c := range s.Characteristics {
				if (c.Property & ble.CharNotify) != 0 {
					if c.UUID.Equal(uartServiceTXCharID) {
						if err := cln.Subscribe(c, false, curr.readBytes); err != nil {
							log.Fatalf("subscribe failed: %s", err)
						}
					}
				}

				if c.UUID.Equal(uartServiceRXCharID) {
					curr.receiveCharacteristic = c
				}
			}
		}
	}

	curr.client = cln
	curr.connected = true

	return nil
}

// Stop the connection
func (curr *Connection) Stop() {
	if curr.client != nil {
		curr.client.CancelConnection()
	}
}

// IsConnected indicates whether a connection is present
func (curr *Connection) IsConnected() bool {
	return curr.connected
}

func (curr *Connection) WriteBytes(b *bytes.Buffer) error {
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

func (curr *Connection) readBytes(b []byte) {
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
}

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
