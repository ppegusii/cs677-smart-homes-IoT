package main

import (
	"flag"
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"github.com/ppegusii/cs677-smart-homes-IoT/util"
	"log"
)

func main() {
	//parse args
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
