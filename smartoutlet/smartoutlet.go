package main

import (
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"github.com/ppegusii/cs677-smart-homes-IoT/structs"
	"log"
	"net"
	"net/rpc"
)

type SmartOutlet struct {
	id          int
	gatewayIp   string
	gatewayPort string
	selfIp      string
	selfPort    string
	state       structs.SyncState
}

func newSmartOutlet(gatewayIp string, gatewayPort string, selfIp string, selfPort string) *SmartOutlet {
	return &SmartOutlet{
		gatewayIp:   gatewayIp,
		gatewayPort: gatewayPort,
		selfIp:      selfIp,
		selfPort:    selfPort,
		state:       *structs.NewSyncState(api.Off),
	}
}

func (s *SmartOutlet) start() {
	//register with gateway
	var client *rpc.Client
	var err error
	client, err = rpc.Dial("tcp", s.gatewayIp+":"+s.gatewayPort)
	if err != nil {
		log.Printf("dialing error: %v", err)
	}
	err = client.Call("Gateway.Register", &api.RegisterParams{api.Device, api.Outlet, s.selfIp, s.selfPort}, &s.id)
	if err != nil {
		log.Printf("calling error: %v", err)
	}
	log.Printf("Device id: %d", s.id)
	//RPC server
	err = rpc.Register(api.DeviceInterface(s))
	if err != nil {
		log.Fatal("rpc.Register error: %s\n", err)
	}
	var listener net.Listener
	listener, err = net.Listen("tcp", s.selfIp+":"+s.selfPort)
	if err != nil {
		log.Fatal("net.Listen error: %s\n", err)
	}
	rpc.Accept(listener)
}

func (s *SmartOutlet) QueryState(params *int, reply *api.QueryStateParams) error {
	//this will not be called in practice
	reply.DeviceId = s.id
	reply.State = s.state.GetState()
	return nil
}

func (s *SmartOutlet) ChangeState(params *api.ChangeStateParams, _ *struct{}) error {
	log.Printf("Received change state request with info: %v", params)
	s.state.SetState(params.State)
	return nil
}