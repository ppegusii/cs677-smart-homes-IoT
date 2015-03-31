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
	Home Mode = iota
	Away Mode = iota
	//These states indicate whether the
	//gateway believes smart outlets are
	//on or off.
	OutletsOn  Mode = iota
	OutletsOff Mode = iota
)

type DatabaseInterface interface {
	AddDeviceOrSensor(params *int, reply *RegisterParams) error
	AddEvent(params *StateInfo, _ *struct{}) error
	AddState(params *StateInfo, _ *struct{}) error
	GetState(params *int, reply *StateInfo) error
	RegisterGateway(params *RegisterGatewayUserParams, _ *struct{}) error
}

type DeviceInterface interface {
	QueryState(params *int, reply *StateInfo) error
	ChangeState(params *StateInfo, _ *struct{}) error
}

type GatewayInterface interface {
	ChangeMode(params *Mode, _ *struct{}) error
	Register(params *RegisterParams, reply *int) error
	RegisterUser(params *RegisterGatewayUserParams, _ *struct{}) error
	ReportMotion(params *StateInfo, _ *struct{}) error
	ReportDoorState(params *StateInfo, _ *struct{}) error
}

type SensorInterface interface {
	QueryState(params *int, reply *StateInfo) error
}

type UserInterface interface {
	TextMessage(params *string, _ *struct{}) error
}

type RegisterParams struct {
	Type    Type
	Name    Name
	Address string
	Port    string
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
