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
	//parse args
	//if len(os.Args) != 2 {
	//	fmt.Println("Usage: ", os.Args[0], "Gateway IP") //IP address of Gateway
	//	os.Exit(1)
	//}

	//	gatewayIp:= &os.Args[1] TODO: We need to pass the gateway IP too.

	//Note from Patrick: I planned to have the gateway register with the database
	//So passing the gateway IP:port would not be necessary.

	/*If different components are running on different IP's then get own IP from
	loopback and non-loop back IP's. */
	//ownIP := util.GetOwnIP()
	//var ip *string = &ownIP

	var ip *string = flag.String("i", "127.0.0.1", "IP address")
	var order *string = flag.String("o", "n", "none=n,clock sync=c,logical clocks=l")
	var port *string = flag.String("p", "6777", "port")
	flag.Parse()

	//start server
	var ordering api.Ordering
	var err error
	ordering, err = util.StringToOrdering(*order)
	if err != nil {
		log.Fatal(err)
	}
	var d *Database = newDatabase(*ip, *port, ordering)
	d.start()
}
