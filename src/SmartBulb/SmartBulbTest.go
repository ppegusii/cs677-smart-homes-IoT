// Code to test the smartoutlet
package main
 
import (
 "fmt"
 "net/rpc"
 "log"
)

func main(){
 
client, err := rpc.Dial("tcp", "127.0.0.1:3456")
 if err != nil {
log.Fatal("dialing:", err)
 }

args := &Newstate{3,Off}
var reply State
err = client.Call("SmartAppliance.Querystate", args, &reply)
 if err != nil {
log.Fatal("Querystate error:", err)
 }
fmt.Printf("Querystate State: %d\n", reply)

nstate := &Newstate{4,Off}
err = client.Call("SmartAppliance.Changestate", nstate, &reply)
 if err != nil {
log.Fatal("Changestate error:", err)
 }
fmt.Printf("Changestate State: %d\n", reply)

nstate = &Newstate{4, On}
err = client.Call("SmartAppliance.Changestate", nstate, &reply)
 if err != nil {
log.Fatal("Changestate error:", err)
 }
fmt.Printf("Changestate State: %d\n", reply)

nstate = &Newstate{4, Off}
err = client.Call("SmartAppliance.Changestate", nstate, &reply)
 if err != nil {
log.Fatal("Changestate error:", err)
 }
fmt.Printf("Changestate State: %d\n", reply)
}

