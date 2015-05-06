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
	iWonWait         time.Duration = 5 * time.Second // duration to wait for IWon replies
	okWait           time.Duration = 2 * time.Second // duration to wait for OKs
	aliveReplyWait   time.Duration = 2 * time.Second // duration to wait for are you alive replies
	aliveRequestWait time.Duration = 5 * time.Second // duration to wait before sending next are you alive probe
	nonleaderAlive   time.Duration = 6 * time.Second // duration leader waits before consider a nonleader dead
)

// Struct that contains all necessary replica information.
type replica struct {
	alive  *structs.SyncBool
	ipPort api.RegisterGatewayUserParams
	timer  *structs.SyncTimer          // timer used by leader to help determine if replica is dead
	nodes  *structs.SyncMapIntRegParam // stores the sensors currently serviced by this replica, data structure will probably change
}

type GatewayLeader struct {
	aliveReplyTimer   *structs.SyncTimer
	aliveRequestTimer *structs.SyncTimer
	api.GatewayInterface
	electionCheckLock  sync.Mutex
	electionInProgress bool
	ipPort             api.RegisterGatewayUserParams
	iWonTimer          *structs.SyncTimer
	leader             *structs.SyncRegGatewayUserParam
	okTimer            *structs.SyncTimer
	replicas           map[string]replica
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
	var replicaMap map[string]replica = make(map[string]replica)
	for _, ipPort := range replicas {
		var key string = util.RegisterGatewayUserParamsToString(ipPort)
		replicaMap[key] = replica{
			alive:  structs.NewSyncBool(false),
			ipPort: ipPort,
			timer:  structs.NewSyncTimer(nonleaderAlive, g.getHandleNonleaderDeath(key)),
			nodes:  structs.NewSyncMapIntRegParam(),
		}
	}
	g.replicas = replicaMap
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
	// TODO load balance by assigning push sensors to one replica and pull sensors to the other
	// Devices just need an ID. They can be assigned randomly.
	var err error = this.GatewayInterface.Register(params, reply)
	this.RUnlock()
	return err
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

// Receive election msg from self or another replica.
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
	this.setAllReplicasDead()
	var thisId string = util.RegisterGatewayUserParamsToString(this.ipPort)
	// Start a timer that when duration elapses sends IWon.
	this.okTimer.Reset()
	// Send election notice to each replica with a higher id (async RPC).
	for _, replica := range this.replicas {
		var id string = util.RegisterGatewayUserParamsToString(replica.ipPort)
		if id > thisId {
			util.RpcAsync(replica.ipPort.Address, replica.ipPort.Port, "Gateway.Election",
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
	// Send IWon messages to all replicas with lower id
	for _, replica := range this.replicas {
		var id string = util.RegisterGatewayUserParamsToString(replica.ipPort)
		if id < thisId {
			util.RpcAsync(replica.ipPort.Address, replica.ipPort.Port,
				"Gateway.IWon", this.ipPort, &api.RegisterGatewayUserParams{},
				this.handleIWonReplies, false)
		}
	}
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
	if err != nil {
		return
	}
	var key string = util.RegisterGatewayUserParamsToString(
		*(reply.(*api.RegisterGatewayUserParams))) // last bit is a type assertion
	this.setReplicaAlive(key)
}

// Receive alive probes. Reset timer for the requesting replica to detect replica crash
// faults if leader.
func (this *GatewayLeader) Alive(replica api.RegisterGatewayUserParams, yes *api.Empty) error {
	log.Println("Received alive probe")
	// TODO Determine if other replicas are active.
	var key string = util.RegisterGatewayUserParamsToString(replica)
	this.replicas[key].timer.Reset()
	return nil
}

// Get a closure that sets a replica to dead. Used as callback functions replica alive request
// timers for detecting dead replicas if leader.
func (this *GatewayLeader) getHandleNonleaderDeath(ipPort string) func() {
	return func() {
		log.Printf("Dead replica: %s\n", ipPort)
		this.replicas[ipPort].alive.Set(false)
	}
}

// Set all replicas to dead.
func (this *GatewayLeader) setAllReplicasDead() {
	for _, replica := range this.replicas {
		replica.alive.Set(false)
		replica.timer.Stop()
	}
}

// Set a sinle replica to alive.
func (this *GatewayLeader) setReplicaAlive(key string) {
	replica, ok := this.replicas[key]
	if ok {
		log.Printf("Alive replica: %s\n", key)
		replica.alive.Set(true)
	}
}
