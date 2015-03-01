package gateway

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
)

type Interface interface {
	Register(params *RegisterParams, reply *int) error
	ReportState(params *ReportStateParams, _ *struct{}) error
	ChangeMode(params *ChangeModeParams, _ *struct{}) error
}

type RegisterParams struct {
	Type Type
	Name Name
}

type ReportStateParams struct {
	DeviceId int
	State    State
}

type ChangeModeParams struct {
	Mode Mode
}