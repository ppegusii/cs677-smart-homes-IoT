// This file declares all the structs and interfaces needed by door sensor
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

// This struct contains all the attributes of the door sensor and information needed for
// ordering for clock synchronization, peer table to keep a track of ip of the peers and reference to its middleware
type DoorSensor struct {
	id           *structs.SyncInt
	gatewayIp    string
	gatewayPort  string
	gatewayIp2   string
	gatewayPort2 string
	ordering     api.Ordering
	orderMW      api.OrderingMiddlewareInterface
	selfIp       string
	selfPort     string
	state        structs.SyncState
	greplica 	 *structs.SyncRegGatewayUserParam // This is the gateway replica assigned for load balancing
}

// initialize a new doorsensor
func newDoorSensor(gatewayIp string, gatewayPort string, gatewayIp2 string, gatewayPort2 string, selfIp string, selfPort string, ordering api.Ordering) *DoorSensor {
	return &DoorSensor{
		id:			  structs.NewSyncInt(api.UNREGISTERED),
		gatewayIp:    gatewayIp,
		gatewayPort:  gatewayPort,
		gatewayIp2:   gatewayIp2,
		gatewayPort2: gatewayPort2,
		ordering:     ordering,
		selfIp:       selfIp,
		selfPort:     selfPort,
		state:        *structs.NewSyncState(api.Closed),
		greplica:	  structs.NewSyncRegGatewayUserParam(),
	}
}

func (d *DoorSensor) start() {
	//register with gateway
	var client *rpc.Client
	var err error
	var regresponse *api.RegisterReturn

	// Dial to the first gateway
	client, err = rpc.Dial("tcp", d.gatewayIp+":"+d.gatewayPort)
	if err != nil {
		log.Fatal("dialing error: %+v", err)
	}
	replycall1 := client.Go("Gateway.Register", &api.RegisterParams{Type: api.Sensor, Name: api.Motion, Address: d.selfIp, Port: d.selfPort}, &regresponse, nil)
	id1 := <-replycall1.Done

	// Dial to the second gateway
	client, err = rpc.Dial("tcp", d.gatewayIp2+":"+d.gatewayPort2)
	if err != nil {
		log.Fatal("dialing error: %+v", err)
	}
	replycall2 := client.Go("Gateway.Register", &api.RegisterParams{Type: api.Sensor, Name: api.Motion, Address: d.selfIp, Port: d.selfPort}, &regresponse, nil)
	id2 := <-replycall2.Done

	if (id1 != nil) || (id2 != nil) {
		log.Println("Registering with the gateway")
	} else {
		log.Println("Register RPC call return value: ", id1, id2)
	}

	/*	client, err = rpc.Dial("tcp", d.gatewayIp+":"+d.gatewayPort)
		if err != nil {
			log.Fatal("dialing error: %+v", err)
		}
		err = client.Call("Gateway.Register", &api.RegisterParams{Type: api.Sensor, Name: api.Door, Address: d.selfIp, Port: d.selfPort}, &d.id)
		if err != nil {
			log.Fatal("calling error: %+v", err)
		}
	*/
	d.id.Set(regresponse.DeviceId)
	d.greplica.Set(api.RegisterGatewayUserParams{Address: regresponse.Address, Port: regresponse.Port})
	replica := d.greplica.Get()
	log.Printf("Device id: %d %s %s", d.id.Get(), replica.Address, replica.Port)

	util.LogCurrentState(d.state.GetState())
	//initialize middleware
	/*	d.orderMW = ordermw.GetOrderingMiddleware(d.ordering, d.id, d.selfIp, d.selfPort)

		//send acknowledgment of registration
		var empty struct{}
		client, err = rpc.Dial("tcp", d.gatewayIp+":"+d.gatewayPort)
		if err != nil {
			log.Printf("dialing error: %+v", err)
			return
		}
		client.Go("Gateway.RegisterAck", d.id, &empty, nil)

	*/
	//start RPC server
	err = rpc.Register(api.SensorInterface(d))
	if err != nil {
		log.Fatal("rpc.Register error: %s\n", err)
	}
	var listener net.Listener
	listener, err = net.Listen("tcp", d.selfIp+":"+d.selfPort)
	if err != nil {
		log.Fatal("net.Listen error: %s\n", err)
	}
	rpc.Accept(listener)
	//listen on stdin for door triggers
	//d.getInput()
}

//RPC stub to change state remotely.
//It is called by the test controller.
func (d *DoorSensor) ChangeState(params *api.StateInfo, reply *api.StateInfo) error {
	switch params.State {
	case api.Open:
		if d.state.GetState() == api.Open {
			fmt.Println("No change")
			break
		}
		d.state.SetState(api.Open)
		util.LogCurrentState(d.state.GetState())
		d.sendState()
		break
	case api.Closed:
		if d.state.GetState() == api.Closed {
			fmt.Println("No change")
			break
		}
		d.state.SetState(api.Closed)
		util.LogCurrentState(d.state.GetState())
		d.sendState()
		break
	default:
		fmt.Println("Invalid change state request")
		break
	}
	reply = &api.StateInfo{
		DeviceId:   d.id.Get(),
		DeviceName: api.Door,
		State:      d.state.GetState(),
	}
	return nil
}

func (d *DoorSensor) getInput() {
	//http://stackoverflow.com/questions/20895552/how-to-read-input-from-console-line
	reader := bufio.NewReader(os.Stdin)
	//var empty struct{}
	for {
		fmt.Print("Enter (0/1) to signal (open/closed): ")
		input, _ := reader.ReadString('\n')
		switch input {
		case "0\n":
			if d.state.GetState() == api.Open {
				fmt.Println("No change")
				continue
			}
			d.state.SetState(api.Open)
			util.LogCurrentState(d.state.GetState())
			break
		case "1\n":
			if d.state.GetState() == api.Closed {
				fmt.Println("No change")
				continue
			}
			d.state.SetState(api.Closed)
			util.LogCurrentState(d.state.GetState())
			break
		default:
			fmt.Println("Invalid input")
			continue
		}
		/*
			var client *rpc.Client
			var err error
			client, err = rpc.Dial("tcp", d.gatewayIp+":"+d.gatewayPort)
			if err != nil {
				log.Printf("dialing error: %+v", err)
				continue
			}
			client.Go("Gateway.ReportDoorState", api.StateInfo{DeviceId: d.id, State: d.state.GetState()}, &empty, nil)
		*/
		//d.sendState()
	}
}

//This is an RPC function that is issued by the gateway to get the state of the door sensor
func (d *DoorSensor) QueryState(params *int, reply *api.StateInfo) error {
	reply.DeviceId = d.id.Get()
	reply.DeviceName = api.Door
	reply.State = d.state.GetState()
	//go d.sendState()
	return nil
}

// The Door sensor is a push based device and can be polled by the gateway.
// sendState() is used to report state to the gateway
func (d *DoorSensor) sendState() {

/*	var err error = d.orderMW.SendState(api.StateInfo{DeviceId: d.id, DeviceName: api.Door, State: d.state.GetState()}, d.gatewayIp, d.gatewayPort)
	if err != nil {
		log.Printf("Error sending state: %+v", err)
	}
	*/
}

// This is an RPC function that is issued by the gateway to update the address port of the 
// loadsharing gateway the device is talking to. It returns the device id
func (d *DoorSensor) ChangeGateway(params *api.RegisterGatewayUserParams, reply *int) error {
	d.greplica.Set(api.RegisterGatewayUserParams{Address: params.Address, Port: params.Port})
	*reply = d.id.Get()
	return nil
}