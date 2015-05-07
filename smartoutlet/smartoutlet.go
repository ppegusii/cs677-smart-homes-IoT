package main

import (
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
//	"github.com/ppegusii/cs677-smart-homes-IoT/ordermw"
	"github.com/ppegusii/cs677-smart-homes-IoT/structs"
	"github.com/ppegusii/cs677-smart-homes-IoT/util"
	"log"
	"net"
	"net/rpc"
)

// This struct contains all the attributes of the smart outlet and information needed for
// ordering for clock synchronization, peer table to keep a track of ip of the peers
// and reference to its middleware
type SmartOutlet struct {
	id          *structs.SyncInt
	gatewayIp   string
	gatewayPort string
	gatewayIp2   string
	gatewayPort2 string
	ordering    api.Ordering
	orderMW     api.OrderingMiddlewareInterface
	selfIp      string
	selfPort    string
	state       structs.SyncState
	greplica 	*structs.SyncRegGatewayUserParam // This is the gateway replica assigned for load balancing
}

// create and initialize a new smart outlet
func newSmartOutlet(gatewayIp string, gatewayPort string, gatewayIp2 string, gatewayPort2 string, selfIp string, selfPort string, ordering api.Ordering) *SmartOutlet {
	return &SmartOutlet{
		id:			 structs.NewSyncInt(api.UNREGISTERED),
		gatewayIp:   gatewayIp,
		gatewayPort: gatewayPort,
		gatewayIp2:   gatewayIp2,
		gatewayPort2: gatewayPort2,
		ordering:    ordering,
		selfIp:      selfIp,
		selfPort:    selfPort,
		state:       *structs.NewSyncState(api.Off),
		greplica:	 structs.NewSyncRegGatewayUserParam(),
	}
}

func (s *SmartOutlet) start() {
	//register with gateway
	var client *rpc.Client
	var err error
	var regresponse *api.RegisterReturn

		// Dial to the first gateway
		client, err = rpc.Dial("tcp", s.gatewayIp+":"+s.gatewayPort)
		if err != nil {
			log.Fatal("dialing error: %+v", err)
		}
	replycall1 := client.Go("Gateway.Register", &api.RegisterParams{Type: api.Sensor, Name: api.Motion, Address: s.selfIp, Port: s.selfPort}, &regresponse, nil)
	id1 :=  <-replycall1.Done

		// Dial to the second gateway
		client, err = rpc.Dial("tcp", s.gatewayIp2+":"+s.gatewayPort2)
		if err != nil {
			log.Fatal("dialing error: %+v", err)
		}
	replycall2 := client.Go("Gateway.Register", &api.RegisterParams{Type: api.Sensor, Name: api.Motion, Address: s.selfIp, Port: s.selfPort}, &regresponse, nil)
	id2 :=  <-replycall2.Done

	if((id1 != nil) || (id2 != nil)) {
		log.Println("Registering with the gateway")
	} else {
		log.Println("Register RPC call return value: ",id1, id2)
	}

	s.id.Set(regresponse.DeviceId)
	s.greplica.Set(api.RegisterGatewayUserParams{Address: regresponse.Address, Port: regresponse.Port})
	replica := s.greplica.Get()
	log.Printf("Device id: %d %s %s", s.id.Get(), replica.Address, replica.Port)

/*	client, err = rpc.Dial("tcp", s.gatewayIp+":"+s.gatewayPort)
	if err != nil {
		log.Fatal("dialing error: %+v", err)
	}
	err = client.Call("Gateway.Register", &api.RegisterParams{Type: api.Device, Name: api.Outlet, Address: s.selfIp, Port: s.selfPort}, &s.id)
	if err != nil {
		log.Fatal("calling error: %+v", err)
	}
*/
//	log.Printf("Device id: %d", s.id)
	//initialize middleware
/*	s.orderMW = ordermw.GetOrderingMiddleware(s.ordering, s.id, s.selfIp, s.selfPort)

	//send acknowledgment of registration
	var empty struct{}
	client, err = rpc.Dial("tcp", s.gatewayIp+":"+s.gatewayPort)
	if err != nil {
		log.Printf("dialing error: %+v", err)
		return
	}
	client.Go("Gateway.RegisterAck", s.id, &empty, nil)
*/
	//start RPC server
	err = rpc.Register(api.DeviceInterface(s))
	if err != nil {
		log.Fatal("rpc.Register error: %s\n", err)
	}
	var listener net.Listener
	listener, err = net.Listen("tcp", s.selfIp+":"+s.selfPort)
	if err != nil {
		log.Fatal("net.Listen error: %s\n", err)
	}
	util.LogCurrentState(s.state.GetState())
	rpc.Accept(listener)
}

//This is an RPC function that is issued by the gateway to get the state of the Smart Outlet
func (s *SmartOutlet) QueryState(params *int, reply *api.StateInfo) error {
	//this will not be called in practice
	reply.DeviceId = s.id.Get()
	reply.State = s.state.GetState()
	//go s.sendState()
	return nil
}

//RPC stub to change state remotely;it is called by the gateway to change the state of the smartoutlet
// based on the current state of the temperature sensor
func (s *SmartOutlet) ChangeState(params *api.StateInfo, reply *api.StateInfo) error {
	log.Printf("Received change state request with info: %+v", params)
	s.state.SetState(params.State)
	util.LogCurrentState(s.state.GetState())
	reply.DeviceId = s.id.Get()
	reply.State = params.State
	//go s.sendState()
	return nil
}

//sendState() is used to report state to the middleware
/*
func (s *SmartOutlet) sendState() {
	var err error = s.orderMW.SendState(api.StateInfo{DeviceId: s.id, DeviceName: api.Outlet, State: s.state.GetState()}, s.gatewayIp, s.gatewayPort)
	if err != nil {
		log.Printf("Error sending state: %+v", err)
	}
}
*/

// This is an RPC function that is issued by the gateway to update the address port of the 
// loadsharing gateway the device is talking to. It returns the device id
func (s *SmartOutlet) ChangeGateway(params *api.RegisterGatewayUserParams, reply *int) error {
	s.greplica.Set(api.RegisterGatewayUserParams{Address: params.Address, Port: params.Port})
	*reply = s.id.Get()
	return nil
}