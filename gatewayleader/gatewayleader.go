// This file wraps the gateway. It provides gateway leader election,
// crash fault detection, and load balancing between gateway
// replicas. It also waits before servicing a request during elections.

package gatewayleader

import (
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"github.com/ppegusii/cs677-smart-homes-IoT/structs"
	"github.com/ppegusii/cs677-smart-homes-IoT/util"
	"log"
	"sync"
	"time"
)

const (
	iWonWait         time.Duration = 5 * time.Second
	okWait           time.Duration = 2 * time.Second
	aliveReplyWait   time.Duration = 2 * time.Second
	aliveRequestWait time.Duration = 5 * time.Second
)

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
	replicas           []api.RegisterGatewayUserParams
	sync.RWMutex       // Used to wait before sending application messages during an election.
}

func NewGatewayLeader(ip, port string, replicas []api.RegisterGatewayUserParams) api.GatewayLeaderInterface {
	var g GatewayLeader = GatewayLeader{
		leader:   structs.NewSyncRegGatewayUserParam(),
		replicas: replicas,
		ipPort: api.RegisterGatewayUserParams{
			Address: ip,
			Port:    port,
		},
	}
	g.iWonTimer = structs.NewSyncTimer(iWonWait, g.startElection)
	g.okTimer = structs.NewSyncTimer(okWait, g.sendIWons)
	g.aliveReplyTimer = structs.NewSyncTimer(aliveReplyWait, g.handleAliveTimeout)
	g.aliveRequestTimer = structs.NewSyncTimer(aliveRequestWait, g.pollLeader)
	return &g
}

func (this *GatewayLeader) RpcSync(ip, port, rpcName string, args interface{}, reply interface{}, isErrFatal bool) error {
	log.Printf("Before RPC reply: %+v\n", reply)
	var err error = util.RpcSync(ip, port, rpcName, args, reply, isErrFatal)
	log.Printf("After RPC reply: %+v\n", reply)
	return err
}

func (this *GatewayLeader) SetGateway(g api.GatewayInterface) {
	this.GatewayInterface = g
}

func (this *GatewayLeader) StartLeader() {
	this.Election(this.ipPort, &api.Empty{})
}

func (this *GatewayLeader) Register(params *api.RegisterParams, reply *int) error {
	this.RLock()
	// TODO only service request if leader,
	// load balance
	log.Printf("Before Register id: %d\n", *reply)
	var err error = this.GatewayInterface.Register(params, reply)
	log.Printf("After Register id: %d\n", *reply)
	this.RUnlock()
	return err
}

func (this *GatewayLeader) RegisterUser(params *api.RegisterGatewayUserParams, empty *struct{}) error {
	this.RLock()
	var err error = this.GatewayInterface.RegisterUser(params, empty)
	this.RUnlock()
	return err
}

func (this *GatewayLeader) ReportDoorState(params *api.StateInfo, empty *struct{}) error {
	this.RLock()
	var err error = this.GatewayInterface.ReportDoorState(params, empty)
	this.RUnlock()
	return err
}

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
func (this *GatewayLeader) IWon(replica api.RegisterGatewayUserParams, _ *api.Empty) error {
	log.Printf("Received iwon msg from: %+v\n", replica)
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

// Handle alive reply timeout.
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

// Handles alive replies from active replica
func (this *GatewayLeader) handleAlive(_ interface{}, err error) {
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

func (this *GatewayLeader) startElection() {
	log.Println("Starting election")
	var thisId string = util.RegisterGatewayUserParamsToString(this.ipPort)
	// Start a timer that when duration elapses sends IWon.
	this.okTimer.Reset()
	// Send election notice to each replica with a higher id (async RPC).
	for _, replica := range this.replicas {
		var id string = util.RegisterGatewayUserParamsToString(replica)
		if id > thisId {
			util.RpcAsync(replica.Address, replica.Port, "Gateway.Election",
				this.ipPort, &api.Empty{}, this.handleOKs, false)
		}
	}
}

// Handles OK replies in response to election messages
func (this *GatewayLeader) handleOKs(_ interface{}, err error) {
	log.Println("Handling OK")
	if err != nil {
		return
	}
	this.okTimer.Stop()
	// Start a timer that when duration elapses restarts election.
	this.iWonTimer.Reset()
}

func (this *GatewayLeader) sendIWons() {
	log.Println("Sending IWons")
	var thisId string = util.RegisterGatewayUserParamsToString(this.ipPort)
	// Declare self leader
	this.leader.Set(this.ipPort)
	// Send IWon messages to all replicas with lower id
	for _, replica := range this.replicas {
		var id string = util.RegisterGatewayUserParamsToString(replica)
		if id < thisId {
			go util.RpcSync(replica.Address, replica.Port, "Gateway.IWon",
				this.ipPort, &api.Empty{}, false)
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

func (this *GatewayLeader) Alive(replica api.RegisterGatewayUserParams, yes *api.Empty) error {
	log.Println("Received alive probe")
	return nil
}
