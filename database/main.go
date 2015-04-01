package main

import (
	"github.com/ppegusii/cs677-smart-homes-IoT/lib"
	"flag"
	"fmt"
	"os"
)

func main() {
	//parse args
	if len(os.Args) != 2 {
		fmt.Println("Usage: ", os.Args[0], "Gateway IP") //IP address of Gateway
		os.Exit(1)
	}

//	gatewayIp:= &os.Args[1] TODO: We need to pass the gateway IP too.
	/*If different components are running on different IP's then get own IP from 
	loopback and non-loop back IP's. */
	ownIP := lib.GetOwnIP()
	var ip *string = &ownIP

	//	var ip *string = flag.String("i", "127.0.0.1", "IP address")
	var port *string = flag.String("p", "6777", "port")
	flag.Parse()

	//start server
	var d *Database = newDatabase(*ip, *port)
	d.start()
}
