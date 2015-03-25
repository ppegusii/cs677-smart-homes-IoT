// Code to test the smartoutlet
package main
 
import (
 "fmt"
 "net/rpc"
 "log"
)

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

type Motionsensor struct {
	Deviceid int
	state State
}

func main(){
 
client, err := rpc.Dial("tcp", "127.0.0.1:2345")
 if err != nil {
log.Fatal("dialing:", err)
 }

var args int
args = 2
var reply Motionsensor
err = client.Call("Motionsensor.Querystate", args, &reply)
 if err != nil {
log.Fatal("Querystate error:", err)
 }
fmt.Printf("Querystate State: %d\n", reply.state)
fmt.Printf("Device ID is %d\n", reply.Deviceid)

var ack int

/* Test for incorrect DeviceID */
nstate := &Newstate{3,MotionStop}
err = client.Call("Motionsensor.Changestate", nstate, &ack)
 if err != nil {
log.Fatal("Changestate error:", err)
 }
fmt.Printf("Changestate State for device is : %d\n", ack)

/* Correct device ID but same state as current state of the motion sensor */
nstate = &Newstate{2,MotionStart}
err = client.Call("Motionsensor.Changestate", nstate, &ack)
 if err != nil {
log.Fatal("Changestate error:", err)
 }
fmt.Printf("Changestate State for device is : %d\n", ack)

/* Correct device ID and  different state than current state of the motion sensor */
nstate = &Newstate{2,MotionStop}
err = client.Call("Motionsensor.Changestate", nstate, &ack)
 if err != nil {
log.Fatal("Changestate error:", err)
 }
fmt.Printf("Changestate State for device is : %d\n", ack)

}

