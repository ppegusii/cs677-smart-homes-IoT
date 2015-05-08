// This file wraps the gateway. It provides gateway leader election,
// crash fault detection, and load balancing between gateway
// replicas. It also waits before servicing a request during elections.

package gatewayleader

import (
	"errors"
	"fmt"
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"github.com/ppegusii/cs677-smart-homes-IoT/structs"
	"github.com/ppegusii/cs677-smart-homes-IoT/util"
	"log"
	"sync"
	"time"
)

const (
	iWonWait           time.Duration = 5 * time.Second // duration to wait for IWon replies
	okWait             time.Duration = 2 * time.Second // duration to wait for OKs
	aliveReplyWait     time.Duration = 2 * time.Second // duration to wait for are you alive replies
	aliveRequestWait   time.Duration = 5 * time.Second // duration to wait before sending next are you alive probe
	nonleaderAliveWait time.Duration = 6 * time.Second // duration leader waits before consider a nonleader dead
)

type GatewayLeader struct {
	aliveReplyTimer   *structs.SyncTimer
	aliveRequestTimer *structs.SyncTimer
	api.GatewayInterface
	electionCheckLock  sync.Mutex
	electionInProgress bool
	ipPort             api.RegisterGatewayUserParams
	iWonReplies        chan int // Used to wait for all IWon replies
	iWonTimer          *structs.SyncTimer
	leader             *structs.SyncRegGatewayUserParam
	okTimer            *structs.SyncTimer
	replicas           *syncMapStringReplica
	sync.RWMutex       // Used to wait before sending application messages during an election.
}

func NewGatewayLeader(ip, port string, replicas []api.RegisterGatewayUserParams) api.GatewayLeaderInterface {
	var g GatewayLeader = GatewayLeader{
		leader: structs.NewSyncRegGatewayUserParam(),
		ipPort: api.RegisterGatewayUserParams{
			Address: ip,
			Port:    port,
		},
	}
	// create timers
	g.iWonTimer = structs.NewSyncTimer(iWonWait, g.startElection)
	g.okTimer = structs.NewSyncTimer(okWait, g.sendIWons)
	g.aliveReplyTimer = structs.NewSyncTimer(aliveReplyWait, g.handleAliveTimeout)
	g.aliveRequestTimer = structs.NewSyncTimer(aliveRequestWait, g.pollLeader)
	// create data structure for replica info
	g.replicas = newSyncMapStringReplica(replicas, &g)
	return &g
}

// Intercept all RPC calls from the gateway application.
func (this *GatewayLeader) RpcSync(ip, port, rpcName string, args interface{}, reply interface{}, isErrFatal bool) error {
	var err error = util.RpcSync(ip, port, rpcName, args, reply, isErrFatal)
	return err
}

// Set the gateway pointer.
func (this *GatewayLeader) SetGateway(g api.GatewayInterface) {
	this.GatewayInterface = g
}

// Start routines necessary for leader.
func (this *GatewayLeader) StartLeader() {
	this.Election(this.ipPort, &api.Empty{})
}

// Intercept registration requests. Service only if election is not in progress and leader.
func (this *GatewayLeader) Register(params *api.RegisterParams, reply *api.RegisterReturn) error {
	this.RLock()
	// Only service request if leader.
	if !this.isLeader() {
		var err error = errors.New(fmt.Sprintf("Inactive gateway replica: %+v\n", this.ipPort))
		this.RUnlock()
		return err
	}
	var start int64 = util.GetTime()
	var err error = this.GatewayInterface.Register(params, reply)
	if err != nil {
		this.RUnlock()
		return err
	}
	// Assign the node to a replica.
	var assigned *api.RegisterGatewayUserParams = this.replicas.loadBalance(*params, reply.DeviceId)
	log.Printf("Node %+v assigned to replica: %+v\n", params, assigned)
	reply.Address = assigned.Address
	reply.Port = assigned.Port
	// Release consistency.
	this.sendDataToAllReplicas(start)
	this.RUnlock()
	return err
}

// Multicast local data to all replicas.
func (this *GatewayLeader) sendDataToAllReplicas(start int64) {
	// Get local data.
	var data api.ConsistencyData
	this.GatewayInterface.PullData(start, &data)
	// Add sensor assignments.
	data.AssignedNodes = *(this.replicas.getAssignments())
	// Send data.
	var thisId string = util.RegisterGatewayUserParamsToString(this.ipPort)
	for _, ipPort := range this.replicas.getIpPorts() {
		var id string = util.RegisterGatewayUserParamsToString(ipPort)
		if id != thisId {
			go util.RpcSync(ipPort.Address, ipPort.Port,
				"Gateway.PushData", &data, &api.Empty{}, false)
		}
	}
}

// Convenience method for determining if this replica is the leader.
func (this *GatewayLeader) isLeader() bool {
	// http://golang.org/ref/spec#Comparison_operators
	// The equality operators == and != apply to operands that are comparable.
	// Struct values are comparable if all their fields are comparable.
	// Two struct values are equal if their corresponding non-blank fields are equal.
	return this.leader.Get() == this.ipPort
}

