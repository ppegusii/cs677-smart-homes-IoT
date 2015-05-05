//Common node interface file
//This file contains all the interfaces common to all devices and sensors

package node

import (
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"github.com/ppegusii/cs677-smart-homes-IoT/structs"
	"github.com/ppegusii/cs677-smart-homes-IoT/util"
	"log"
//	"sync"
//	"time"
)

// This struct contains all the attributes needed to smoothly perform the operations 
// on devices and sensors irrespective of the type of device (referred to as node)
type Node struct {
	id           int
//	gatewayIp    string //Not sure if I will need these
//	gatewayPort  string //Not sure if I will need these
//	gatewayIp2   string //Not sure if I will need these
//	gatewayPort2 string //Not sure if I will need these
	state 		 structs.SyncState
}

func NewNode(ip, port string, replicas []api.RegisterGatewayUserParams) api.GatewayLeaderInterface {
	return &NewNode{
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

// This is an RPC function that is issued by the gateway to get the state of the motion sensor
func (n *Node) QueryState(params *int, reply *api.StateInfo) error {
	reply.DeviceId = n.id
	reply.DeviceName = api.Motion //Amee Todo: Change to corresponding state
	reply.State = n.state.GetState()
	return nil
}

// Amee TODO: Add cases to handle all device states 
/* The different states it should handle are:
	InvalidState State = iota //0
	Closed       State = iota //1
	MotionStart  State = iota //2
	MotionStop   State = iota //3
	Off          State = iota //4
	On           State = iota //5
	Open         State = iota //6
*/
//RPC stub to change state remotely.
//It is called by the test controller.
func (n *Node) ChangeState(params *api.StateInfo, reply *api.StateInfo) error {
	log.Printf("Received request to change state to: %s\n", util.StateToString(params.State))
	switch params.State {
	case api.MotionStop:
		if n.state.GetState() == api.MotionStop {
			log.Printf("No change\n")
			break
		}
		n.state.SetState(api.MotionStop)
		util.LogCurrentState(n.state.GetState())
		m.sendState()
		break
	case api.MotionStart:
		if n.state.GetState() == api.MotionStart {
			log.Printf("No change\n")
			break
		}
		n.state.SetState(api.MotionStart)
		util.LogCurrentState(n.state.GetState())
		n.sendState()
		break
	default:
		log.Printf("Invalid change state request")
		break
	}
	return nil
}