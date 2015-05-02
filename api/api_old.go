//These are old declarations that are probably not needed for lab 3.

package api

import (
	"github.com/nu7hatch/gouuid"
)

//Peer Map Table
type PMAP map[int]string

//GatewayID
const GatewayID int = 0

// Ordering type
type Ordering int

const (
	InvalidOrdering Ordering = iota
	ClockSync       Ordering = iota
	LogicalClock    Ordering = iota
	NoOrder         Ordering = iota
	FaultTolerant   Ordering = iota
)

// Device and Sensor interfaces are the same?***
//Interface or RPC stubs provided by the Devices
type DeviceInterface interface {
	QueryState(params *int, reply *StateInfo) error
	ChangeState(params *StateInfo, reply *StateInfo) error
}

//Interfaces provided by the Sensor
type SensorInterface interface {
	ChangeState(params *StateInfo, reply *StateInfo) error
	QueryState(params *int, reply *StateInfo) error
}

type OrderingMiddlewareInterface interface {
	//Multicasts new node notification to all other nodes.
	//Called only by the gateway front-end application.
	SendNewNodeNotify(o OrderingNode) error
	//**Ordinary unicast for clock sync.
	//Logical clocks:
	//Multicasts event notification to all other nodes.
	//Called by applications instead of reporting state directly to another process.
	SendState(s StateInfo, destAddr string, destPort string) error
	//Register functions that handle the states received inside events.
	RegisterReportState(name Name, reportState ReportState)
	//This function is called when a new device send acknowledgement of registration
	//The main purpose of this function is to send the peertable to all the peers.
	SendPeertableNotification(i int)
	//Bully()
	//GetTime()
}

type OrderingMiddlewareRPCInterface interface {
	//Accepts new node notifications
	//Called only by other ordering implementations.
	ReceiveNewNodesNotify(params map[int]OrderingNode, _ *struct{}) error
	//**Simple delivery of state info to registered report state functions for clock sync.
	//Logical clocks:
	//Multicasts acknowledgement of event to all other nodes.
	//Maintains a queue of messages delivering the one with the least clock value once
	//all acknowledgments have been received. Therefore, there is a total ordering
	//on messages delivered to the application. Those messages are delivered to
	//registered report state functions.
	//Called only by other ordering implementations.
	ReceiveEvent(params *Event, _ *struct{}) error
	ReceivePeertableNotification(params *PMAP, _ *struct{}) error
	//	SendPeertableNotification(params *PMAP, _ *struct{}) error
	Election(id int, _ *struct{}) error
	OKAY(id int, _ *struct{}) error
	IWIN(id int, _ *struct{}) error
	SendTime(id int, timestamp *BTimeStamp) error
	ReceiveOffset(offset int32, _ *struct{}) error
}

//Berkeley Clock Timestamp
type BTimeStamp struct {
	DeviceId  int
	Timestamp int32
}

type OrderingMiddlewareLogicalRPCInterface interface {
	//Accepts new node notifications
	//Called only by other ordering implementations.
	ReceiveNewNodesNotify(params map[int]OrderingNode, _ *struct{}) error
	//**Simple delivery of state info to registered report state functions for clock sync.
	//Logical clocks:
	//Multicasts acknowledgement of event to all other nodes.
	//Maintains a queue of messages delivering the one with the least clock value once
	//all acknowledgments have been received. Therefore, there is a total ordering
	//on messages delivered to the application. Those messages are delivered to
	//registered report state functions.
	//Called only by other ordering implementations.
	ReceiveEvent(params LogicalEvent, _ *struct{}) error
}

// Logical clock event struct
type LogicalEvent struct {
	EventID    uuid.UUID
	DestIDs    []int
	IsAck      bool
	SrcAddress string
	SrcId      int
	SrcPort    string
	StateInfo  StateInfo
}

type StaticOrderingID int

const (
	DatabaseOID StaticOrderingID = -1
	GatewayOID  StaticOrderingID = 0
)

type ReportState func(*StateInfo, *struct{}) error

// General Event struct
type Event struct {
	IsAck      bool
	SrcAddress string
	SrcId      int
	SrcPort    string
	StateInfo  StateInfo
}

type OrderingNode struct {
	Address string
	ID      int
	Port    string
}
