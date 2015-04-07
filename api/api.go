package api

import ()

type Type int

type PMAP map[int]string

const (
	Sensor Type = iota
	Device Type = iota
)

type Name int

const GatewayID int = 100000

const (
	Bulb        Name = iota
	Door        Name = iota
	Motion      Name = iota
	Outlet      Name = iota
	Temperature Name = iota
)

type State int

const (
	Closed      State = iota
	MotionStart State = iota
	MotionStop  State = iota
	Off         State = iota
	On          State = iota
	Open        State = iota
)

type Mode int

const (
	Away Mode = iota
	Home Mode = iota
	//Logical Mode = iota
	//These states indicate whether the
	//gateway believes smart outlets are
	//on or off.
	OutletsOn  Mode = iota
	OutletsOff Mode = iota
	//Time       Mode = iota
)

type Ordering int

const (
	ClockSync    Ordering = iota
	LogicalClock Ordering = iota
	NoOrder      Ordering = iota
)

type ReportState func(*StateInfo, *struct{}) error

type DatabaseInterface interface {
	AddDeviceOrSensor(params *RegisterParams, _ *struct{}) error
	AddEvent(params *StateInfo, _ *struct{}) error
	AddState(params *StateInfo, _ *struct{}) error
	GetState(params *int, reply *StateInfo) error
	RegisterGateway(params *RegisterGatewayUserParams, _ *struct{}) error
}

type DeviceInterface interface {
	QueryState(params *int, reply *StateInfo) error
	ChangeState(params *StateInfo, reply *StateInfo) error
}

type GatewayInterface interface {
	ChangeMode(params *Mode, _ *struct{}) error
	Register(params *RegisterParams, reply *int) error
	RegisterUser(params *RegisterGatewayUserParams, _ *struct{}) error
	ReportMotion(params *StateInfo, _ *struct{}) error
	ReportDoorState(params *StateInfo, _ *struct{}) error
	SendPeerTable(id int, peers *PMAP) error
}

type OrderingMiddlewareInterface interface {
	//Multicasts new node notification to all other nodes.
	//Called only by the gateway front-end application.
	SendNewNodeNotify(o *OrderingNode) error
	//**Ordinary unicast for clock sync.
	//Logical clocks:
	//Multicasts event notification to all other nodes.
	//Called by applications instead of reporting state directly to another process.
	SendState(s *StateInfo, destAddr string, destPort string) error
	//Register functions that handle the states received inside events.
	RegisterReportState(name Name, reportState ReportState)
}

type OrderingMiddlewareRPCInterface interface {
	//Accepts new node notifications
	//Called only by other ordering implementations.
	ReceiveNewNodeNotify(params *OrderingNode, _ *struct{}) error
	//**Simple delivery of state info to registered report state functions for clock sync.
	//Logical clocks:
	//Multicasts acknowledgement of event to all other nodes.
	//Maintains a queue of messages delivering the one with the least clock value once
	//all acknowledgments have been received. Therefore, there is a total ordering
	//on messages delivered to the application. Those messages are delivered to
	//registered report state functions.
	//Called only by other ordering implementations.
	ReceiveEvent(params *Event, _ *struct{}) error
}

type SensorInterface interface {
	QueryState(params *int, reply *StateInfo) error
}

type UserInterface interface {
	TextMessage(params *string, _ *struct{}) error
}

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

type RegisterParams struct {
	Address  string
	DeviceId int
	Name     Name
	Port     string
	State    State
	Type     Type
}

type RegisterGatewayUserParams struct {
	Address string
	Port    string
}

type StateInfo struct {
	Clock      int
	DeviceId   int
	DeviceName Name
	State      State
}

// Used when gateway sends an update to the other peers about a newly registered device
type PeerInfo struct {
	Token int // Token value 0 means add the new peer , token value 1 means delete the old peer
	DeviceId int
	Address string
}
