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
	Querystate(args *int, reply *SmartAppliance) error // args is the deviceid
	Changestate(args *SmartAppliance, reply *int) error // Possible values of reply and its indication are as below:
	/* Value Meaning
		-1 -> The DeviceID in the args send by the gateway was incorrect so no state change has been done
		 0 -> Device ID is correct but the device is already in the state requested by gateway eg: Changestate to Motionstart
		 for a motion device in start state
		 1 -> Device ID is correct and state toggle by new state change
	*/
}