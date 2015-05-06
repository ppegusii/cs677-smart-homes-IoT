// This file declares all the structs and interfaces needed by temperature sensor
package main

import (
	"bufio"
	"fmt"
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
//	"github.com/ppegusii/cs677-smart-homes-IoT/ordermw"
	"github.com/ppegusii/cs677-smart-homes-IoT/structs"
	"log"
	"net"
	"net/rpc"
	"os"
)

// This struct contains all the attributes of the temperature sensor and information needed for
// ordering for clock synchronization, peer table to keep a track of ip of the peers
// and reference to its middleware
type TemperatureSensor struct {
	id           int
	gatewayIp    string
	gatewayPort  string
	gatewayIp2   string
	gatewayPort2 string
	ordering     api.Ordering
	orderMW      api.OrderingMiddlewareInterface
	selfIp       string
	selfPort     string
	temperature  structs.SyncState
	gRPCIp 		 string
	gRPCPort	 string
}

// create and initialize a new temperature sensor
func newTemperatureSensor(temperature api.State, gatewayIp string, gatewayPort string, gatewayIp2 string, gatewayPort2 string, selfIp string, selfPort string, ordering api.Ordering) *TemperatureSensor {
	return &TemperatureSensor{
		gatewayIp:    gatewayIp,
		gatewayPort:  gatewayPort,
		gatewayIp2:   gatewayIp2,
		gatewayPort2: gatewayPort2,
		ordering:     ordering,
		selfIp:       selfIp,
		selfPort:     selfPort,
		temperature:  *structs.NewSyncState(temperature),
	}
}

func (t *TemperatureSensor) start() {
	//register with gateway
	var client *rpc.Client
	var err error
	var regresponse *api.RegisterReturn

	// Dial to the first gateway
	client, err = rpc.Dial("tcp", t.gatewayIp+":"+t.gatewayPort)
	if err != nil {
		log.Fatal("dialing error: %+v", err)
	}
	replycall1 := client.Go("Gateway.Register", &api.RegisterParams{Type: api.Sensor, Name: api.Motion, Address: t.selfIp, Port: t.selfPort}, &regresponse, nil)
	id1 := <-replycall1.Done

	// Dial to the second gateway
	client, err = rpc.Dial("tcp", t.gatewayIp2+":"+t.gatewayPort2)
	if err != nil {
		log.Fatal("dialing error: %+v", err)
	}
	replycall2 := client.Go("Gateway.Register", &api.RegisterParams{Type: api.Sensor, Name: api.Motion, Address: t.selfIp, Port: t.selfPort}, &regresponse, nil)
	id2 := <-replycall2.Done

	if (id1 != nil) || (id2 != nil) {
		log.Println("Registering with the gateway")
	} else {
		log.Println("Register RPC call return value: ", id1, id2)
	}

	t.id = regresponse.DeviceId
	t.gRPCIp = regresponse.Address
	t.gRPCPort = regresponse.Port
	log.Printf("Device id: %d %s %s", t.id, t.gRPCIp, t.gRPCPort)

	/*
		client, err = rpc.Dial("tcp", t.gatewayIp+":"+t.gatewayPort)
		if err != nil {
			log.Fatal("dialing error: %+v", err)
		}
		err = client.Call("Gateway.Register", &api.RegisterParams{Type: api.Sensor, Name: api.Temperature, Address: t.selfIp, Port: t.selfPort}, &t.id)
		if err != nil {
			log.Fatal("calling error: %+v", err)
		}
	*/
//	log.Printf("Device id: %d", t.id)
	logCurrentTemp(t.temperature.GetState())

	//Amee: Remove the middleware stuff

	//initialize middleware
/*	t.orderMW = ordermw.GetOrderingMiddleware(t.ordering, t.id, t.selfIp, t.selfPort)

		//send acknowledgment of registration
		var empty struct{}
		client, err = rpc.Dial("tcp", t.gatewayIp+":"+t.gatewayPort)
		if err != nil {
			log.Printf("dialing error: %+v", err)
			return
		}
		client.Go("Gateway.RegisterAck", t.id, &empty, nil)
	*/
	//Amee: Add nodeinterface update api.SensorTnterface to nodeinterface.Interface()

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
	rpc.Accept(listener)
	//listen on stdin for temperature triggers
	//t.getInput()
}

//RPC stub to change state remotely.
//It is called by the test controller.
func (t *TemperatureSensor) ChangeState(params *api.StateInfo, reply *api.StateInfo) error {
	t.temperature.SetState(params.State)
	return nil
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

//This is an RPC function that is issued by the gateway to get the state of the Temperature sensor
func (t *TemperatureSensor) QueryState(params *int, reply *api.StateInfo) error {
	reply.DeviceId = t.id
	reply.DeviceName = api.Temperature
	reply.State = t.temperature.GetState()
	//go t.sendState()
	return nil
}

// sendState() is used to report state to the middleware
/*
func (t *TemperatureSensor) sendState() {
	var err error = t.orderMW.SendState(api.StateInfo{DeviceId: t.id, DeviceName: api.Temperature, State: t.temperature.GetState()}, t.gatewayIp, t.gatewayPort)
	if err != nil {
		log.Printf("Error sending state: %+v", err)
	}
}
*/

//Print current temperature to the console
func logCurrentTemp(t api.State) {
	log.Printf("Current temp: %d", t)
}
