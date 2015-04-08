package api

import ()

type Type int

// Device types
const (
	Sensor Type = iota
	Device Type = iota
)

type Name int

//Device Names
const (
	Bulb        Name = iota
	Door        Name = iota
	Motion      Name = iota
	Outlet      Name = iota
	Temperature Name = iota
)

type State int

// Different states of devices and sensors
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

// Interfaces provided by the Database layer
type DatabaseInterface interface {
	AddDeviceOrSensor(params *RegisterParams, _ *struct{}) error
	AddEvent(params *StateInfo, _ *struct{}) error
	AddState(params *StateInfo, _ *struct{}) error
	GetState(params *int, reply *StateInfo) error
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
	Register(params *RegisterParams, reply *int) error
	RegisterUser(params *RegisterGatewayUserParams, _ *struct{}) error
	ReportBulbState(params *StateInfo, _ *struct{}) error
	ReportDoorState(params *StateInfo, _ *struct{}) error
	ReportMotion(params *StateInfo, _ *struct{}) error
	ReportOutletState(params *StateInfo, _ *struct{}) error
	ReportTemperature(params *StateInfo, _ *struct{}) error
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

//Interfaces provided by the Sensor
type SensorInterface interface {
	QueryState(params *int, reply *StateInfo) error
}

// Interface needed to send text messages to the user incase the Mode is set to AWAY and motion is detected
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

type StateInfo struct {
	Clock      int
	DeviceId   int
	DeviceName Name
	State      State
}
