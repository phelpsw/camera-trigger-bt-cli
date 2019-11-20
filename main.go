package main

import (
  "log"
  "github.com/backwardn/gatt"
)

var done = make(chan struct{})

/*
sudo setcap 'cap_net_raw,cap_net_admin=eip' bluetooth-test
*/

var uartServiceId = gatt.MustParseUUID("49535343-fe7d-4ae5-8fa9-9fafd205e455")
var uartServiceRXCharId = gatt.MustParseUUID("49535343-8841-43f4-a8d4-ecbe34729bb3")
var uartServiceTXCharId = gatt.MustParseUUID("49535343-1e4d-4bd9-ba61-23c647249616")

func onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
    log.Printf("Discovered %s\n", a.LocalName)

    if (a.LocalName == "camera-trigger-001") {
        log.Printf("Peripheral Discovered: %s \n", p.Name())
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

        if (service.UUID().Equal(uartServiceId)) {
            log.Printf("Expected service found: %s (%s)\n",
                       service.Name(),
                       service.UUID())

            cs, _ := p.DiscoverCharacteristics(nil, service)

            for _, c := range cs {
                log.Printf("Characteristic %s\n", c.UUID())

                if (c.UUID().Equal(uartServiceTXCharId)) {
                    log.Println("TX Characteristic Found")

                    p.DiscoverDescriptors(nil, c)

                    p.SetNotifyValue(c, func(c *gatt.Characteristic, b []byte, e error) {
                        log.Printf("Got back %x\n", b)
                    })
                } else if (c.UUID().Equal(uartServiceRXCharId)) {
                    log.Println("RX Characteristic Found")
                    p.WriteCharacteristic(c, []byte{0x74}, true)
                    log.Printf("Wrote %s\n", string([]byte{0x74}))
                }
            }
        }
    }

    return
}

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

func onPeriphDisconnected(p gatt.Peripheral, err error) {
    log.Println("Disconnected")
}

func main() {
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
    <-done
    log.Println("Done")
}
