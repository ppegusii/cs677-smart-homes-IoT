package api

import (
//"net"
)

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
	Home       Mode = iota
	Away       Mode = iota
	OutletsOn  Mode = iota
	OutletsOff Mode = iota
)

type GatewayInterface interface {
	Register(params *RegisterParams, reply *int) error
	ReportMotion(params *ReportMotionParams, _ *struct{}) error
	ChangeMode(params *ChangeModeParams, _ *struct{}) error
}

type SensorInterface interface {
	QueryState(params *int, reply *QueryStateParams) error
}

type RegisterParams struct {
	Type Type
	Name Name
	//Cannot get caller IP from rpc library.
	//Might as well send listening port too.
	Address string
	Port    string
	//	ListenSocket net.TCPAddr
}

type ReportMotionParams struct {
	DeviceId int
	State    State
}

type ChangeModeParams struct {
	Mode Mode
}

type ChangeStateParams struct {
	DeviceId int
	State    State
}

type QueryStateParams struct {
	DeviceId int
	State    State
}
