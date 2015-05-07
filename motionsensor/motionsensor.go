// This file declares all the structs and interfaces needed by motion sensor
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
	"time"
)

// This struct contains all the attributes of the motion sensor and information needed for
// ordering for clock synchronization, peer table to keep a track of ip of the peers
// and reference to its middleware
type MotionSensor struct {
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
	greplica     *structs.SyncRegGatewayUserParam // This is the gateway replica assigned for load balancing
	rpcSync      api.RpcSyncInterface
}

// create and initialize a new motion sensor
func newMotionSensor(gatewayIp string, gatewayPort string, gatewayIp2 string, gatewayPort2 string, selfIp string, selfPort string, ordering api.Ordering) *MotionSensor {
	return &MotionSensor{
		id:           structs.NewSyncInt(api.UNREGISTERED),
		gatewayIp:    gatewayIp,
		gatewayPort:  gatewayPort,
		gatewayIp2:   gatewayIp2,
		gatewayPort2: gatewayPort2,
		ordering:     ordering,
		selfIp:       selfIp,
		selfPort:     selfPort,
		state:        *structs.NewSyncState(api.MotionStop),
		greplica:     structs.NewSyncRegGatewayUserParam(),
	}
}

func (m *MotionSensor) start() {
	//register with gateway
	var err error
	var regparam *api.RegisterParams = &api.RegisterParams{
		Address: m.selfIp,
		Name:    api.Motion,
		Port:    m.selfPort,
		Type:    api.Sensor,
	}
	var regresponse api.RegisterReturn

	// Dial to the first gateway
	err = util.RpcSync(m.gatewayIp, m.gatewayPort,
		"Gateway.Register", regparam, &regresponse, false)
	if err != nil {
		// Dial to the second gateway
		err = util.RpcSync(m.gatewayIp2, m.gatewayPort2,
			"Gateway.Register", regparam, &regresponse, false)
		if err != nil {
			log.Fatal("Could not register with a gateway.\n")
		}
	}

	m.id.Set(regresponse.DeviceId)
	m.greplica.Set(api.RegisterGatewayUserParams{Address: regresponse.Address, Port: regresponse.Port})
	replica := m.greplica.Get()
	log.Printf("Device id: %d %s %s", m.id.Get(), replica.Address, replica.Port)

	util.LogCurrentState(m.state.GetState())
	//initialize middleware
	//	m.orderMW = ordermw.GetOrderingMiddleware(m.ordering, m.id, m.selfIp, m.selfPort)
	//send acknowledgment of registration

	/*		var empty struct{}
			client, err = rpc.Dial("tcp", m.gatewayIp+":"+m.gatewayPort)
			if err != nil {
				log.Printf("dialing error: %+v", err)
				return
			}
			client.Go("Gateway.RegisterAck", m.id, &empty, nil)
	*/

	//start RPC server
	err = rpc.Register(api.SensorInterface(m))
	if err != nil {
		log.Fatal("rpc.Register error: %s\n", err)
	}
	var listener net.Listener
	listener, err = net.Listen("tcp", m.selfIp+":"+m.selfPort)
	if err != nil {
		log.Fatal("net.Listen error: %s\n", err)
	}
	rpc.Accept(listener)
	//listen on stdin for motion triggers
	//m.getInput()
}

//RPC stub to change state remotely.
//It is called by the test controller.
func (m *MotionSensor) ChangeState(params *api.StateInfo, reply *api.StateInfo) error {
	log.Printf("Received request to change state to: %s\n", util.StateToString(params.State))
	switch params.State {
	case api.MotionStop:
		if m.state.GetState() == api.MotionStop {
			log.Printf("No change\n")
			break
		}
		m.state.SetState(api.MotionStop)
		util.LogCurrentState(m.state.GetState())
		m.sendState()
		break
	case api.MotionStart:
		if m.state.GetState() == api.MotionStart {
			log.Printf("No change\n")
			break
		}
		m.state.SetState(api.MotionStart)
		util.LogCurrentState(m.state.GetState())
		m.sendState()
		break
	default:
		log.Printf("Invalid change state request")
		break
	}
	return nil
}

func (m *MotionSensor) getInput() {
	//http://stackoverflow.com/questions/20895552/how-to-read-input-from-console-line
	reader := bufio.NewReader(os.Stdin)
	//var empty struct{}
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
		/*
			var client *rpc.Client
			var err error
			client, err = rpc.Dial("tcp", m.gatewayIp+":"+m.gatewayPort)
			if err != nil {
				log.Printf("dialing error: %+v", err)
				continue
			}
			client.Go("Gateway.ReportMotion", api.StateInfo{DeviceId: m.id, State: m.state.GetState()}, &empty, nil)
		*/
		m.sendState()
	}
}

//This is an RPC function that is issued by the gateway to get the state of the motion sensor
func (m *MotionSensor) QueryState(params *int, reply *api.StateInfo) error {
	reply.DeviceId = m.id.Get()
	reply.DeviceName = api.Motion
	reply.State = m.state.GetState()
	// ***Commented out gateway will get state from reply***
	//go m.sendState()
	return nil
}

// The motion sensor is a push based device; sendState() is used to report state to the middleware
func (m *MotionSensor) sendState() {
	replica := m.greplica.Get()
	var stateInfo *api.StateInfo = &api.StateInfo{
		Clock:      int(time.Now().Unix()), //current timestamp for event ordering
		DeviceId:   m.id.Get(),
		DeviceName: api.Motion,
		State:      m.state.GetState(),
	}
	util.RpcSync(replica.Address, replica.Port,
		"Gateway.ReportMotion",
		stateInfo, &api.Empty{}, false)
}

/*
	var err error = m.orderMW.SendState(api.StateInfo{DeviceId: m.id.Get(), DeviceName: api.Motion, State: m.state.GetState()}, m.gatewayIp, m.gatewayPort)
	if err != nil {
		log.Printf("Error sending state: %+v", err)
	}
*/

// This is an RPC function that is issued by the gateway to update the address port of the
// loadsharing gateway the device is talking to. It returns the device id
func (m *MotionSensor) ChangeGateway(params *api.RegisterGatewayUserParams, reply *int) error {
	m.greplica.Set(api.RegisterGatewayUserParams{Address: params.Address, Port: params.Port})
	*reply = m.id.Get()
	return nil
}
