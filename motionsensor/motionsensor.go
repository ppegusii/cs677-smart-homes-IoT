package main

import (
	"bufio"
	"fmt"
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"log"
	"net"
	"net/rpc"
	"os"
)

type MotionSensor struct {
	gatewayIp   string
	gatewayPort string
	selfIp      string
	selfPort    string
	id          int
}

func newMotionSensor(gatewayIp string, gatewayPort string, selfIp string, selfPort string) *MotionSensor {
	return &MotionSensor{
		gatewayIp:   gatewayIp,
		gatewayPort: gatewayPort,
		selfIp:      selfIp,
		selfPort:    selfPort,
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
		log.Printf("dialing error: %v", err)
	}
	err = client.Call("Gateway.Register", &api.RegisterParams{api.Sensor, api.Motion, m.selfIp, m.selfPort}, &m.id)
	if err != nil {
		log.Printf("calling error: %v", err)
	}
	//listen on stdin for motion triggers
	m.getInput()
}

func (m *MotionSensor) getInput() {
	//http://stackoverflow.com/questions/20895552/how-to-read-input-from-console-line
	reader := bufio.NewReader(os.Stdin)
	var empty struct{}
	for {
		fmt.Print("Hit enter to trigger motion sensor")
		reader.ReadString('\n')
		var client *rpc.Client
		var err error
		client, err = rpc.Dial("tcp", m.gatewayIp+":"+m.gatewayPort)
		if err != nil {
			log.Printf("dialing error: %v", err)
		}
		client.Go("Gateway.ReportMotion", api.ReportMotionParams{m.id, api.MotionStart}, empty, nil)
	}
}

func (m *MotionSensor) QueryState(params *int, reply *api.QueryStateParams) error {
	//MotionSensor is stateless in that it does
	//not store the current motion state.
	//It will return no motion by default.
	reply.DeviceId = *params
	reply.State = api.MotionStop
	return nil
}
