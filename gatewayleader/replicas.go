// This file contains structures representing replicas, their storage,
// and load balancing.

package gatewayleader

import (
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"github.com/ppegusii/cs677-smart-homes-IoT/structs"
	"github.com/ppegusii/cs677-smart-homes-IoT/util"
	"math"
	"sync"
)

// Struct that contains all necessary replica information.
type replica struct {
	alive  *structs.SyncBool
	ipPort api.RegisterGatewayUserParams
	isThis bool                        // true if replica represents this gatewayleader
	timer  *structs.SyncTimer          // timer used by leader to help determine if replica is dead
	nodes  *structs.SyncMapIntRegParam // stores the sensors currently serviced by this replica
}

// Data structure to store replicas.
// Makes load balancing decisions
// Read operations:
// 		Get replica ipPort, stop or reset timer.
// 		Alive operations also OK since registrations will not occur during elections.
// Write operations:
//		Any load balancing operations that change the nodes structure.
type syncMapStringReplica struct {
	sync.RWMutex
	m map[string]*replica
}

func newSyncMapStringReplica(replicasIpPort []api.RegisterGatewayUserParams, g *GatewayLeader) *syncMapStringReplica {
	var s *syncMapStringReplica = &syncMapStringReplica{
		m: make(map[string]*replica),
	}
	//add this gateway's ipPort
	replicasIpPort = append(replicasIpPort, g.ipPort)
	for _, ipPort := range replicasIpPort {
		var key string = util.RegisterGatewayUserParamsToString(ipPort)
		var isSelf bool = g.ipPort == ipPort
		s.m[key] = &replica{
			alive:  structs.NewSyncBool(isSelf),
			ipPort: ipPort,
			isThis: isSelf,
			timer:  structs.NewSyncTimer(nonleaderAliveWait, g.getHandleNonleaderDeath(key)),
			nodes:  structs.NewSyncMapIntRegParam(),
		}
	}
	return s
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
		replica.timer.Stop()
	}
	this.RUnlock()
}

func (this *syncMapStringReplica) ResetTimer(key string) {
	this.RLock()
	r, ok := this.m[key]
	//Never reset your own timer
	if !ok || r.isThis {
		this.RUnlock()
		return
	}
	r.timer.Reset()
	this.RUnlock()
}

func (this *syncMapStringReplica) StopTimer(key string) {
	this.RLock()
	r, ok := this.m[key]
	if !ok {
		this.RUnlock()
		return
	}
	r.timer.Stop()
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
		// Never assign to a dead replica
		if !r.alive.Get() {
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
	//create the data structure to hold nodes
	var assignees []api.RegisterParams
	// all nodes will be reassigned, gather them
	for _, r := range this.m {
		assignees = append(assignees, *(r.nodes.GetAllRegParams())...)
		// empty replica's nodes by creating new data structure
		r.nodes = structs.NewSyncMapIntRegParam()
	}
	// reassign nodes
	for _, assignee := range assignees {
		// assign node to gateway
		var assign *api.RegisterGatewayUserParams = this.simpleLoadBalance(assignee, assignee.DeviceId)
		// add that node to return map for node notification
		newAssigns[*assign] = append(newAssigns[*assign], assignee)
	}
	this.Unlock()
	return &newAssigns
}