package main

import (
	"flag"
)

func main() {
	//parse args
	var ip *string = flag.String("i", "127.0.0.1", "IP address")
	var port *string = flag.String("p", "6777", "port")
	flag.Parse()

	//start server
	var d *Database = newDatabase(*ip, *port)
	d.start()
}
