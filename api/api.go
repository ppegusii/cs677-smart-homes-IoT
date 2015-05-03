// This file declares all the structs and interfaces needed throughout the system

package api

import ()

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
	// ***These modes should be replaced with cache queries.***
	//These states indicate whether the
	//gateway believes smart outlets are
	//on or off.
	OutletsOn  Mode = iota
	OutletsOff Mode = iota
)

// Used to log gateway state in database.
type ModeAndClock struct {
	Clock int
	Mode  Mode
}

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

// Interface implemented by devices and sensors
// Sensors support ChangeState for testing
type NodeInterface interface {
	QueryState(params *int, reply *StateInfo) error
	ChangeState(params *StateInfo, reply *StateInfo) error
}

// ***This interface needs to be cleaned. We should be able to
// have meaningful returns to node RPC calls.***
// Interface provided by the Gateway
type GatewayInterface interface {
	//ChangeMode(params *Mode, _ *struct{}) error
	Query(params Name, _ *struct{}) error
	Register(params *RegisterParams, reply *int) error
	RegisterUser(params *RegisterGatewayUserParams, _ *struct{}) error
	//ReportBulbState(params *StateInfo, _ *struct{}) error
	ReportDoorState(params *StateInfo, _ *struct{}) error
	ReportMotion(params *StateInfo, _ *struct{}) error
	//ReportOutletState(params *StateInfo, _ *struct{}) error
	//ReportTemperature(params *StateInfo, _ *struct{}) error
	//RegisterAck(id int, _ *struct{}) error
	// Initialize the gateway and start routines
	Start()
}

// A dummy wrapper of GatewayInterface to test
// embedding.
type GatewayWrapperInterface interface {
	// embedded gateway
	GatewayInterface
	RpcSyncInterface
	SetGateway(GatewayInterface)
}

// Provides the following consistency guarantees between replicated gateways:
// 1. Entry consistency before home/away mode decisions.
// 2. Eventual consistency:
//    1. Sync when gateway comes up.
//    2. Syncs must occur within a certain duration.
type GatewayConsistencyInterface interface {
	GatewayInterface
}

// Elects a leader gateway that will respond to all registration requests.
// Load balances connected devices between gateways.
type GatewayLeaderInterface interface {
	GatewayConsistencyInterface
}

type RpcSyncInterface interface {
	// Function to make a synchronous RPC
	// Arguments: IP, port, rpcName, args, reply, isErrorFatal
	// Returns: err
	RpcSync(string, string, string, interface{}, interface{}, bool) error
}

// Interface needed to send text messages to the user incase the Mode is set to AWAY and motion is detected
type UserInterface interface {
	TextMessage(params *string, _ *struct{}) error
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
