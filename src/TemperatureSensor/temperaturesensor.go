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

type TemperatureSensor struct {
	temperature float64
	gatewayIp   string
	gatewayPort string
	selfIp      string
	selfPort    string
	id          int
}

func newTemperatureSensor(temperature float64, gatewayIp string, gatewayPort string, selfIp string, selfPort string) *TemperatureSensor {
	return &TemperatureSensor{
		temperature: temperature,
		gatewayIp:   gatewayIp,
		gatewayPort: gatewayPort,
		selfIp:      selfIp,
		selfPort:    selfPort,
	}
}

func (m *TemperatureSensor) start() {
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
	err = client.Call("Gateway.Register", &api.RegisterParams{api.Sensor, api.Temperature, m.selfIp, m.selfPort}, &m.id)
	if err != nil {
		log.Printf("calling error: %v", err)
	}
	log.Printf("Device id: %d", m.id)
	//listen on stdin for temperature triggers
	m.getInput()
}

func (m *TemperatureSensor) getInput() {
	//http://stackoverflow.com/questions/20895552/how-to-read-input-from-console-line
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter 1 to increase the temperature , Enter 0 to decrease the temperature : \t")
		text, _ := reader.ReadString('\n')
		fmt.Println(text)

		// Based on the user input simulate the increase or decrease of temperature of the 
		switch text {
			case "0" : m.temperature -=1
			case "1" : m.temperature +=1
			default : fmt.Println("Invalid Input, Enter either 1 or 0")
		}
	}
}

func (m *TemperatureSensor) QueryState(params *int, reply *api.QueryTemperatureParams) error {
	//MotionSensor is stateless in that it does
	//not store the current motion state.
	//It will return no motion by default.
	reply.DeviceId = *params
	reply.Temperature = m.temperature
	return nil
}
