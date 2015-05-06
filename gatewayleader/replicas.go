// This file contains structures representing replicas, their storage,
// and load balancing.

package gatewayleader

import (
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"github.com/ppegusii/cs677-smart-homes-IoT/structs"
	"github.com/ppegusii/cs677-smart-homes-IoT/util"
	"sync"
)

// Struct that contains all necessary replica information.
type replica struct {
	alive  *structs.SyncBool
	ipPort api.RegisterGatewayUserParams
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
	m map[string]replica
}

func newSyncMapStringReplica(replicasIpPort []api.RegisterGatewayUserParams, g *GatewayLeader) *syncMapStringReplica {
	var s *syncMapStringReplica = &syncMapStringReplica{
		m: make(map[string]replica),
	}
	for _, ipPort := range replicasIpPort {
		var key string = util.RegisterGatewayUserParamsToString(ipPort)
		s.m[key] = replica{
			alive:  structs.NewSyncBool(false),
			ipPort: ipPort,
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
		replica.alive.Set(false)
		replica.timer.Stop()
	}
	this.RUnlock()
}

func (this *syncMapStringReplica) ResetTimer(key string) {
	this.RLock()
	r, ok := this.m[key]
	if !ok {
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
	var ipPorts []api.RegisterGatewayUserParams = make([]api.RegisterGatewayUserParams, len(this.m))
	var idx int = 0
	for _, r := range this.m {
		ipPorts[idx] = r.ipPort
		idx++
	}
	this.RUnlock()
	return ipPorts
}
