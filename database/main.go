package main

import (
	"flag"
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
)

func main() {
	//parse args
	var ip *string = flag.String("i", "127.0.0.1", "IP address")
	var ordering *bool = flag.Bool("o", true, "clock sync")
	var port *string = flag.String("p", "6777", "port")
	flag.Parse()

	//start server
	var orderMode api.Mode
	if *ordering {
		orderMode = api.Time
	} else {
		orderMode = api.Logical
	}
	var d *Database = newDatabase(*ip, *port, orderMode)
	d.start()
}
