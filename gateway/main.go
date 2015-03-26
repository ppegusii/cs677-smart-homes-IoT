package main

import (
	"flag"
)

func main() {
	//parse args
	var mode Mode = Mode(*(flag.Int("m", 0, "home=0,away=1")))
	var pollingInterval int = *flag.Int("P", 60, "polling interval in seconds")
	var port *string = flag.String("p", "6770", "port")
	flag.Parse()

	//start server
	var g *Gateway = newGateway(mode, pollingInterval, *port)
	g.start()
}
