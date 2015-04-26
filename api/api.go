// This file declares all the structs and interfaces needed throughout the system

package api

import (
	"github.com/nu7hatch/gouuid"
)

//Peer Map Table
type PMAP map[int]string

//GatewayID
const GatewayID int = 0

type Type int

// Device types
const (
	InvalidType Type = iota
	Sensor      Type = iota
	Device      Type = iota
)

type Name int

//Device Names
const (
	InvalidName Name = iota
	Bulb        Name = iota
	Door        Name = iota
	Motion      Name = iota
	Outlet      Name = iota
	Temperature Name = iota
)

type State int

// Different states of devices and sensors
const (
	InvalidState State = iota //0
	Closed       State = iota //1
	MotionStart  State = iota //2
	MotionStop   State = iota //3
	Off          State = iota //4
	On           State = iota //5
	Open         State = iota //6
)

type Mode int

const (
	InvalidMode Mode = iota
	Away        Mode = iota
	Home        Mode = iota
	//Logical Mode = iota
	//These states indicate whether the
	//gateway believes smart outlets are
	//on or off.
	OutletsOn  Mode = iota
	OutletsOff Mode = iota
	//Time       Mode = iota
)

type ModeAndClock struct {
	Clock int
	Mode  Mode
}

// Ordering type
type Ordering int

const (
	InvalidOrdering Ordering = iota
	ClockSync       Ordering = iota
	LogicalClock    Ordering = iota
	NoOrder         Ordering = iota
	FaultTolerant	Ordering = iota
)

type StaticOrderingID int

const (
	DatabaseOID StaticOrderingID = -1
	GatewayOID  StaticOrderingID = 0
)

type ReportState func(*StateInfo, *struct{}) error

// Interfaces provided by the Database layer
type DatabaseInterface interface {
	AddDeviceOrSensor(params *RegisterParams, _ *struct{}) error
	AddEvent(params *StateInfo, _ *struct{}) error
	AddState(params *StateInfo, _ *struct{}) error
	GetHappensBefore(params StateInfo, reply *StateInfo) error
	//log the gateway mode
	LogMode(params ModeAndClock, _ *struct{}) error
	RegisterGateway(params *RegisterGatewayUserParams, _ *struct{}) error
}

//Interface or RPC stubs provided by the Devices
type DeviceInterface interface {
	QueryState(params *int, reply *StateInfo) error
	ChangeState(params *StateInfo, reply *StateInfo) error
}

// Interface provided by the Gateway
type GatewayInterface interface {
	ChangeMode(params *Mode, _ *struct{}) error
	Query(params Name, _ *struct{}) error
	Register(params *RegisterParams, reply *int) error
	RegisterUser(params *RegisterGatewayUserParams, _ *struct{}) error
	ReportBulbState(params *StateInfo, _ *struct{}) error
	ReportDoorState(params *StateInfo, _ *struct{}) error
	ReportMotion(params *StateInfo, _ *struct{}) error
	ReportOutletState(params *StateInfo, _ *struct{}) error
	ReportTemperature(params *StateInfo, _ *struct{}) error
	RegisterAck(id int, _ *struct{}) error
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

//Interfaces provided by the Sensor
type SensorInterface interface {
	ChangeState(params *StateInfo, reply *StateInfo) error
	QueryState(params *int, reply *StateInfo) error
}

// Interface needed to send text messages to the user incase the Mode is set to AWAY and motion is detected
type UserInterface interface {
	TextMessage(params *string, _ *struct{}) error
}

// General Event struct
type Event struct {
	IsAck      bool
	SrcAddress string
	SrcId      int
	SrcPort    string
	StateInfo  StateInfo
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

type OrderingNode struct {
	Address string
	ID      int
	Port    string
}

//Structure used during device registration,
//it is send as one of the parameters during RPC Register call to gateway
type RegisterParams struct {
	Address  string
	DeviceId int
	Name     Name
	Port     string
	State    State
	Type     Type
}

//Struct for set and get methods where only IP and port are needed
type RegisterGatewayUserParams struct {
	Address string
	Port    string
}

// To report the state use this struct
type StateInfo struct {
	Clock      int
	DeviceId   int
	DeviceName Name
	State      State
}
