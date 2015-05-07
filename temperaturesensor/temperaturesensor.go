// This file declares all the structs and interfaces needed by temperature sensor
package main

import (
	"bufio"
	"fmt"
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	//	"github.com/ppegusii/cs677-smart-homes-IoT/ordermw"
	"github.com/ppegusii/cs677-smart-homes-IoT/structs"
	"github.com/ppegusii/cs677-smart-homes-IoT/util"
	"log"
	"net"
	"net/rpc"
	"os"
)

// This struct contains all the attributes of the temperature sensor and information needed for
// ordering for clock synchronization, peer table to keep a track of ip of the peers
// and reference to its middleware
type TemperatureSensor struct {
	id           *structs.SyncInt
	gatewayIp    string
	gatewayPort  string
	gatewayIp2   string
	gatewayPort2 string
	ordering     api.Ordering
	orderMW      api.OrderingMiddlewareInterface
	selfIp       string
	selfPort     string
	temperature  structs.SyncState
	greplica     *structs.SyncRegGatewayUserParam // This is the gateway replica assigned for load balancing
}

// create and initialize a new temperature sensor
func newTemperatureSensor(temperature api.State, gatewayIp string, gatewayPort string, gatewayIp2 string, gatewayPort2 string, selfIp string, selfPort string, ordering api.Ordering) *TemperatureSensor {
	return &TemperatureSensor{
		id:           structs.NewSyncInt(api.UNREGISTERED),
		gatewayIp:    gatewayIp,
		gatewayPort:  gatewayPort,
		gatewayIp2:   gatewayIp2,
		gatewayPort2: gatewayPort2,
		ordering:     ordering,
		selfIp:       selfIp,
		selfPort:     selfPort,
		temperature:  *structs.NewSyncState(temperature),
		greplica:     structs.NewSyncRegGatewayUserParam(),
	}
}

func (t *TemperatureSensor) start() {
	//register with gateway
	var err error
	var regparam *api.RegisterParams = &api.RegisterParams{
		Address: t.selfIp,
		Name:    api.Temperature,
		Port:    t.selfPort,
		Type:    api.Sensor,
	}
	var regresponse api.RegisterReturn

	// Dial to the first gateway
	err = util.RpcSync(t.gatewayIp, t.gatewayPort,
		"Gateway.Register", regparam, &regresponse, false)
	if err != nil {
		// Dial to the second gateway
		err = util.RpcSync(t.gatewayIp2, t.gatewayPort2,
			"Gateway.Register", regparam, &regresponse, false)
		if err != nil {
			log.Fatal("Could not register with a gateway.\n")
		}
	}

	t.id.Set(regresponse.DeviceId)
	t.greplica.Set(api.RegisterGatewayUserParams{Address: regresponse.Address, Port: regresponse.Port})
	replica := t.greplica.Get()
	log.Printf("Device id: %d %s %s", t.id.Get(), replica.Address, replica.Port)
	logCurrentTemp(t.temperature.GetState())

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
	reply.DeviceId = t.id.Get()
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

// This is an RPC function that is issued by the gateway to update the address port of the
// loadsharing gateway the device is talking to. It returns the device id
func (t *TemperatureSensor) ChangeGateway(params *api.RegisterGatewayUserParams, reply *int) error {
	log.Printf("Changing gateway to: %+v\n", *params)
	t.greplica.Set(api.RegisterGatewayUserParams{Address: params.Address, Port: params.Port})
	*reply = t.id.Get()
	return nil
}
