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
	AddDeviceOrSensor(params *int, reply *QueryStateParams) error
	GetState(params *int, reply *CurrentState) error
	AddState(params *CurrentState, _ *struct{}) error
	AddEvent(params *CurrentState, _ *struct{}) error
}

type DeviceInterface interface {
	QueryState(params *int, reply *QueryStateParams) error
	ChangeState(params *ChangeStateParams, _ *struct{}) error
}

type GatewayInterface interface {
	ChangeMode(params *Mode, _ *struct{}) error
	Register(params *RegisterParams, reply *int) error
	RegisterUser(params *RegisterUserParams, _ *struct{}) error
	ReportMotion(params *ReportStateParams, _ *struct{}) error
	ReportDoorState(params *ReportStateParams, _ *struct{}) error
}

type SensorInterface interface {
	QueryState(params *int, reply *QueryStateParams) error
}

type TemperatureSensorInterface interface {
	QueryState(params *int, reply *QueryTemperatureParams) error
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

type RegisterUserParams struct {
	Address string
	Port    string
}

type CurrentState struct {
	//TODO add clock
	DeviceId int
	State    State
	UnixTime int64
}

//To be replaced by CurrentState
type ReportStateParams struct {
	DeviceId int
	State    State
}

//To be replaced by CurrentState
type ChangeStateParams struct {
	DeviceId int
	State    State
}

//To be replaced by CurrentState
type QueryStateParams struct {
	DeviceId int
	State    State
}

//Maybe be replaced by CurrentState
type QueryTemperatureParams struct {
	DeviceId    int
	Temperature float64
}
