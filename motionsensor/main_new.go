package main

import (
	"flag"
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"github.com/ppegusii/cs677-smart-homes-IoT/structs"
	"github.com/ppegusii/cs677-smart-homes-IoT/node"
//	"log"
)

func main() {
	//parse args
	var gatewayIp *string = flag.String("I", "127.0.0.1", "gateway IP address")
	var gatewayPort *string = flag.String("P", "6770", "gateway TCP port")
	var gatewayIp2 *string = flag.String("I2", "127.0.0.1", "gateway IP address")
	var gatewayPort2 *string = flag.String("P2", "6778", "gateway TCP port")
	var selfIp *string = flag.String("i", "127.0.0.1", "IP address")
	var selfPort *string = flag.String("p", "6771", "TCP port")
//	var order *string = flag.String("o", "n", "none=n,clock sync=c,logical clocks=l,fault tolerant =f")
	flag.Parse()

	//start sensor
/*	var ordering api.Ordering
	var err error
	ordering, err = util.StringToOrdering(*order)
	if err != nil {
		log.Fatal(err)
	}
*/

//Amee: Not sure if the below 2 lines are correct. Here, the problem is for newnode the id and state are needed
// which we still do not have since device has not registered

	var n api.NodeInterface = node.NewNode(-1,*structs.NewSyncState(api.MotionStop))
// How to specify motion sensor, should api have MotionSensorInterface and then use that here?
	var m api.NodeInterface = newMotionSensor(*gatewayIp, *gatewayPort, *gatewayIp2, *gatewayPort2, *selfIp, *selfPort, ordering, n)
	m.start()
}