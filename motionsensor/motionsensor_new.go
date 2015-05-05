// This file declares all the structs and interfaces needed by motion sensor
package main

import (
//	"bufio"
//	"fmt"
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
//	"github.com/ppegusii/cs677-smart-homes-IoT/ordermw"
	"github.com/ppegusii/cs677-smart-homes-IoT/structs"
	"github.com/ppegusii/cs677-smart-homes-IoT/util"
	"log"
	"net"
	"net/rpc"
//	"os"
)

// This struct contains all the attributes of the motion sensor and information needed for
// ordering for clock synchronization, peer table to keep a track of ip of the peers
// and reference to its middleware
type MotionSensor struct {
	id            int
	gatewayIp     string
	gatewayPort   string
	gatewayIp2    string
	gatewayPort2  string
//	ordering      api.Ordering
//	orderMW       api.OrderingMiddlewareInterface
	selfIp        string
	selfPort      string
	state         structs.SyncState
	nodeinterface api.NodeInterface
}

// create and initialize a new motion sensor
func newMotionSensor(gatewayIp string, gatewayPort string, gatewayIp2 string, gatewayPort2 string, selfIp string, selfPort string, ordering api.Ordering, nodeinterface api.NodeInterface) *MotionSensor {
	return &MotionSensor{
		gatewayIp:     gatewayIp,
		gatewayPort:   gatewayPort,
		gatewayIp2:    gatewayIp2,
		gatewayPort2:  gatewayPort2,
		ordering:      ordering,
		selfIp:        selfIp,
		selfPort:      selfPort,
		state:         *structs.NewSyncState(api.MotionStop),
		nodeinterface: nodeinterface,
	}
}

func (m *MotionSensor) start() {
	//register with gateway
	//Send an async message to both the gateways

	//Sending to gateway replica 1
	util.RpcAsync(m.gatewayIp, m.gatewayPort, "Gateway.Register",
		&api.RegisterParams{Type: api.Sensor, Name: api.Motion, Address: m.selfIp, Port: m.selfPort}, 
		&m.id,
		this.RPCStart, //Amee: Check the afterfunc to handle 
		false)

	//Sending to gateway replica 2
	util.RpcAsync(m.gatewayIp2, m.gatewayPort2, "Gateway.Register",
		&api.RegisterParams{Type: api.Sensor, Name: api.Motion, Address: m.selfIp, Port: m.selfPort}, 
		&m.id,
		this.RPCStart, //Amee: Check the afterfunc to handle 
		false)
}

func (m *MotionSensor) RPCStart() {
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
}