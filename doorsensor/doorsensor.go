package main

import (
	"bufio"
	"fmt"
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"github.com/ppegusii/cs677-smart-homes-IoT/structs"
	"github.com/ppegusii/cs677-smart-homes-IoT/util"
	"log"
	"net"
	"net/rpc"
	"os"
)

type DoorSensor struct {
	id          int
	gatewayIp   string
	gatewayPort string
	selfIp      string
	selfPort    string
	state       structs.SyncState
	peers		map[int]string// To keep a track of all peers
}

func newDoorSensor(gatewayIp string, gatewayPort string, selfIp string, selfPort string) *DoorSensor {
	return &DoorSensor{
		gatewayIp:   gatewayIp,
		gatewayPort: gatewayPort,
		selfIp:      selfIp,
		selfPort:    selfPort,
		state:       *structs.NewSyncState(api.Closed),
		peers:       make(map[int]string),
	}
}

func (d *DoorSensor) start() {
	//RPC server
	var err error = rpc.Register(api.SensorInterface(d))
	if err != nil {
		log.Fatal("rpc.Register error: %s\n", err)
	}
	var listener net.Listener
	listener, err = net.Listen("tcp", d.selfIp+":"+d.selfPort)
	if err != nil {
		log.Fatal("net.Listen error: %s\n", err)
	}
	go rpc.Accept(listener)
	//register with gateway
	var client *rpc.Client
	client, err = rpc.Dial("tcp", d.gatewayIp+":"+d.gatewayPort)
	if err != nil {
		log.Fatal("dialing error: %+v", err)
	}
	err = client.Call("Gateway.Register", &api.RegisterParams{Type: api.Sensor, Name: api.Door, Address: d.selfIp, Port: d.selfPort}, &d.id)
	if err != nil {
		log.Fatal("calling error: %+v", err)
	}
	log.Printf("Device id: %d", d.id)
	d.getPeerTable()
	util.LogCurrentState(d.state.GetState())
	//listen on stdin for door triggers
	d.getInput()
}

func (d *DoorSensor) getInput() {
	//http://stackoverflow.com/questions/20895552/how-to-read-input-from-console-line
	reader := bufio.NewReader(os.Stdin)
	var empty struct{}
	for {
		fmt.Print("Enter (0/1) to signal (open/closed): ")
		input, _ := reader.ReadString('\n')
		switch input {
		case "0\n":
			if d.state.GetState() == api.Open {
				fmt.Println("No change")
				continue
			}
			d.state.SetState(api.Open)
			util.LogCurrentState(d.state.GetState())
			break
		case "1\n":
			if d.state.GetState() == api.Closed {
				fmt.Println("No change")
				continue
			}
			d.state.SetState(api.Closed)
			util.LogCurrentState(d.state.GetState())
			break
		default:
			fmt.Println("Invalid input")
			continue
		}
		var client *rpc.Client
		var err error
		client, err = rpc.Dial("tcp", d.gatewayIp+":"+d.gatewayPort)
		if err != nil {
			log.Printf("dialing error: %+v", err)
			continue
		}
		client.Go("Gateway.ReportDoorState", api.StateInfo{DeviceId: d.id, State: d.state.GetState()}, &empty, nil)
	}
}

func (d *DoorSensor) QueryState(params *int, reply *api.StateInfo) error {
	reply.DeviceId = d.id
	reply.State = d.state.GetState()
	return nil
}

// This is an asynchronous call to fetch the PeerTable from the Gateway
func (d *DoorSensor) getPeerTable() {
	var client *rpc.Client
	var err error
	client, err = rpc.Dial("tcp", d.gatewayIp+":"+d.gatewayPort)
	if err != nil {
		log.Printf("dialing error: %+v", err)
	}
	replycall := client.Go("Gateway.SendPeerTable", d.id, &d.peers, nil)
	pt :=  <-replycall.Done
	if(pt != nil) {
		log.Println("Fetching PeerTable from the gateway")
	} else {
		log.Println("SendPeerTable RPC call return value: ",pt)
	}

	// Add the gateway to the peertable
	d.peers[api.GatewayID] = d.gatewayIp+":"+d.gatewayPort

	// Testing to check if the entire peertable has been received
	fmt.Println("Received the peer table from Gateway as below:")
	for k, v := range d.peers {
		fmt.Println(k, v)
	}
}

func (d *DoorSensor) UpdatePeerTable(params *api.PeerInfo, _ *struct{}) error {
	switch params.Token {
	case 0:
		//Add new peer
		d.peers[params.DeviceId] = params.Address
	case 1:
		//Delete the old peer that got disconnected from the system
		delete(d.peers,params.DeviceId)
	default:
		log.Println("Unexpected Token")
	}
	return nil
}