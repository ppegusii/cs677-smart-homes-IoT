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

// create and initialize a new smart bulb
func newSmartBulb(gatewayIp string, gatewayPort string, gatewayIp2 string, gatewayPort2 string, selfIp string, selfPort string, ordering api.Ordering) *SmartBulb {
	return &SmartBulb{
		id:			  structs.NewSyncInt(api.UNREGISTERED),
		gatewayIp:   gatewayIp,
		gatewayPort: gatewayPort,
		gatewayIp2:   gatewayIp2,
		gatewayPort2: gatewayPort2,
		ordering:    ordering,
		selfIp:      selfIp,
		selfPort:    selfPort,
		state:       *structs.NewSyncState(api.Off),
		greplica:	  structs.NewSyncRegGatewayUserParam(),
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

	s.id.Set(regresponse.DeviceId)
	s.greplica.Set(api.RegisterGatewayUserParams{Address: regresponse.Address, Port: regresponse.Port})
	replica := s.greplica.Get()
	log.Printf("Device id: %d %s %s", s.id.Get(), replica.Address, replica.Port)

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
	reply.DeviceId = s.id.Get()
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
	reply.DeviceId = s.id.Get()
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

// This is an RPC function that is issued by the gateway to update the address port of the 
// loadsharing gateway the device is talking to. It returns the device id
func (s *SmartBulb) ChangeGateway(params *api.RegisterGatewayUserParams, reply *int) error {
	s.greplica.Set(api.RegisterGatewayUserParams{Address: params.Address, Port: params.Port})
	*reply = s.id.Get()
	return nil
}
