package connection

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/JuulLabs-OSS/ble"
	"github.com/JuulLabs-OSS/ble/examples/lib/dev"
	"github.com/pkg/errors"
)

var uartServiceID = ble.MustParse("49535343fe7d4ae58fa99fafd205e455")
var uartServiceRXCharID = ble.MustParse("49535343884143f4a8d4ecbe34729bb3")
var uartServiceTXCharID = ble.MustParse("495353431e4d4bd9ba6123c647249616")

type readBytesCallbackType func(b []byte) error

var mutex sync.RWMutex
var devices map[string]Device

type Connection struct {
	device                ble.Device
	client                ble.Client
	profile               *ble.Profile
	receiveCharacteristic *ble.Characteristic
	debug                 bool
	callback              readBytesCallbackType
	connected             bool
}

type Device struct {
	Address  string    `json:"address"`
	Detected time.Time `json:"detected"`
	Since    string    `json:"since"`
	Name     string    `json:"name"`
	RSSI     int       `json:"rssi"`
	//Advertisement string    `json:"advertisement"`
	//ScanResponse  string    `json:"scanresponse"`
}

func (curr *Connection) setup() error {
	if devices == nil {
		devices = make(map[string]Device)
	}
	if curr.device != nil {
		return nil
	}
	fmt.Printf("Initializing interface...")
	var err error
	curr.device, err = dev.NewDevice("default")
	if err != nil {
		return errors.Wrap(err, "can't init new device")
	}
	ble.SetDefaultDevice(curr.device)
	fmt.Printf("complete\n")
	return nil
}

func adScanHandler(a ble.Advertisement) {
	mutex.Lock()
	device := Device{
		Address:  a.Addr().String(),
		Detected: time.Now(),
		Name:     clean(a.LocalName()),
		RSSI:     a.RSSI(),
		//Advertisement: formatHex(hex.EncodeToString(a.LEAdvertisingReportRaw())),
		//ScanResponse:  formatHex(hex.EncodeToString(a.ScanResponseRaw())),
	}
	devices[a.Addr().String()] = device
	mutex.Unlock()
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

var scanStop bool

// Scan for eligible devices and print details when they are found
func (curr *Connection) Scan(dur time.Duration) (map[string]Device, error) {
	err := curr.setup()

	if err != nil {
		return devices, err
	}

	scanStop = false
	go func() {
		var allowDuplicates bool = false
		for !scanStop {
			ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), dur))
			err = ble.Scan(ctx, allowDuplicates, adScanHandler, advFilter())
			if err != nil {
				return
			}
		}
		return
	}()

	return devices, nil
}

func (curr *Connection) StopScan() error {
	scanStop = true
	return nil
}

func (curr *Connection) ListDevices() map[string]Device {
	return devices
}

// Set the callback to be used when receiving bytes
func (curr *Connection) Callback(_callback readBytesCallbackType) {
	curr.callback = _callback
}

func (curr *Connection) connect(name string) error {
	curr.client = nil

	var cln ble.Client
	var err error

	// Default to search device with name specified by user
	filter := func(a ble.Advertisement) bool {
		return strings.ToUpper(a.LocalName()) == strings.ToUpper(name)
	}

	ctx := ble.WithSigHandler(context.WithCancel(context.Background()))
	if cln, err = ble.Connect(ctx, filter); err == nil {
		fmt.Printf("Connected to %s [%s]\n", name, cln.Addr())
	}

	if err == nil {
		curr.client = cln

		// Make sure we had the chance to print out the disconnected message.
		done := make(chan struct{})
		go func() {
			<-cln.Disconnected()
			curr.client = nil
			curr.connected = false
			fmt.Printf("\n%s disconnected\n", cln.Addr().String())
			close(done)
		}()
	}
	return err
}

// Init a connection to the a bluetooth device with the specified name.
func (curr *Connection) Init(_device string, _callback readBytesCallbackType, _debug bool) error {
	curr.debug = _debug
	curr.callback = _callback
	curr.connected = false

	if err := curr.setup(); err != nil {
		return err
	}

	if err := curr.connect(_device); err != nil {
		return err
	}

	fmt.Printf("Discovering profile...")
	p, err := curr.client.DiscoverProfile(true)
	if err != nil {
		log.Fatalf("can't discover profile: %s", err)
	}
	curr.profile = p
	fmt.Printf("complete\n")

	if curr.debug {
		for _, s := range curr.profile.Services {
			fmt.Printf("    Service: %s %s, Handle (0x%02X)\n", s.UUID, ble.Name(s.UUID), s.Handle)
			for _, c := range s.Characteristics {
				fmt.Printf("      Characteristic: %s %s, Property: 0x%02X (%s), Handle(0x%02X), VHandle(0x%02X)\n",
					c.UUID, ble.Name(c.UUID), c.Property, propString(c.Property), c.Handle, c.ValueHandle)
			}
		}
	}

	if u := curr.profile.Find(ble.NewCharacteristic(uartServiceTXCharID)); u != nil {
		if curr.debug {
			fmt.Println("Found TX Characteristic")
		}
		indication := false
		if err := curr.client.Subscribe(u.(*ble.Characteristic), indication, curr.readBytes); err != nil {
			log.Fatalf("subscribe failed: %s", err)
		}
	} else if u == nil {
		return fmt.Errorf("Could not find TX Characteristic")
	}

	if u := curr.profile.Find(ble.NewCharacteristic(uartServiceRXCharID)); u != nil {
		if curr.debug {
			fmt.Println("Found RX Characteristic")
		}
		curr.receiveCharacteristic = u.(*ble.Characteristic)
	} else if u == nil {
		return fmt.Errorf("Could not find RX Characteristic")
	}

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

	var noResp bool = true
	err := curr.client.WriteCharacteristic(curr.receiveCharacteristic, b.Bytes(), noResp)
	if err != nil {
		return err
	}

	return nil
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

// reformat string for proper display of hex
func formatHex(instr string) (outstr string) {
	outstr = ""
	for i := range instr {
		if i%2 == 0 {
			outstr += instr[i:i+2] + " "
		}
	}
	return
}

// clean up the non-ASCII characters
func clean(input string) string {
	return strings.TrimFunc(input, func(r rune) bool {
		return !unicode.IsGraphic(r)
	})
}
