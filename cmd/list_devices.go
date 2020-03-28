package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/backwardn/gatt"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List Devices",
	Long:  "List all valid device ids for command and control.",
	Run:   discover,
}

var done = make(chan struct{})

var uartServiceID = gatt.MustParseUUID("49535343-fe7d-4ae5-8fa9-9fafd205e455")
var uartServiceRXCharID = gatt.MustParseUUID("49535343-8841-43f4-a8d4-ecbe34729bb3")
var uartServiceTXCharID = gatt.MustParseUUID("49535343-1e4d-4bd9-ba61-23c647249616")
var devicePeripheral gatt.Peripheral
var receiveCharacteristic *gatt.Characteristic

func onStateChanged(d gatt.Device, s gatt.State) {
	log.Println("State:", s)
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
	}

	if a.LocalName == deviceID {
		p.Device().StopScanning()
		p.Device().Connect(p)
	}

	return
}

func onPeriphConnected(p gatt.Peripheral, err error) {
	log.Printf("Peripheral connected\n")

	services, err := p.DiscoverServices(nil)
	if err != nil {
		log.Printf("Failed to discover services, err: %s\n", err)
		return
	}

	for _, service := range services {
		if service.Name() == "" {
			log.Printf("Service Found: Unknown (%s)\n", service.UUID())
		} else {
			log.Printf("service found: %s (%s)\n",
				service.Name(),
				service.UUID())
		}

		if service.UUID().Equal(uartServiceID) {
			log.Printf("Expected service found: %s (%s)\n",
				service.Name(),
				service.UUID())

			cs, _ := p.DiscoverCharacteristics(nil, service)

			for _, c := range cs {
				log.Printf("Characteristic %s\n", c.UUID())

				if c.UUID().Equal(uartServiceTXCharID) {
					log.Println("TX Characteristic Found")
					p.DiscoverDescriptors(nil, c)
					p.SetNotifyValue(c, handlePacket)
				} else if c.UUID().Equal(uartServiceRXCharID) {
					log.Println("RX Characteristic Found")
					devicePeripheral = p
					receiveCharacteristic = c
				}
			}
		}
	}

	return
}

func handlePacket(c *gatt.Characteristic, b []byte, e error) {
	fmt.Printf("Got back (%d bytes) ", len(b))
	for i := 0; i < len(b); i++ {
		fmt.Printf("0x%.2x, ", b[i])
	}
	fmt.Printf("\n")

	// TODO: Add to buffer and then repeatedly attempt to read messages
	//       from this buffer.  For larger packets, multiple reads are
	//       necessary
	//messages.ReadMessage(b)

	return
}

func onPeriphDisconnected(p gatt.Peripheral, err error) {
	log.Println("Disconnected")
}

func discover(cmd *cobra.Command, args []string) {
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
	)

	d.Init(onStateChanged)
	<-done
	log.Println("Done")
}
