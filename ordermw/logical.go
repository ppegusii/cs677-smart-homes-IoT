// This file declares all the structs and interfaces needed by logical clock
package ordermw

import (
	"github.com/nu7hatch/gouuid"
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"github.com/ppegusii/cs677-smart-homes-IoT/structs"
	"log"
	"net/rpc"
	"time"
)

// Defines the logical clock structure
type Logical struct {
	clock        *structs.SyncInt
	events       *structs.SyncLogicalEventContainer
	id           int
	ip           string
	nodes        *structs.SyncMapIntOrderingNode
	port         string
	reportStates *structs.SyncMapNameReportState
}

// Initialize a new logical clock
func NewLogical(id int, ip string, port string) *Logical {
	var l *Logical = &Logical{
		clock:        structs.NewSyncInt(0),
		events:       structs.NewSyncLogicalEventContainer(),
		id:           id,
		ip:           ip,
		nodes:        structs.NewSyncMapIntOrderingNode(),
		port:         port,
		reportStates: structs.NewSyncMapNameReportState(),
	}
	l.nodes.Set(id,
		api.OrderingNode{
			Address: ip,
			ID:      id,
			Port:    port,
		})
	l.start()
	return l
}
func (this *Logical) start() {
	//register RPC server
	var err error = rpc.Register(api.OrderingMiddlewareLogicalRPCInterface(this))
	if err != nil {
		log.Fatal("rpc.Register error: %s\n", err)
	}
}

//Multicasts new node notification to all other nodes.
//Called only by the gateway front-end application.
func (this *Logical) SendNewNodeNotify(o api.OrderingNode) error {
	//poor style the remote RPC server should be up
	//but we'll give it a second to get it's ID and start
	timer := time.NewTimer(time.Second)
	<-timer.C

	this.nodes.Set(o.ID, o)
	var err error
	var empty struct{}
	var client *rpc.Client
	var nodes map[int]api.OrderingNode = this.nodes.GetMap()
	for id, node := range nodes {
		if id == this.id {
			continue
		}
		client, err = rpc.Dial("tcp", node.Address+":"+node.Port)
		if err != nil {
			log.Printf("dialing error: %+v\n", err)
			return err
		}
		//client.Go("Logical.ReceiveNewNodesNotify", nodes, &empty, nil)
		err = client.Call("Logical.ReceiveNewNodesNotify", nodes, &empty)
		client.Close()
		if err != nil {
			log.Printf("calling error: %+v\n", err)
		}
	}
	return nil
}

//Accepts new node notifications
//Called only by other ordering implementations.
func (this *Logical) ReceiveNewNodesNotify(params map[int]api.OrderingNode, _ *struct{}) error {
	//log.Printf("Received nodes: %+v\n", params)
	for id, node := range params {
		this.nodes.Set(id, node)
	}
	//log.Printf("My nodes are now: %+v\n", this.nodes.GetMap())
	return nil
}

//**Ordinary unicast for clock sync.
//Logical clocks:
//Multicasts event notification to all other nodes.
//Called by applications instead of reporting state directly to another process.
func (this *Logical) SendState(s api.StateInfo, destAddr string, destPort string) error {
	//increment clock then add clock to state info
	s.Clock = this.clock.IncThenGet()
	var eventID *uuid.UUID
	var err error
	eventID, err = uuid.NewV4()
	if err != nil {
		log.Fatal("Error creating uuid: %+v\n", err)
	}
	var event api.LogicalEvent = api.LogicalEvent{
		DestIDs:    this.nodes.GetKeys(),
		EventID:    *eventID,
		IsAck:      false,
		SrcAddress: this.ip,
		SrcId:      this.id,
		SrcPort:    this.port,
		StateInfo:  s,
	}
	return this.multicastEvent(event)
}

//Multicasts events to all other nodes.
func (this *Logical) multicastEvent(event api.LogicalEvent) error {
	var client *rpc.Client
	var empty struct{}
	var err error
	var nodes map[int]api.OrderingNode = this.nodes.GetMap()
	for idx := range event.DestIDs {
		var id int = event.DestIDs[idx]
		node, _ := nodes[id]
		client, err = rpc.Dial("tcp", node.Address+":"+node.Port)
		if err != nil {
			log.Printf("multicast dialing error id=%d nodes=%+v: %+v\n", id, nodes, err)
			return err
		}
		//client.Go("Logical.ReceiveEvent", event, &empty, nil)
		err = client.Call("Logical.ReceiveEvent", event, &empty)
		client.Close()
		if err != nil {
			log.Printf("calling error: %+v\n", err)
		}
	}
	return nil
}

//**Simple delivery of state info to registered report state functions for clock sync.
//Logical clocks:
//Multicasts acknowledgement of event to all other nodes.
//Maintains a queue of messages delivering the one with the least clock value once
//all acknowledgments have been received. Therefore, there is a total ordering
//on messages delivered to the application. Those messages are delivered to
//registered report state functions.
//Called only by other ordering implementations.
func (this *Logical) ReceiveEvent(params api.LogicalEvent, _ *struct{}) error {
	//set clock to the max of own and received
	if params.StateInfo.Clock > this.clock.Get() {
		this.clock.Set(params.StateInfo.Clock)
	}
	//increment clock then set the clock value in state info
	params.StateInfo.Clock = this.clock.IncThenGet()
	if !params.IsAck {
		//enqueue the event
		this.events.AddEvent(params)
		//multicast event acknowledgement
		params.SrcId = this.id
		params.IsAck = true
		return this.multicastEvent(params)
	}
	//this is an acknowledgement
	//add acknowledgement to event
	this.events.AddAck(params)
	//while event at head of queue is fully acknowledged, deliver to application
	var event *api.LogicalEvent
	var ok bool
	for {
		//try to get event
		event, ok = this.events.GetHeadIfAcked()
		if !ok {
			break
		}
		log.Printf("Sending message with this id to app: %+v\n", event.EventID)
		//send event to application if it has a func registered
		var rsPtr *api.ReportState
		rsPtr, ok = this.reportStates.Get(params.StateInfo.DeviceName)
		if !ok {
			log.Printf("No registered func to handle device name: %d\n", params.StateInfo.DeviceName)
			continue
		}
		var empty struct{}
		var rs api.ReportState = *rsPtr
		go rs(&(event.StateInfo), &empty)
	}
	return nil
}

//Register functions that handle the states received inside events.
func (this *Logical) RegisterReportState(name api.Name, reportState api.ReportState) {
	this.reportStates.Set(name, &reportState)
}

//Send PeerTable to other middlewares
func (this *Logical) SendPeertableNotification(i int) {}
