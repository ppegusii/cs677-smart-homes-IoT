package main

import (
	"flag"
	//	"fmt"
	//	"github.com/ppegusii/cs677-smart-homes-IoT/util"
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
	var selfPort *string = flag.String("P", "6771", "TCP port")
	flag.Parse()

	//start sensor
	var m *MotionSensor = newMotionSensor(*gatewayIp, *gatewayPort, *selfIp, *selfPort)
	m.start()
}
