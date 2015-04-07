package main

import (
	"bufio"
	"fmt"
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"github.com/ppegusii/cs677-smart-homes-IoT/ordermw"
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
	ordering    api.Ordering
	orderMW     api.OrderingMiddlewareInterface
	selfIp      string
	selfPort    string
	temperature structs.SyncState
}

func newTemperatureSensor(temperature api.State, gatewayIp string, gatewayPort string, selfIp string, selfPort string, ordering api.Ordering) *TemperatureSensor {
	return &TemperatureSensor{
		gatewayIp:   gatewayIp,
		gatewayPort: gatewayPort,
		ordering:    ordering,
		selfIp:      selfIp,
		selfPort:    selfPort,
		temperature: *structs.NewSyncState(temperature),
	}
}

func (t *TemperatureSensor) start() {
	//register with gateway
	var client *rpc.Client
	var err error
	client, err = rpc.Dial("tcp", t.gatewayIp+":"+t.gatewayPort)
	if err != nil {
		log.Fatal("dialing error: %+v", err)
	}
	err = client.Call("Gateway.Register", &api.RegisterParams{Type: api.Sensor, Name: api.Temperature, Address: t.selfIp, Port: t.selfPort}, &t.id)
	if err != nil {
		log.Fatal("calling error: %+v", err)
	}
	log.Printf("Device id: %d", t.id)
	logCurrentTemp(t.temperature.GetState())
	//initialize middleware
	t.orderMW = ordermw.GetOrderingMiddleware(t.ordering, t.id, t.selfIp, t.selfPort)
	//start RPC server
	err = rpc.Register(api.SensorInterface(t))
	if err != nil {
		log.Fatal("rpc.Register error: %s\n", err)
	}
	var listener net.Listener
	listener, err = net.Listen("tcp", t.selfIp+":"+t.selfPort)
	if err != nil {
		log.Fatal("net.Listen error: %s\n", err)
	}
	go rpc.Accept(listener)
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
	/*
		reply.DeviceId = t.id
		reply.State = t.temperature.GetState()
	*/
	go t.sendState()
	return nil
}

func (t *TemperatureSensor) sendState() {
	var err error = t.orderMW.SendState(api.StateInfo{DeviceId: t.id, DeviceName: api.Temperature, State: t.temperature.GetState()}, t.gatewayIp, t.gatewayPort)
	if err != nil {
		log.Printf("Error sending state: %+v", err)
	}
}

func logCurrentTemp(t api.State) {
	log.Printf("Current temp: %d", t)
}
