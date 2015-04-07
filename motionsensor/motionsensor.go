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

type MotionSensor struct {
	id          int
	gatewayIp   string
	gatewayPort string
	selfIp      string
	selfPort    string
	state       structs.SyncState
	peers		map[int]string // To keep a track of all peers
}

func newMotionSensor(gatewayIp string, gatewayPort string, selfIp string, selfPort string) *MotionSensor {
	return &MotionSensor{
		gatewayIp:   gatewayIp,
		gatewayPort: gatewayPort,
		selfIp:      selfIp,
		selfPort:    selfPort,
		state:       *structs.NewSyncState(api.MotionStop),
		peers:       make(map[int]string),
	}
}

func (m *MotionSensor) start() {
	//RPC server
	var err error = rpc.Register(api.SensorInterface(m))
	if err != nil {
		log.Fatal("rpc.Register error: %s\n", err)
	}
	var listener net.Listener
	listener, err = net.Listen("tcp", m.selfIp+":"+m.selfPort)
	if err != nil {
		log.Fatal("net.Listen error: %s\n", err)
	}
	go rpc.Accept(listener)
	//register with gateway
	var client *rpc.Client
	client, err = rpc.Dial("tcp", m.gatewayIp+":"+m.gatewayPort)
	if err != nil {
		log.Fatal("dialing error: %+v", err)
	}
	err = client.Call("Gateway.Register", &api.RegisterParams{Type: api.Sensor, Name: api.Motion, Address: m.selfIp, Port: m.selfPort}, &m.id)
	if err != nil {
		log.Fatal("calling error: %+v", err)
	}
	log.Printf("Device id: %d", m.id)
	util.LogCurrentState(m.state.GetState())
	m.getPeerTable()
	//listen on stdin for motion triggers
	m.getInput()
}

func (m *MotionSensor) getInput() {
	//http://stackoverflow.com/questions/20895552/how-to-read-input-from-console-line
	reader := bufio.NewReader(os.Stdin)
	var empty struct{}
	for {
		fmt.Print("Enter (0/1) to signal (nomotion/motion): ")
		input, _ := reader.ReadString('\n')
		switch input {
		case "0\n":
			if m.state.GetState() == api.MotionStop {
				fmt.Println("No change")
				continue
			}
			m.state.SetState(api.MotionStop)
			util.LogCurrentState(m.state.GetState())
			break
		case "1\n":
			if m.state.GetState() == api.MotionStart {
				fmt.Println("No change")
				continue
			}
			m.state.SetState(api.MotionStart)
			util.LogCurrentState(m.state.GetState())
			break
		default:
			fmt.Println("Invalid input")
			continue
		}
		var client *rpc.Client
		var err error
		client, err = rpc.Dial("tcp", m.gatewayIp+":"+m.gatewayPort)
		if err != nil {
			log.Printf("dialing error: %+v", err)
			continue
		}
		client.Go("Gateway.ReportMotion", api.StateInfo{DeviceId: m.id, State: m.state.GetState()}, &empty, nil)
	}
}

func (m *MotionSensor) QueryState(params *int, reply *api.StateInfo) error {
	reply.DeviceId = m.id
	reply.State = m.state.GetState()
	return nil
}

// This is an asynchronous call to fetch the PeerTable from the Gateway
func (m *MotionSensor) getPeerTable() {
	var client *rpc.Client
	var err error
	client, err = rpc.Dial("tcp", m.gatewayIp+":"+m.gatewayPort)
	if err != nil {
		log.Printf("dialing error: %+v", err)
	}
	replycall := client.Go("Gateway.SendPeerTable", m.id, &m.peers, nil)
	pt :=  <-replycall.Done
	if(pt != nil) {
		log.Println("Fetching PeerTable from the gateway")
	} else {
		log.Println("SendPeerTable RPC call return value: ",pt)
	}

	// Add the gateway to the peertable
	m.peers[api.GatewayID] = m.gatewayIp+":"+m.gatewayPort

	// Testing to check if the entire peertable has been received
	fmt.Println("Received the peer table from Gateway as below:")
	for k, v := range m.peers {
		fmt.Println(k, v)
	}
}

func (m *MotionSensor) UpdatePeerTable(params *api.PeerInfo, _ *struct{}) error {
	switch params.Token {
	case 0:
		//Add new peer
		m.peers[params.DeviceId] = params.Address
		log.Println("Received a new peer: DeviceID - ",params.DeviceId," Address - ", m.peers[params.DeviceId])
	case 1:
		//Delete the old peer that got disconnected from the system
		delete(m.peers,params.DeviceId)
	case 2:
		//IAmAlive msg from gateway
	case 3:
		//Election message
		//If device id is less then do nothing else Negate the response from the device and
		// send a request to device with higher id than current device id.
		if(m.id > params.DeviceId) {
			//Call Deny RPC and set the electionleader flag to false
			//For this we need the device type to issue the call
		}
	case 4:
		//Leader is announced
//		m.leaderid = params.DeviceId
	default:
		log.Println("Unexpected Token")
	}
	return nil
}
