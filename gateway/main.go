package main

import (
	"flag"
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	//"github.com/ppegusii/cs677-smart-homes-IoT/gatewaywrapper"
	"github.com/ppegusii/cs677-smart-homes-IoT/gatewayleader"
	"github.com/ppegusii/cs677-smart-homes-IoT/util"
	"log"
)

func main() {
	//	parse args
	var ip *string = flag.String("i", "127.0.0.1", "IP address")
	var modeInt *int = flag.Int("m", 2, "home=2,away=1")
	var pollingInterval *int = flag.Int("s", 60, "polling interval in seconds")
	var port *string = flag.String("p", "6770", "port")
	var dbIP *string = flag.String("I", "127.0.0.1", "database IP address")
	var dbPort *string = flag.String("P", "6777", "database port")
	var replicaIp *string = flag.String("ri", "127.0.0.1", "gateway replica IP")
	var replicaPort *string = flag.String("rp", "", "gateway replica port")
	//var order *string = flag.String("o", "n", "none=n,clock sync=c,logical clocks=l, fault tolerant=f")
	// TODO remove the following
	flag.String("o", "n", "none=n,clock sync=c,logical clocks=l, fault tolerant=f")
	flag.Parse()
	var mode api.Mode = api.Mode(*modeInt)

	//start server
	//var ordering api.Ordering
	var err error
	//ordering, err = util.StringToOrdering(*order)
	if err != nil {
		log.Fatal(err)
	}
	/*
		var gw api.GatewayWrapperInterface = gatewaywrapper.NewGatewayWrapper()
		var g api.GatewayInterface = newGateway(*dbIP, *dbPort, *ip, mode, *pollingInterval, *port, gw)
		gw.SetGateway(g)
		util.RpcRegister(gw, *ip, *port, "Gateway", false)
		g.Start()
	*/
	var replica = api.RegisterGatewayUserParams{
		Address: *replicaIp,
		Port:    *replicaPort,
	}
	var replicas = []api.RegisterGatewayUserParams{replica}
	var gl api.GatewayLeaderInterface = gatewayleader.NewGatewayLeader(*ip, *port, *dbIP, *dbPort, replicas)
	var g api.GatewayInterface = newGateway(*dbIP, *dbPort, *ip, mode, *pollingInterval, *port, gl)
	gl.SetGateway(g)
	util.RpcRegister(gl, *ip, *port, "Gateway", false)
	gl.StartLeader()
	g.Start()
}
