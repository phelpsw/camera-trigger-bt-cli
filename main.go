package main

import (
    "bytes"
    "encoding/binary"
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

var devicePeripheral gatt.Peripheral
var receiveCharacteristic *gatt.Characteristic
type Message struct {
    // For now lets put together a generic trigger type message to parse
    temperature float32
}
var bm71ReceiveChan = make(chan Message)

func onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
    log.Printf("Discovered %s\n", a.LocalName)

    if (a.LocalName == "camera-trigger-001") {
        log.Printf("Peripheral Discovered: %s \n", p.Name())
        p.Device().StopScanning()
        p.Device().Connect(p)
    }

    return
}

func readFloat32(b []byte) (float32, error) {
    var val float32
    buf := bytes.NewReader(b)
    err := binary.Read(buf, binary.LittleEndian, &val)
    if err != nil {
        return 0, err
    }
    return val, nil
}

func handlePacket(c *gatt.Characteristic, b []byte, e error) {
    log.Printf("Got back %x\n", b)
    return
}

func sendPacket(ch chan Message, stop chan bool) {
    for {
        select {
            case msg := <- ch:
                if receiveCharacteristic == nil {
                    continue
                }
                if devicePeripheral == nil {
                    continue
                }

                log.Println(msg)

                devicePeripheral.WriteCharacteristic(receiveCharacteristic,
                                                     []byte{0x74},
                                                     true)
                log.Printf("Wrote %s\n", string([]byte{0x74}))
            case <- stop:
                break
        }
    }
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
                    p.SetNotifyValue(c, handlePacket)
                } else if (c.UUID().Equal(uartServiceRXCharId)) {
                    log.Println("RX Characteristic Found")
                    devicePeripheral = p
                    receiveCharacteristic = c
                    //p.WriteCharacteristic(c, []byte{0x74}, true)
                    //log.Printf("Wrote %s\n", string([]byte{0x74}))
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
    if receiveCharacteristic == nil {
        log.Println("nil found")
    }

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

    go 

    d.Init(onStateChanged)
    <-done
    log.Println("Done")
}
