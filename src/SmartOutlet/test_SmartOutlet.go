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

func main(){
 
client, err := rpc.Dial("tcp", "127.0.0.1:1234")
 if err != nil {
log.Fatal("dialing:", err)
 }

args := &Newstate{3,Off}
var reply State
err = client.Call("SmartOutlet.Querystate", args, &reply)
 if err != nil {
log.Fatal("Querystate error:", err)
 }
fmt.Printf("Querystate State: %d\n", reply)

nstate := &Newstate{3,On}
err = client.Call("SmartOutlet.Changestate", nstate, &reply)
 if err != nil {
log.Fatal("Changestate error:", err)
 }
fmt.Printf("Changestate State: %d\n", reply)

nstate = &Newstate{3, Off}
err = client.Call("SmartOutlet.Changestate", nstate, &reply)
 if err != nil {
log.Fatal("Changestate error:", err)
 }
fmt.Printf("Changestate State: %d\n", reply)

nstate = &Newstate{3, On}
err = client.Call("SmartOutlet.Changestate", nstate, &reply)
 if err != nil {
log.Fatal("Changestate error:", err)
 }
fmt.Printf("Changestate State: %d\n", reply)
}

