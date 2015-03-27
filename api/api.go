package api

import ()

type Type int

const (
	Sensor Type = iota
	Device Type = iota
)

type Name int

const (
	Temperature Name = iota
	Motion      Name = iota
	Bulb        Name = iota
	Outlet      Name = iota
)

type State int

const (
	On          State = iota
	Off         State = iota
	MotionStart State = iota
	MotionStop  State = iota
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

type GatewayInterface interface {
	ChangeMode(params *Mode, _ *struct{}) error
	Register(params *RegisterParams, reply *int) error
	RegisterUser(params *RegisterUserParams, _ *struct{}) error
	ReportMotion(params *ReportMotionParams, _ *struct{}) error
}

type MotionSensorInterface interface {
	QueryState(params *int, reply *QueryStateParams) error
}

type TemperatureSensorInterface interface {
	QueryState(params *int, reply *QueryTemperatureParams) error
}

type DeviceInterface interface {
	QueryState(params *int, reply *QueryStateParams) error
	ChangeState(params *ChangeStateParams, _ *struct{}) error
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

type ReportMotionParams struct {
	DeviceId int
	State    State
}

type ChangeStateParams struct {
	DeviceId int
	State    State
}

type QueryStateParams struct {
	DeviceId int
	State    State
}

type QueryTemperatureParams struct {
	DeviceId    int
	Temperature float64
}
