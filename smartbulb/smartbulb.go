// This file declares all the structs and interfaces needed by smartbulb
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

// This struct contains all the attributes of the smart bulb and information needed for
// ordering for clock synchronization, peer table to keep a track of ip of the peers
// and reference to its middleware
type SmartBulb struct {
	id          int
	gatewayIp   string
	gatewayPort string
	gatewayIp2   string
	gatewayPort2 string
	ordering    api.Ordering
	orderMW     api.OrderingMiddlewareInterface
	selfIp      string
	selfPort    string
	state       structs.SyncState
	gRPCIp 		 string
	gRPCPort	 string
}

// create and initialize a new smart bulb
func newSmartBulb(gatewayIp string, gatewayPort string, gatewayIp2 string, gatewayPort2 string, selfIp string, selfPort string, ordering api.Ordering) *SmartBulb {
	return &SmartBulb{
		gatewayIp:   gatewayIp,
		gatewayPort: gatewayPort,
		gatewayIp2:   gatewayIp2,
		gatewayPort2: gatewayPort2,
		ordering:    ordering,
		selfIp:      selfIp,
		selfPort:    selfPort,
		state:       *structs.NewSyncState(api.Off),
	}
}

func (s *SmartBulb) start() {
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

	s.id = regresponse.DeviceId
	s.gRPCIp = regresponse.Address
	s.gRPCPort = regresponse.Port
	log.Printf("Device id: %d %s %s", s.id, s.gRPCIp, s.gRPCPort)

/*	client, err = rpc.Dial("tcp", s.gatewayIp+":"+s.gatewayPort)
	if err != nil {
		log.Fatal("dialing error: %+v", err)
	}
	err = client.Call("Gateway.Register", &api.RegisterParams{Type: api.Device, Name: api.Bulb, Address: s.selfIp, Port: s.selfPort}, &s.id)
	if err != nil {
		log.Fatal("calling error: %+v", err)
	}
*/
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

//This is an RPC function that is issued by the gateway to get the state of the SmartBulb
func (s *SmartBulb) QueryState(params *int, reply *api.StateInfo) error {
	reply.DeviceId = s.id
	reply.State = s.state.GetState()
	//go s.sendState()
	return nil
}

//RPC stub to change state remotely;it is called by the gateway to change the state of the smartbulb
// based on the current state of the motion sensor
func (s *SmartBulb) ChangeState(params *api.StateInfo, reply *api.StateInfo) error {
	log.Printf("Received change state request with info: %+v", params)
	s.state.SetState(params.State)
	util.LogCurrentState(s.state.GetState())
	reply.DeviceId = s.id
	reply.State = params.State
	//go s.sendState()
	return nil
}

// sendState() is used to report state to the middleware
/*
func (s *SmartBulb) sendState() {
	var err error = s.orderMW.SendState(api.StateInfo{DeviceId: s.id, DeviceName: api.Bulb, State: s.state.GetState()}, s.gatewayIp, s.gatewayPort)
	if err != nil {
		log.Printf("Error sending state: %+v", err)
	}
}
*/
