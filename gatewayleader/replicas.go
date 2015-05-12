// This file contains structures representing replicas, their storage,
// and load balancing.

package gatewayleader

import (
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"github.com/ppegusii/cs677-smart-homes-IoT/structs"
	"github.com/ppegusii/cs677-smart-homes-IoT/util"
	"log"
	"math"
	"sync"
)

// Struct that contains all necessary replica information.
type replica struct {
	alive     *structs.SyncBool
	ipPort    api.RegisterGatewayUserParams
	isThis    bool                        // true if replica represents this gatewayleader
	deadTimer *structs.SyncTimer          // timer used by leader to help determine if replica is dead
	lastSynch *structs.SyncInt64          // last time replica synced with this gatewayleader
	nodes     *structs.SyncMapIntRegParam // stores the sensors currently serviced by this replica
	syncTimer *structs.SyncTimer          // timer used to enforce eventual consistency
}

// Data structure to store replicas.
// Makes load balancing decisions
// Read operations:
// 		Get replica ipPort, stop or reset timer.
// 		Alive operations also OK since registrations will not occur during elections.
//		Update last sync time.
// Write operations:
//		Any load balancing operations that change the nodes structure.
type syncMapStringReplica struct {
	sync.RWMutex
	dbIpPort api.RegisterGatewayUserParams
	m        map[string]*replica
}

func newSyncMapStringReplica(replicasIpPort []api.RegisterGatewayUserParams, g *GatewayLeader) *syncMapStringReplica {
	var s *syncMapStringReplica = &syncMapStringReplica{
		dbIpPort: g.dbIpPort,
		m:        make(map[string]*replica),
	}
	//add this gateway's ipPort
	replicasIpPort = append(replicasIpPort, g.ipPort)
	for _, ipPort := range replicasIpPort {
		var key string = util.RegisterGatewayUserParamsToString(ipPort)
		var isSelf bool = g.ipPort == ipPort
		var r replica = replica{
			alive:     structs.NewSyncBool(isSelf),
			ipPort:    ipPort,
			isThis:    isSelf,
			deadTimer: structs.NewSyncTimer(nonleaderAliveWait, g.getHandleNonleaderDeath(key)),
			lastSynch: structs.NewSyncInt64(api.EarliestTime),
			nodes:     structs.NewSyncMapIntRegParam(),
		}
		// Create and start a sync timer unless replica represents this gateway leader.
		if g.ipPort != ipPort {
			log.Printf("Starting timer g.ipPort = %+v | ipPort = %+v", g.ipPort, ipPort)
			r.syncTimer = structs.NewSyncTimer(syncWait, g.getHandleSyncTimeout(ipPort))
			r.syncTimer.Reset()
		}
		s.m[key] = &r
	}
	return s
}

func (this *syncMapStringReplica) getAlive(key string) bool {
	this.RLock()
	defer this.RUnlock()
	r, ok := this.m[key]
	if !ok {
		log.Printf("Replica key not found: %s", key)
		return false
	}
	return r.alive.Get()
}

func (this *syncMapStringReplica) setAlive(key string, alive bool) {
	this.RLock()
	r, ok := this.m[key]
	if !ok {
		this.RUnlock()
		return
	}
	r.alive.Set(alive)
	this.RUnlock()
}

func (this *syncMapStringReplica) setAllReplicasDead() {
	this.RLock()
	for _, replica := range this.m {
		// Never kill yourself!
		if replica.isThis {
			continue
		}
		replica.alive.Set(false)
		replica.deadTimer.Stop()
	}
	this.RUnlock()
}

func (this *syncMapStringReplica) resetDeadTimer(key string) {
	this.RLock()
	r, ok := this.m[key]
	//Never reset your own timer
	if !ok || r.isThis {
		this.RUnlock()
		return
	}
	r.deadTimer.Reset()
	this.RUnlock()
}

func (this *syncMapStringReplica) stopDeadTimer(key string) {
	this.RLock()
	r, ok := this.m[key]
	if !ok {
		this.RUnlock()
		return
	}
	r.deadTimer.Stop()
	this.RUnlock()
}

func (this *syncMapStringReplica) resetSyncTimer(key string) {
	this.RLock()
	r, ok := this.m[key]
	//Never reset your own timer
	if !ok || r.isThis {
		this.RUnlock()
		return
	}
	r.syncTimer.Reset()
	this.RUnlock()
}

func (this *syncMapStringReplica) getIpPorts() []api.RegisterGatewayUserParams {
	this.RLock()
	//length - 1 to disclude your own ipPort
	var ipPorts []api.RegisterGatewayUserParams = make([]api.RegisterGatewayUserParams, len(this.m)-1)
	var idx int = 0
	for _, r := range this.m {
		// Replica representing this gatewayleader is for internal
		// use only so do not return it.
		if r.isThis {
			continue
		}
		ipPorts[idx] = r.ipPort
		idx++
	}
	this.RUnlock()
	return ipPorts
}

