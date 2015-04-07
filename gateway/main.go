package main

import (
	"flag"
	//	"fmt"
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"github.com/ppegusii/cs677-smart-homes-IoT/util"
	"log"
	//	"os"
)

func main() {
	//	parse args
	//	if len(os.Args) != 2 {
	//		fmt.Println("Usage: ", os.Args[0], "Synchronization Mode") //Discuss what exactly is needed
	//		os.Exit(1)
	//	}

	//	clock:= &os.Args[1]
	/*If different components are running on different IP's then get own IP from
	loopback and non-loop back IP's. */
	//	ownIP := util.GetOwnIP()
	//	var ip *string = &ownIP
	//	fmt.Println("Gateway IP is ", *ip)

	var ip *string = flag.String("i", "127.0.0.1", "IP address")
	var mode api.Mode = api.Mode(*(flag.Int("m", 1, "home=1,away=0")))
	var pollingInterval int = *flag.Int("s", 60, "polling interval in seconds")
	var port *string = flag.String("p", "6770", "port")
	var dbIP *string = flag.String("I", "127.0.0.1", "database IP address")
	var dbPort *string = flag.String("P", "6777", "database port")
	flag.Parse()
	var order *string = flag.String("o", "n", "none=n,clock sync=c,logical clocks=l")

	//start server
	var ordering api.Ordering
	var err error
	ordering, err = util.StringToOrdering(*order)
	if err != nil {
		log.Fatal(err)
	}
	var g *Gateway = newGateway(*dbIP, *dbPort, *ip, mode, pollingInterval, *port, ordering)
	g.start()
}
