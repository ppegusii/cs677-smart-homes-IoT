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
	temperature structs.SyncFloat64
}

func newTemperatureSensor(temperature float64, gatewayIp string, gatewayPort string, selfIp string, selfPort string) *TemperatureSensor {
	return &TemperatureSensor{
		gatewayIp:   gatewayIp,
		gatewayPort: gatewayPort,
		selfIp:      selfIp,
		selfPort:    selfPort,
		temperature: *structs.NewSyncFloat64(temperature),
	}
}

func (t *TemperatureSensor) start() {
	//RPC server
	var err error = rpc.Register(api.TemperatureSensorInterface(t))
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
		log.Printf("dialing error: %v", err)
	}
	err = client.Call("Gateway.Register", &api.RegisterParams{api.Sensor, api.Temperature, t.selfIp, t.selfPort}, &t.id)
	if err != nil {
		log.Printf("calling error: %v", err)
	}
	log.Printf("Device id: %d", t.id)
	//listen on stdin for temperature triggers
	t.getInput()
}

func (t *TemperatureSensor) getInput() {
	//http://stackoverflow.com/questions/20895552/how-to-read-input-from-console-line
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter 1 to increase the temperature , Enter 0 to decrease the temperature : \t")
		text, _ := reader.ReadString('\n')
		fmt.Println(text)

		// Based on the user input simulate the increase or decrease of temperature
		switch text {
		case "0\n":
			t.temperature.Change(-1.0)
			break
		case "1\n":
			t.temperature.Change(1.0)
			break
		default:
			fmt.Println("Invalid Input, Enter either 1 or 0")
			break
		}
	}
}

func (t *TemperatureSensor) QueryState(params *int, reply *api.QueryTemperatureParams) error {
	reply.DeviceId = t.id
	reply.Temperature = t.temperature.Get()
	return nil
}