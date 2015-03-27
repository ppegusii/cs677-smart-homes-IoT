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

func (m *MotionSensor) start() {
	//RPC server
	var err error = rpc.Register(api.MotionSensorInterface(m))
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
		log.Printf("dialing error: %+v", err)
	}
	err = client.Call("Gateway.Register", &api.RegisterParams{api.Sensor, api.Motion, m.selfIp, m.selfPort}, &m.id)
	if err != nil {
		log.Printf("calling error: %+v", err)
	}
	log.Printf("Device id: %d", m.id)
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
			break
		case "1\n":
			if m.state.GetState() == api.MotionStart {
				fmt.Println("No change")
				continue
			}
			m.state.SetState(api.MotionStart)
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
		}
		client.Go("Gateway.ReportMotion", api.ReportMotionParams{m.id, m.state.GetState()}, empty, nil)
	}
}

func (m *MotionSensor) QueryState(params *int, reply *api.QueryStateParams) error {
	reply.DeviceId = m.id
	reply.State = m.state.GetState()
	return nil
}