// Intercept request. Block service if election is in progress.
func (this *GatewayLeader) RegisterUser(params *api.RegisterGatewayUserParams, empty *struct{}) error {
	this.RLock()
	var err error = this.GatewayInterface.RegisterUser(params, empty)
	this.RUnlock()
	return err
}

// Intercept request. Block service if election is in progress.
func (this *GatewayLeader) ReportDoorState(params *api.StateInfo, empty *struct{}) error {
	this.RLock()
	var err error = this.GatewayInterface.ReportDoorState(params, empty)
	this.RUnlock()
	return err
}

// Intercept request. Block service if election is in progress.
func (this *GatewayLeader) ReportMotion(params *api.StateInfo, empty *struct{}) error {
	this.RLock()
	var err error = this.GatewayInterface.ReportMotion(params, empty)
	this.RUnlock()
	return err
}

// Data requested by another replica
func (this *GatewayLeader) PullData(clock int64, data *api.ConsistencyData) error {
	var err error = this.GatewayInterface.PullData(clock, data)
	log.Printf("Sending data: %+v\n", data)
	// Add sensor assignments.
	data.AssignedNodes = *(this.replicas.getAssignments())
	return err
}

// Data sent from another replica
func (this *GatewayLeader) PushData(data *api.ConsistencyData, e *api.Empty) error {
	log.Printf("Received data: %+v\n", data)
	//TODO if not leader update node assignments
	if !this.isLeader() {
		this.replicas.setAssignments(&(data.AssignedNodes))
	}
	return this.GatewayInterface.PushData(data, e)
}

// Receive election msg from self or another replica.
// Only entry into election process.
func (this *GatewayLeader) Election(replica api.RegisterGatewayUserParams, ok *api.Empty) error {
	log.Printf("Received election msg from: %+v\n", replica)
	// Ensure no parallel elections.
	this.electionCheckLock.Lock()
	if !this.electionInProgress {
		this.electionInProgress = true
		this.Lock()
	} else {
		defer this.electionCheckLock.Unlock()
		return nil
	}
	this.electionCheckLock.Unlock()

	go this.startElection()
	return nil
}

// Receive an IWon message from another replica.
func (this *GatewayLeader) IWon(replica api.RegisterGatewayUserParams, reply *api.RegisterGatewayUserParams) error {
	log.Printf("Received iwon msg from: %+v\n", replica)
	*reply = this.ipPort
	this.iWonTimer.Stop()
	this.leader.Set(replica)

	// End election.
	this.electionCheckLock.Lock()
	if this.electionInProgress {
		this.electionInProgress = false
		this.Unlock()
	}
	this.electionCheckLock.Unlock()
	log.Printf("Elected other replica: %+v\n", replica)

	// Start polling the leader replica for life.
	go this.pollLeader()
	return nil
}

// Poll the leader
func (this *GatewayLeader) pollLeader() {
	log.Println("Polling leader")
	// start timer
	// timer will call start election if duration elapses
	this.aliveReplyTimer.Reset()
	var leader api.RegisterGatewayUserParams = this.leader.Get()
	// async RPC to poll leader
	util.RpcAsync(leader.Address, leader.Port, "Gateway.Alive",
		this.ipPort, &api.Empty{}, this.handleAlive, false)
}

// Handle alive reply timeout by starting an election if its not already in progress.
func (this *GatewayLeader) handleAliveTimeout() {
	log.Println("Handling alive timeout")
	// Do nothing if election in progress.
	this.electionCheckLock.Lock()
	if this.electionInProgress {
		defer this.electionCheckLock.Unlock()
		return
	}
	this.electionCheckLock.Unlock()
	this.Election(this.ipPort, &api.Empty{})
}

// Handles alive replies from active replicas by restarting a timer to request an alive response from the leader.
func (this *GatewayLeader) handleAlive(_ interface{}, _ interface{}, err error) {
	log.Println("Handling alive reply")
	if err != nil {
		return
	}
	// handle RPC reply by stopping timer and recalling pollLeader
	this.aliveReplyTimer.Stop()
	// Do nothing if election in progress.
	this.electionCheckLock.Lock()
	if this.electionInProgress {
		defer this.electionCheckLock.Unlock()
		return
	}
	this.electionCheckLock.Unlock()
	// Start timer to poll leader replica.
	this.aliveRequestTimer.Reset()
}

// Start an election. Dead replicas will be detected during the election if this replica is elected.
// This happens because any alive replica with a greater ID would be elected and all replicas with
// lower IDs must acknowledge the IWon message.
func (this *GatewayLeader) startElection() {
	log.Println("Starting election")
	this.replicas.setAllReplicasDead()
	var thisId string = util.RegisterGatewayUserParamsToString(this.ipPort)
	// Start a timer that when duration elapses sends IWon.
	this.okTimer.Reset()
	// Send election notice to each replica with a higher id (async RPC).
	for _, ipPort := range this.replicas.getIpPorts() {
		var id string = util.RegisterGatewayUserParamsToString(ipPort)
		if id > thisId {
			util.RpcAsync(ipPort.Address, ipPort.Port, "Gateway.Election",
				this.ipPort, &api.Empty{}, this.handleOKs, false)
		}
	}
}

