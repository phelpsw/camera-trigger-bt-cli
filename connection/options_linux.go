package connection

import "github.com/photostorm/gatt"

var DefaultClientOptions = []gatt.Option{
	gatt.LnxMaxConnections(1),
	gatt.LnxDeviceID(-1, false),
}
