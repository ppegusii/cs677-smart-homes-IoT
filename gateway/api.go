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
)

type Gateway struct {
}

type RegisterParams struct {
	Type Type
	Name Name
}

func (g *Gateway) Register(params *RegisterParams, reply *int) error {
	return nil
}

type ReportStateParams struct {
	DeviceId int
	State    State
}

func (g *Gateway) ReportState(params *ReportStateParams, _ *struct{}) error {
	return nil
}

type ChangeModeParams struct {
	Mode Mode
}

func (g *Gateway) ChangeMode(params *ChangeModeParams, _ *struct{}) error {
	return nil
}
