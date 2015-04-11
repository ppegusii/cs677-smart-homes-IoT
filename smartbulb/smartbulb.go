package main

import (
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"github.com/ppegusii/cs677-smart-homes-IoT/ordermw"
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
	ordering    api.Ordering
	orderMW     api.OrderingMiddlewareInterface
	selfIp      string
	selfPort    string
	state       structs.SyncState
}

// create and initialize a new smart bulb
func newSmartBulb(gatewayIp string, gatewayPort string, selfIp string, selfPort string, ordering api.Ordering) *SmartBulb {
	return &SmartBulb{
		gatewayIp:   gatewayIp,
		gatewayPort: gatewayPort,
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
	client, err = rpc.Dial("tcp", s.gatewayIp+":"+s.gatewayPort)
	if err != nil {
		log.Fatal("dialing error: %+v", err)
	}
	err = client.Call("Gateway.Register", &api.RegisterParams{Type: api.Device, Name: api.Bulb, Address: s.selfIp, Port: s.selfPort}, &s.id)
	if err != nil {
		log.Fatal("calling error: %+v", err)
	}
	log.Printf("Device id: %d", s.id)
	//initialize middleware
	s.orderMW = ordermw.GetOrderingMiddleware(s.ordering, s.id, s.selfIp, s.selfPort)

	//send acknowledgment of registration
	var empty struct{}
	client, err = rpc.Dial("tcp", s.gatewayIp+":"+s.gatewayPort)
	if err != nil {
		log.Printf("dialing error: %+v", err)
		return
	}
	client.Go("Gateway.RegisterAck", s.id, &empty, nil)

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
	//this will not be called in practice
	reply.DeviceId = s.id
	reply.State = s.state.GetState()
	go s.sendState()
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
	go s.sendState()
	return nil
}

// sendState() is used to report state to the middleware
func (s *SmartBulb) sendState() {
	var err error = s.orderMW.SendState(api.StateInfo{DeviceId: s.id, DeviceName: api.Outlet, State: s.state.GetState()}, s.gatewayIp, s.gatewayPort)
	if err != nil {
		log.Printf("Error sending state: %+v", err)
	}
}
