package api

import ()

type Type int

const (
	Sensor Type = iota
	Device Type = iota
)

type Name int

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
	Away    Mode = iota
	Home    Mode = iota
	Logical Mode = iota
	//These states indicate whether the
	//gateway believes smart outlets are
	//on or off.
	OutletsOn  Mode = iota
	OutletsOff Mode = iota
	Time       Mode = iota
)

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
}

type OrderingInterface interface {
	NewNodeNotify(params *OrderingNode, _ *struct{}) error
	EventNotify(_ *struct{}, _ *struct{}) error
	StampStateInfo(params *StateInfo, reply *StateInfo) error
}

type SensorInterface interface {
	QueryState(params *int, reply *StateInfo) error
}

type UserInterface interface {
	TextMessage(params *string, _ *struct{}) error
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
	//TODO add clock
	DeviceId int
	State    State
	UnixTime int64
}
