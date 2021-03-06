package main

import (
	"flag"
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"github.com/ppegusii/cs677-smart-homes-IoT/util"
	"log"
)

func main() {
	//parse args
	var gatewayIp *string = flag.String("I", "127.0.0.1", "gateway IP address")
	var gatewayPort *string = flag.String("P", "6770", "gateway TCP port")
	var gatewayIp2 *string = flag.String("I2", "127.0.0.1", "gateway IP address")
	var gatewayPort2 *string = flag.String("P2", "9999", "gateway TCP port")
	var selfIp *string = flag.String("i", "127.0.0.1", "IP address")
	var selfPort *string = flag.String("p", "6776", "TCP port")
	var order *string = flag.String("o", "n", "none=n,clock sync=c,logical clocks=l")
	flag.Parse()

	//start sensor
	var ordering api.Ordering
	var err error
	ordering, err = util.StringToOrdering(*order)
	if err != nil {
		log.Fatal(err)
	}
	var d *DoorSensor = newDoorSensor(*gatewayIp, *gatewayPort, *gatewayIp2, *gatewayPort2, *selfIp, *selfPort, ordering)
	d.start()
}
