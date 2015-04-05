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
}

func newMotionSensor(gatewayIp string, gatewayPort string, selfIp string, selfPort string) *MotionSensor {
	return &MotionSensor{
		gatewayIp:   gatewayIp,
		gatewayPort: gatewayPort,
		selfIp:      selfIp,
		selfPort:    selfPort,
		state:       *structs.NewSyncState(api.MotionStop),
	}
}

func (m *MotionSensor) getPeerTable() {
	var client *rpc.Client
	var err error
	var peers = make(map[int]string)
	client, err = rpc.Dial("tcp", m.gatewayIp+":"+m.gatewayPort)
	if err != nil {
		log.Printf("dialing error: %+v", err)
	}
	client.Go("Gateway.SendPeerTable", m.id, &peers, nil)
	// Testing to check if the entire peertable has been received
	fmt.Println("Received the peer information from Gateway as")
	fmt.Println("Address of device 2 is ", peers[2])
	for k, v := range peers {
		fmt.Println(k, v)
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