// Handles OK replies in response to election messages by stopping the OK timers.
func (this *GatewayLeader) handleOKs(_ interface{}, _ interface{}, err error) {
	log.Println("Handling OK")
	if err != nil {
		return
	}
	this.okTimer.Stop()
	// Start a timer that when duration elapses restarts election.
	this.iWonTimer.Reset()
}

// Send IWons to replicas with lower IDs and end election.
func (this *GatewayLeader) sendIWons() {
	log.Println("Sending IWons")
	var thisId string = util.RegisterGatewayUserParamsToString(this.ipPort)
	// Declare self leader
	this.leader.Set(this.ipPort)
	var ipPorts []api.RegisterGatewayUserParams = this.replicas.getIpPorts()
	if len(ipPorts) > 0 {
		// Create a buffered channel for pausing until all IWon replies received
		this.iWonReplies = make(chan int, len(ipPorts))
		// Send IWon messages to all replicas with lower id
		for _, ipPort := range ipPorts {
			var id string = util.RegisterGatewayUserParamsToString(ipPort)
			if id < thisId {
				util.RpcAsync(ipPort.Address, ipPort.Port,
					"Gateway.IWon", this.ipPort, &api.RegisterGatewayUserParams{},
					this.handleIWonReplies, false)
			} else {
				this.iWonReplies <- 1
			}
		}
		// Wait for all IWon replies
		for i := 0; i < len(ipPorts); i++ {
			<-this.iWonReplies
		}
	}
	// Rebalance load
	this.rebalanceLoad()
	// End election.
	this.electionCheckLock.Lock()
	if this.electionInProgress {
		this.electionInProgress = false
		this.Unlock()
	}
	this.electionCheckLock.Unlock()

	log.Printf("Elected self: %+v\n", this.ipPort)
}

// Handle IWon replies by declaring the responding replica alive.
func (this *GatewayLeader) handleIWonReplies(_, reply interface{}, err error) {
	this.iWonReplies <- 1
	if err != nil {
		return
	}
	var key string = util.RegisterGatewayUserParamsToString(
		*(reply.(*api.RegisterGatewayUserParams))) // last bit is a type assertion
	this.replicas.setAlive(key, true)
}

// Receive alive probes. Reset timer for the requesting replica to detect replica crash
// faults if leader.
func (this *GatewayLeader) Alive(replica api.RegisterGatewayUserParams, yes *api.Empty) error {
	log.Println("Received alive probe")
	// TODO Determine if other replicas are active.
	var key string = util.RegisterGatewayUserParamsToString(replica)
	this.replicas.ResetTimer(key)
	return nil
}

// Get a closure that sets a replica to dead. Used as callback functions replica alive request
// timers for detecting dead replicas if leader.
func (this *GatewayLeader) getHandleNonleaderDeath(ipPort string) func() {
	return func() {
		log.Printf("Dead replica: %s\n", ipPort)
		this.replicas.setAlive(ipPort, false)
		// Starting an election to load balance. Election is not necessary, but elections already
		// have the nice property that they block request processing.
		this.Election(this.ipPort, &api.Empty{})
	}
}

// Rebalance the load on the replicas.
// Use the replica data structure to get new assignments.
// Notify sensors of new assignments.
// This will only be called within elections.
func (this *GatewayLeader) rebalanceLoad() {
	// Use the replica data structure to get new assignments.
	var assigns *map[api.RegisterGatewayUserParams][]api.RegisterParams = this.replicas.rebalanceLoad()
	log.Printf("new assignments: %+v\n", assigns)
	// Notify sensors of new assignments.
	for ipPort, assignees := range *assigns {
		for _, assignee := range assignees {
			var rpcName string
			//Need to map Type to RPC name
			switch assignee.Name {
			case api.Door:
				rpcName = "DoorSensor"
				break
			case api.Motion:
				rpcName = "MotionSensor"
				break
			case api.Temperature:
				rpcName = "TemperatureSensor"
				break
			default:
				log.Printf("Assignee not a sensor: %+v\n", assignee)
				continue
			}
			rpcName += ".ChangeGateway"
			// Calling a synchronous RPC in a new routine.
			// Don't care if the sensor is dead or other communication error.
			var id int
			go util.RpcSync(assignee.Address, assignee.Port, rpcName, ipPort, &id, false)
		}
	}
}

// Set a sinle replica to alive.
func (this *GatewayLeader) setReplicaAlive(key string) {
	log.Printf("Alive replica: %s\n", key)
	this.replicas.setAlive(key, true)
}
