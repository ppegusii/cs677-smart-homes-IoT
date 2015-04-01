package main

import (
	"flag"
<<<<<<< HEAD
	"fmt"
	"os"
	"github.com/ppegusii/cs677-smart-homes-IoT/lib"
=======
>>>>>>> 9055a109108b4e68213c1efbed068d64bb994c34
)

func main() {
	//parse args
<<<<<<< HEAD
	if len(os.Args) != 2 {
		fmt.Println("Usage: ", os.Args[0], "Gateway IP") //IP address of Gateway
		os.Exit(1)
	}

	gatewayIp:= &os.Args[1]
	/*If different components are running on different IP's then get own IP from 
	loopback and non-loop back IP's. */
	ownIP := lib.GetOwnIP()
	var selfIp *string = &ownIP

	//var gatewayIp *string = flag.String("i", "127.0.0.1", "gateway IP address")
	var gatewayPort *string = flag.String("p", "6770", "gateway TCP port")
	//var selfIp *string = flag.String("I", "127.0.0.1", "IP address")
=======
	var gatewayIp *string = flag.String("i", "127.0.0.1", "gateway IP address")
	var gatewayPort *string = flag.String("p", "6770", "gateway TCP port")
	var selfIp *string = flag.String("I", "127.0.0.1", "IP address")
>>>>>>> 9055a109108b4e68213c1efbed068d64bb994c34
	var selfPort *string = flag.String("P", "6774", "TCP port")
	flag.Parse()

	//start smartoutlet
	var s *SmartOutlet = newSmartOutlet(*gatewayIp, *gatewayPort, *selfIp, *selfPort)
	s.start()
}