// Choose replica for load balancing
func (this *syncMapStringReplica) loadBalance(regParams api.RegisterParams, id int) *api.RegisterGatewayUserParams {
	this.Lock()
	defer this.Unlock()
	return this.simpleLoadBalance(regParams, id)
}

// Assigns device/sensor to the replica with the lowest number of sensors.
// Devices will not be tracked since they do not affect gateway load.
func (this *syncMapStringReplica) simpleLoadBalance(regParams api.RegisterParams, id int) *api.RegisterGatewayUserParams {
	regParams.DeviceId = id
	var smallestLoad int = math.MaxInt32
	var assigned *replica
	for _, r := range this.m {
		log.Printf("r = %s, r.nodes.Size() = %d\n", r, r.nodes.Size())
		// Never assign to a dead replica
		if !r.alive.Get() {
			log.Printf("simpleLoadBalance dead replica\n")
			continue
		}
		if r.nodes.Size() < smallestLoad {
			smallestLoad = r.nodes.Size()
			assigned = r
		}
	}
	ipPort := assigned.ipPort
	// Only track if sensor.
	if regParams.Type == api.Sensor {
		assigned.nodes.AddExistingRegParam(&regParams, regParams.DeviceId)
	}
	return &ipPort
}

// Called after an election or after replica crash fault detection.
// Redistibutes load and returns a map of replica ip ports to sensor info.
// Rebalances load by calling a load balancing function on reassigned sensors.
func (this *syncMapStringReplica) rebalanceLoad() *map[api.RegisterGatewayUserParams][]api.RegisterParams {
	this.Lock()
	// create the map to return
	// no need to create slice, since nils act like zero length slices
	var newAssigns map[api.RegisterGatewayUserParams][]api.RegisterParams = make(
		map[api.RegisterGatewayUserParams][]api.RegisterParams)
	// all nodes will be reassigned, pull from db
	var data api.ConsistencyData
	var err error = util.RpcSync(this.dbIpPort.Address, this.dbIpPort.Port,
		"Database.GetDataSince", api.EarliestTime, &data, false)
	if err != nil {
		return nil
	}
	// delete assignments by creating new data structure
	for _, r := range this.m {
		// empty replica's nodes by creating new data structure
		r.nodes = structs.NewSyncMapIntRegParam()
	}
	// reassign nodes
	//for _, assignee := range assignees {
	for _, assignee := range data.RegisteredNodes {
		// assign node to gateway
		var assign *api.RegisterGatewayUserParams = this.simpleLoadBalance(assignee, assignee.DeviceId)
		// add that node to return map for node notification
		newAssigns[*assign] = append(newAssigns[*assign], assignee)
	}
	this.Unlock()
	return &newAssigns
}

func (this *syncMapStringReplica) getAssignments() *map[api.RegisterGatewayUserParams][]api.RegisterParams {
	this.RLock()
	var assigns map[api.RegisterGatewayUserParams][]api.RegisterParams = make(
		map[api.RegisterGatewayUserParams][]api.RegisterParams)
	for _, r := range this.m {
		assigns[r.ipPort] = *(r.nodes.GetAllRegParams())
	}
	this.RUnlock()
	return &assigns
}

func (this *syncMapStringReplica) setAssignments(assigns *map[api.RegisterGatewayUserParams][]api.RegisterParams) {
	this.Lock()
	for ipPort, nodes := range *assigns {
		var id string = util.RegisterGatewayUserParamsToString(ipPort)
		this.m[id].nodes = structs.NewSyncMapIntRegParam()
		for _, node := range nodes {
			this.m[id].nodes.AddExistingRegParam(&node, node.DeviceId)
		}
		log.Printf("After setAssignments %s has %+v\n", id, this.m[id].nodes.GetAllRegParams())
	}
	this.Unlock()
}

func (this *syncMapStringReplica) getReplicaLastSyncTime(id string) int64 {
	this.RLock()
	defer this.RUnlock()
	return this.m[id].lastSynch.Get()
}

// Change sync time and restart timer
func (this *syncMapStringReplica) setReplicaLastSyncTime(clock int64, id string) {
	this.RLock()
	this.m[id].syncTimer.Stop()
	this.m[id].lastSynch.Set(clock)
	this.m[id].syncTimer.Reset()
	this.RUnlock()
}

// Return true if the node is assigned to this gatewayleader
func (this *syncMapStringReplica) isNodeIsAssignedToMe(selfId string, nodeId int) bool {
	this.RLock()
	defer this.RUnlock()
	var self *replica
	var ok bool
	self, ok = this.m[selfId]
	if !ok || !self.isThis {
		log.Printf("Given incorrect selfId: %s\n", selfId)
	}
	_, ok = self.nodes.GetRegParam(nodeId)
	return ok
}
