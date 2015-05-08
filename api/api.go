// This file declares all the structs and interfaces needed throughout the system

package api

import ()

const UNREGISTERED = -10
const EMPTY = 0

//type Type int
type Type string

// Device types
const (
	/*
		InvalidType Type = iota
		Sensor      Type = iota
		Device      Type = iota
	*/
	Sensor Type = "sensor"
	Device Type = "device"
)

//type Name int
type Name string

//Device Names
const (
	/*
		InvalidName Name = iota
		Bulb        Name = iota
		Door        Name = iota
		Motion      Name = iota
		Outlet      Name = iota
		Temperature Name = iota
	*/
	Bulb        Name = "bulb"
	Door        Name = "door"
	Motion      Name = "motion"
	Outlet      Name = "outlet"
	Temperature Name = "temp"
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

//type Mode int
type Mode string

const (
	/*
		InvalidMode Mode = iota
		Away        Mode = iota
		Home        Mode = iota
		// ***These modes should be replaced with cache queries.***
		//These states indicate whether the
		//gateway believes smart outlets are
		//on or off.
		OutletsOn  Mode = iota
		OutletsOff Mode = iota
	*/
	Away       Mode = "away"
	Home       Mode = "home"
	OutletsOn  Mode = "outletson"
	OutletsOff Mode = "outletsoff"
)

const EarliestTime int64 = 0

// Interfaces provided by the Database layer
type DatabaseInterface interface {
	AddDeviceOrSensor(params *RegisterParams, _ *struct{}) error
	AddEvent(params *StateInfo, _ *struct{}) error
	AddState(params *StateInfo, _ *struct{}) error
	GetDataSince(clock int64, data *ConsistencyData) error // For synchronization
	GetHappensBefore(params StateInfo, reply *StateInfo) error
	//log the gateway mode
	LogMode(params ModeAndClock, _ *struct{}) error
	LogLoad(params *map[RegisterGatewayUserParams][]RegisterParams, _ *Empty) error
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
	PullData(clock int64, data *ConsistencyData) error // For synchronization
	PushData(data *ConsistencyData, _ *Empty) error    // For synchronization
	Query(params Name, _ *struct{}) error              // Used for testing
	Register(params *RegisterParams, reply *RegisterReturn) error
	RegisterUser(params *RegisterGatewayUserParams, _ *struct{}) error
	ReportDoorState(params *StateInfo, _ *struct{}) error
	ReportMotion(params *StateInfo, _ *struct{}) error
	Start() // Initialize the gateway and start routines
	//UpdateSensorAssignments(params []api.RegisterParams, _ *Empty) error
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
	// TODO Switch from embedding GatewayInterface
	// to GatewayConsistencyInterface
	//GatewayConsistencyInterface
	GatewayInterface
	RpcSyncInterface
	SetGateway(GatewayInterface)
	Alive(replica RegisterGatewayUserParams, yes *Empty) error
	Election(replica RegisterGatewayUserParams, ok *Empty) error
	IWon(replica RegisterGatewayUserParams, reply *RegisterGatewayUserParams) error
	StartLeader()
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

// Returns a clock value
type ClockInterface interface {
	GetClock() int64
}

//Structure used during device registration,
//it is send as one of the parameters during RPC Register call to gateway
type RegisterParams struct {
	Address  string
	Clock    int64
	DeviceId int
	Name     Name
	Port     string
	State    State
	Type     Type
}

// You don't see this.
func (this RegisterParams) GetClock() int64 {
	return this.Clock
}

type RegisterReturn struct {
	DeviceId int
	Address  string // Gateway Address to send subsequent requests to
	Port     string // Gateway Port number to send subsequent requests to
}

//Struct for set and get methods where only IP and port are needed
type RegisterGatewayUserParams struct {
	Address string
	Port    string
}

// Used to log gateway state in database.
type ModeAndClock struct {
	Clock int64
	Mode  Mode
}

// You don't see this.
func (this ModeAndClock) GetClock() int64 {
	return this.Clock
}

// To report the state use this struct
type StateInfo struct {
	Clock      int64
	DeviceId   int
	DeviceName Name
	State      State
}

// You don't see this.
func (this StateInfo) GetClock() int64 {
	return this.Clock
}

// Used for no arguments or replies in RPCs
type Empty struct{}

// Sent during synchronization RPCs
type ConsistencyData struct {
	AssignedNodes   map[RegisterGatewayUserParams][]RegisterParams
	Clock           int64
	HomeAway        ModeAndClock
	RegisteredNodes []RegisterParams
	Replica         RegisterGatewayUserParams // data source
	StateInfos      []StateInfo
	User            RegisterGatewayUserParams
}
