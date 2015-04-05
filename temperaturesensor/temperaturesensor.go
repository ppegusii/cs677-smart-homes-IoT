package main

import (
	"bufio"
	"fmt"
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"github.com/ppegusii/cs677-smart-homes-IoT/structs"
	"log"
	"net"
	"net/rpc"
	"os"
)

type TemperatureSensor struct {
	id          int
	gatewayIp   string
	gatewayPort string
	selfIp      string
	selfPort    string
	temperature structs.SyncState
	peers		map[int]string// To keep a track of all peers
}

func newTemperatureSensor(temperature api.State, gatewayIp string, gatewayPort string, selfIp string, selfPort string) *TemperatureSensor {
	return &TemperatureSensor{
		gatewayIp:   gatewayIp,
		gatewayPort: gatewayPort,
		selfIp:      selfIp,
		selfPort:    selfPort,
		temperature: *structs.NewSyncState(temperature),
		peers:       make(map[int]string),
	}
}

func (t *TemperatureSensor) start() {
	//RPC server
	var err error = rpc.Register(api.SensorInterface(t))
	if err != nil {
		log.Fatal("rpc.Register error: %s\n", err)
	}
	var listener net.Listener
	listener, err = net.Listen("tcp", t.selfIp+":"+t.selfPort)
	if err != nil {
		log.Fatal("net.Listen error: %s\n", err)
	}
	go rpc.Accept(listener)
	//register with gateway
	var client *rpc.Client
	client, err = rpc.Dial("tcp", t.gatewayIp+":"+t.gatewayPort)
	if err != nil {
		log.Fatal("dialing error: %+v", err)
	}
	err = client.Call("Gateway.Register", &api.RegisterParams{Type: api.Sensor, Name: api.Temperature, Address: t.selfIp, Port: t.selfPort}, &t.id)
	if err != nil {
		log.Fatal("calling error: %+v", err)
	}
	log.Printf("Device id: %d", t.id)
	t.getPeerTable()
	logCurrentTemp(t.temperature.GetState())
	//listen on stdin for temperature triggers
	t.getInput()
}

func (t *TemperatureSensor) getInput() {
	//http://stackoverflow.com/questions/20895552/how-to-read-input-from-console-line
	reader := bufio.NewReader(os.Stdin)
	var temp api.State
	for {
		fmt.Print("Enter 1 to increase the temperature , Enter 0 to decrease the temperature : \t")
		text, _ := reader.ReadString('\n')
		fmt.Println(text)

		// Based on the user input simulate the increase or decrease of temperature
		switch text {
		case "0\n":
			temp = t.temperature.GetState()
			t.temperature.SetState(temp - 1)
			logCurrentTemp(t.temperature.GetState())
			break
		case "1\n":
			temp = t.temperature.GetState()
			t.temperature.SetState(temp + 1)
			logCurrentTemp(t.temperature.GetState())
			break
		default:
			fmt.Println("Invalid Input, Enter either 1 or 0")
			break
		}
	}
}

func (t *TemperatureSensor) QueryState(params *int, reply *api.StateInfo) error {
	reply.DeviceId = t.id
	reply.State = t.temperature.GetState()
	return nil
}

func logCurrentTemp(t api.State) {
	log.Printf("Current temp: %d", t)
}

// This is an asynchronous call to fetch the PeerTable from the Gateway
func (t *TemperatureSensor) getPeerTable() {
	var client *rpc.Client
	var err error
	client, err = rpc.Dial("tcp", t.gatewayIp+":"+t.gatewayPort)
	if err != nil {
		log.Printf("dialing error: %+v", err)
	}
	replycall := client.Go("Gateway.SendPeerTable", t.id, &t.peers, nil)
	pt :=  <-replycall.Done
	if(pt != nil) {
		log.Println("Fetching PeerTable from the gateway")
	} else {
		log.Println("SendPeerTable RPC call return value: ",pt)
	}

	// Add the gateway to the peertable
	t.peers[api.GatewayID] = t.gatewayIp+":"+t.gatewayPort

	// Testing to check if the entire peertable has been received
	fmt.Println("Received the peer table from Gateway as below:")
	for k, v := range t.peers {
		fmt.Println(k, v)
	}
}

func (t *TemperatureSensor) UpdatePeerTable(params *api.PeerInfo, _ *struct{}) error {
	switch params.Token {
	case 0:
		//Add new peer
		t.peers[params.DeviceId] = params.Address
	case 1:
		//Delete the old peer that got disconnected from the system
		delete(t.peers,params.DeviceId)
	default:
		log.Println("Unexpected Token")
	}
	return nil
}