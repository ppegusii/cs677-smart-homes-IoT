package main

import (
	"flag"
	"log"
	//	"fmt"
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"github.com/ppegusii/cs677-smart-homes-IoT/util"
	//	"os"
)

func main() {
	//parse args
	//	if len(os.Args) != 2 {
	//		fmt.Println("Usage: ", os.Args[0], "Gateway IP") //IP address of Gateway
	//		os.Exit(1)
	//	}

	//	gatewayIp := &os.Args[1]
	/*If different components are running on different IP's then get own IP from
	loopback and non-loop back IP's. */
	//	ownIP := util.GetOwnIP()
	//	var selfIp *string = &ownIP

	var gatewayIp *string = flag.String("i", "127.0.0.1", "gateway IP address")
	var gatewayPort *string = flag.String("p", "6770", "gateway TCP port")
	var selfIp *string = flag.String("I", "127.0.0.1", "IP address")
	var selfPort *string = flag.String("P", "6773", "TCP port")
	var order *string = flag.String("o", "n", "none=n,clock sync=c,logical clocks=l")
	flag.Parse()

	//start sensor
	var ordering api.Ordering
	var err error
	ordering, err = util.StringToOrdering(*order)
	if err != nil {
		log.Fatal(err)
	}
	var t *TemperatureSensor = newTemperatureSensor(0.0, *gatewayIp, *gatewayPort, *selfIp, *selfPort, ordering)
	t.start()
}
