package ordermw

import (
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"github.com/ppegusii/cs677-smart-homes-IoT/structs"
	"log"
	//"net"
	"net/rpc"
)

type Dummy struct {
	id           int
	ip           string
	port         string
	reportStates *structs.SyncMapNameReportState
}

func NewDummy(id int, ip string, port string) *Dummy {
	var d *Dummy = &Dummy{
		id:           id,
		ip:           ip,
		port:         port,
		reportStates: structs.NewSyncMapNameReportState(),
	}
	d.start()
	return d
}
func (this *Dummy) start() {
	//register RPC server
	var err error = rpc.Register(api.OrderingMiddlewareRPCInterface(this))
	if err != nil {
		log.Fatal("rpc.Register error: %s\n", err)
	}
	/*
		var listener net.Listener
		listener, err = net.Listen("tcp", this.ip+":"+this.port)
		if err != nil {
			log.Fatal("net.Listen error: %s\n", err)
		}
		rpc.Accept(listener)
	*/
}

//Multicasts new node notification to all other nodes.
//Called only by the gateway front-end application.
func (this *Dummy) SendNewNodeNotify(o api.OrderingNode) error {
	return nil
}

//Accepts new node notifications
//Called only by other ordering implementations.
func (this *Dummy) ReceiveNewNodeNotify(params *api.OrderingNode, _ *struct{}) error {
	return nil
}

//**Ordinary unicast for clock sync.
//Logical clocks:
//Multicasts event notification to all other nodes.
//Called by applications instead of reporting state directly to another process.
func (this *Dummy) SendState(s api.StateInfo, destAddr string, destPort string) error {
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
	err = client.Call("Dummy.ReceiveEvent", event, &empty)
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
func (this *Dummy) ReceiveEvent(params *api.Event, _ *struct{}) error {
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
func (this *Dummy) RegisterReportState(name api.Name, reportState api.ReportState) {
	this.reportStates.Set(name, &reportState)
}
