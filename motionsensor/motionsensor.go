package main

import (
	"bufio"
	"fmt"
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"github.com/ppegusii/cs677-smart-homes-IoT/ordermw"
	"github.com/ppegusii/cs677-smart-homes-IoT/structs"
	"github.com/ppegusii/cs677-smart-homes-IoT/util"
	"log"
	"net"
	"net/rpc"
	"os"
)

// This struct contains all the attributes of the motion sensor and information needed for
// ordering for clock synchronization, peer table to keep a track of ip of the peers
// and reference to its middleware
type MotionSensor struct {
	id          int
	gatewayIp   string
	gatewayPort string
	ordering    api.Ordering
	orderMW     api.OrderingMiddlewareInterface
	selfIp      string
	selfPort    string
	state       structs.SyncState
}

// create and initialize a new motion sensor
func newMotionSensor(gatewayIp string, gatewayPort string, selfIp string, selfPort string, ordering api.Ordering) *MotionSensor {
	return &MotionSensor{
		gatewayIp:   gatewayIp,
		gatewayPort: gatewayPort,
		ordering:    ordering,
		selfIp:      selfIp,
		selfPort:    selfPort,
		state:       *structs.NewSyncState(api.MotionStop),
	}
}

func (m *MotionSensor) start() {
	//register with gateway
	var client *rpc.Client
	var err error
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
	//initialize middleware
	m.orderMW = ordermw.GetOrderingMiddleware(m.ordering, m.id, m.selfIp, m.selfPort)

	//send acknowledgment of registration
	var empty struct{}
	client, err = rpc.Dial("tcp", m.gatewayIp+":"+m.gatewayPort)
	if err != nil {
		log.Printf("dialing error: %+v", err)
		return
	}
	client.Go("Gateway.RegisterAck", m.id, &empty, nil)

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
	reply.DeviceId = m.id
	reply.DeviceName = api.Motion
	reply.State = m.state.GetState()
	go m.sendState()
	return nil
}

// The motion sensor is a push based device; sendState() is used to report state to the middleware
func (m *MotionSensor) sendState() {
	var err error = m.orderMW.SendState(api.StateInfo{DeviceId: m.id, DeviceName: api.Motion, State: m.state.GetState()}, m.gatewayIp, m.gatewayPort)
	if err != nil {
		log.Printf("Error sending state: %+v", err)
	}
}
