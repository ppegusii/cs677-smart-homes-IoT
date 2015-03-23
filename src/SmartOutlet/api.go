package main
 
type State int

const (
	On          State = iota
	Off         State = iota
	MotionStart State = iota
	MotionStop  State = iota
)

type Newstate struct {
	Deviceid int
	Nstate State
}
 
type SmartOutlet struct {
	Deviceid int
	state State
}