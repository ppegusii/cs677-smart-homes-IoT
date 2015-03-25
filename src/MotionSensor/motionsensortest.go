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

type SmartAppliance struct {
	Deviceid int
	State State
}

func main(){
 
client, err := rpc.Dial("tcp", "127.0.0.1:2345")
 if err != nil {
log.Fatal("dialing:", err)
 }

var args int
args = 2
var reply SmartAppliance
err = client.Call("SmartAppliance.Querystate", args, &reply)
 if err != nil {
log.Fatal("Querystate error:", err)
 }
fmt.Printf("Querystate State: %d\n", reply.State)
fmt.Printf("Device ID is %d\n", reply.Deviceid)

var ack int

/* Test for incorrect DeviceID */
nstate := &SmartAppliance{2,MotionStop}
err = client.Call("SmartAppliance.ManualMotion", nstate, &ack)
 if err != nil {
log.Fatal("Changestate error:", err)
 }
fmt.Printf("Changestate State for device is : %d\n", ack)

/* Correct device ID but same state as current state of the motion sensor */
nstate1 := &SmartAppliance{2,MotionStart}
err = client.Call("SmartAppliance.ManualMotion", nstate1, &ack)
 if err != nil {
log.Fatal("Changestate error:", err)
 }
fmt.Printf("Changestate State for device is : %d\n", ack)

/* Correct device ID and  different state than current state of the motion sensor */
nstate2 := &SmartAppliance{2,MotionStop}
err = client.Call("SmartAppliance.ManualMotion", nstate2, &ack)
 if err != nil {
log.Fatal("Changestate error:", err)
 }
fmt.Printf("Changestate State for device is : %d\n", ack)

}

