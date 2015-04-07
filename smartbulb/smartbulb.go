package main

import (
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"github.com/ppegusii/cs677-smart-homes-IoT/structs"
	"github.com/ppegusii/cs677-smart-homes-IoT/util"
	"log"
	"net"
	"net/rpc"
	"fmt"
)

type SmartBulb struct {
	id          int
	gatewayIp   string
	gatewayPort string
	selfIp      string
	selfPort    string
	state       structs.SyncState
	peers		map[int]string// To keep a track of all peers
}

func newSmartBulb(gatewayIp string, gatewayPort string, selfIp string, selfPort string) *SmartBulb {
	return &SmartBulb{
		gatewayIp:   gatewayIp,
		gatewayPort: gatewayPort,
		selfIp:      selfIp,
		selfPort:    selfPort,
		state:       *structs.NewSyncState(api.Off),
		peers:       make(map[int]string),
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
	s.getPeerTable()
	util.LogCurrentState(s.state.GetState())
	rpc.Accept(listener)
}

func (s *SmartBulb) QueryState(params *int, reply *api.StateInfo) error {
	//this will not be called in practice
	reply.DeviceId = s.id
	reply.State = s.state.GetState()
	return nil
}

func (s *SmartBulb) ChangeState(params *api.StateInfo, reply *api.StateInfo) error {
	log.Printf("Received change state request with info: %+v", params)
	s.state.SetState(params.State)
	util.LogCurrentState(s.state.GetState())
	reply.DeviceId = s.id
	reply.State = params.State
	return nil
}

// This is an asynchronous call to fetch the PeerTable from the Gateway
func (s *SmartBulb) getPeerTable() {
	var client *rpc.Client
	var err error
	client, err = rpc.Dial("tcp", s.gatewayIp+":"+s.gatewayPort)
	if err != nil {
		log.Printf("dialing error: %+v", err)
	}
	replycall := client.Go("Gateway.SendPeerTable", s.id, &s.peers, nil)
	pt :=  <-replycall.Done
	if(pt != nil) {
		log.Println("Fetching PeerTable from the gateway")
	} else {
		log.Println("SendPeerTable RPC call return value: ",pt)
	}

	// Add the gateway to the peertable
	s.peers[api.GatewayID] = s.gatewayIp+":"+s.gatewayPort

	// Testing to check if the entire peertable has been received
	fmt.Println("Received the peer table from Gateway as below:")
	for k, v := range s.peers {
		fmt.Println(k, v)
	}
}

func (s *SmartBulb) UpdatePeerTable(params *api.PeerInfo, _ *struct{}) error {
	switch params.Token {
	case 0:
		//Add new peer
		s.peers[params.DeviceId] = params.Address
		log.Println("Received a new peer: DeviceID - ",params.DeviceId," Address - ", s.peers[params.DeviceId])
	case 1:
		//Delete the old peer that got disconnected from the system
		delete(s.peers,params.DeviceId)
	default:
		log.Println("Unexpected Token")
	}
	return nil
}