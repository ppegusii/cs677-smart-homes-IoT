package main

import (
	"flag"
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"github.com/ppegusii/cs677-smart-homes-IoT/lib"
	"fmt"
	"os"
)

func main() {
	//	parse args
	if len(os.Args) != 2 {
		fmt.Println("Usage: ", os.Args[0], "Synchronization Mode") //Discuss what exactly is needed
		os.Exit(1)
	}

	//	clock:= &os.Args[1]
	/*If different components are running on different IP's then get own IP from 
	loopback and non-loop back IP's. */
	ownIP := lib.GetOwnIP()
	var ip *string = &ownIP
	fmt.Println("Gateway IP is ",*ip)

	//	var ip *string = flag.String("i", "127.0.0.1", "IP address")
	var mode api.Mode = api.Mode(*(flag.Int("m", 0, "home=0,away=1")))
	var pollingInterval int = *flag.Int("P", 60, "polling interval in seconds")
	var port *string = flag.String("p", "6770", "port")
	flag.Parse()

	//start server
	var g *Gateway = newGateway(*ip, mode, pollingInterval, *port)
	g.start()
}
