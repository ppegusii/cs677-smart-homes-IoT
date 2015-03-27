package api

import ()

type Type int

const (
	Sensor Type = iota //0
	Device Type = iota //1
)

type Name int

const (
	Temperature Name = iota //0
	Motion      Name = iota //1
	Bulb        Name = iota //2
	Outlet      Name = iota //3
)

type State int

const (
	On          State = iota //0
	Off         State = iota //1
	MotionStart State = iota //2
	MotionStop  State = iota //3
)

type Mode int

const (
	Home Mode = iota //0
	Away Mode = iota //1
	//These states indicate whether the
	//gateway believes smart outlets are
	//on or off.
	OutletsOn  Mode = iota //2
	OutletsOff Mode = iota //3
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
