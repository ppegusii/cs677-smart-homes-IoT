package ordermw

import (
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"github.com/ppegusii/cs677-smart-homes-IoT/structs"
	"log"
	"net/rpc"
	"time"
)

type Logical struct {
	id           int
	ip           string
	nodes        *structs.SyncMapIntOrderingNode
	port         string
	reportStates *structs.SyncMapNameReportState
}

func NewLogical(id int, ip string, port string) *Logical {
	var l *Logical = &Logical{
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
	var err error = rpc.Register(api.OrderingMiddlewareRPCInterface(this))
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
			log.Printf("dialing error: %+v", err)
			return err
		}
		client.Go("Logical.ReceiveNewNodesNotify", nodes, &empty, nil)
	}
	return nil
}

//Accepts new node notifications
//Called only by other ordering implementations.
func (this *Logical) ReceiveNewNodesNotify(params map[int]api.OrderingNode, _ *struct{}) error {
	log.Printf("Received nodes: %+v", params)
	for id, node := range params {
		this.nodes.Set(id, node)
	}
	log.Printf("My nodes are now: %+v", this.nodes.GetMap())
	return nil
}

//**Ordinary unicast for clock sync.
//Logical clocks:
//Multicasts event notification to all other nodes.
//Called by applications instead of reporting state directly to another process.
func (this *Logical) SendState(s api.StateInfo, destAddr string, destPort string) error {
	var event api.Event = api.Event{
		IsAck:      false,
		SrcAddress: this.ip,
		SrcId:      s.DeviceId,
		SrcPort:    this.port,
		StateInfo:  s,
	}
	var client *rpc.Client
	var err error
	client, err = rpc.Dial("tcp", destAddr+":"+destPort)
	if err != nil {
		log.Fatal("dialing error: %+v", err)
	}
	var empty struct{}
	err = client.Call("Logical.ReceiveEvent", event, &empty)
	if err != nil {
		log.Fatal("calling error: %+v", err)
		return err
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
func (this *Logical) ReceiveEvent(params *api.Event, _ *struct{}) error {
	var rsPtr *api.ReportState
	var ok bool
	rsPtr, ok = this.reportStates.Get(params.StateInfo.DeviceName)
	if !ok {
		log.Printf("No registered func to handle device name: %d", params.StateInfo.DeviceName)
		return nil
	}
	var empty struct{}
	var rs api.ReportState = *rsPtr
	return rs(&(params.StateInfo), &empty)
}

//Register functions that handle the states received inside events.
func (this *Logical) RegisterReportState(name api.Name, reportState api.ReportState) {
	this.reportStates.Set(name, &reportState)
}
