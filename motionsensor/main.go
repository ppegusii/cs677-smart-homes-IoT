package main

import (
	"flag"
)

func main() {
	//parse args
	var gatewayIp *string = flag.String("i", "127.0.0.1", "gateway IP address")
	var gatewayPort *string = flag.String("p", "6770", "gateway TCP port")
	var selfIp *string = flag.String("I", "127.0.0.1", "IP address")
	var selfPort *string = flag.String("P", "6771", "TCP port")
	flag.Parse()

	//start sensor
	var m *MotionSensor = newMotionSensor(*gatewayIp, *gatewayPort, *selfIp, *selfPort)
	m.start()
}
