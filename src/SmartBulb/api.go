package main
 
type State int

const (
	On          State = iota
	Off         State = iota
	MotionStart State = iota
	MotionStop  State = iota
)
 
type SmartAppliance struct {
	Deviceid int
	State State
}

type RegisterParams struct {
	Type Type
	Name Name
	//Cannot get caller IP from rpc library.
	//Might as well send listening port too.
	Address string
	Port string
//	ListenSocket net.TCPAddr
}

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

type Interface interface {
	Changestate(args *SmartAppliance, reply *int) error
	Querystate(args *SmartAppliance, reply *State) error // reply is the ack
}