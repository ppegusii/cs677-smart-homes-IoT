package main

import (
	"flag"
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
)

func main() {
	//parse args
	var ip *string = flag.String("i", "127.0.0.1", "IP address")
	var mode api.Mode = api.Mode(*(flag.Int("m", 0, "home=0,away=1")))
	var pollingInterval int = *flag.Int("P", 60, "polling interval in seconds")
	var port *string = flag.String("p", "6770", "port")
	flag.Parse()

	//start server
	var g *Gateway = newGateway(*ip, mode, pollingInterval, *port)
	g.start()
}
