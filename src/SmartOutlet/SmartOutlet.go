package main

import (
 "net"
 "net/rpc"
 "log"
 "fmt"
)

func (t *SmartOutlet) Querystate(args *Newstate, reply *State) error {
	if(args.Deviceid == t.Deviceid){
		*reply = t.state
		return nil
		} else {
			log.Println("Incorrect device ID")
		}
}

// This would be used to manually change state of the device
func (t *SmartOutlet) Manualswitch(args *Newstate, reply *int) error {
	if(args.Deviceid == t.Deviceid){
		if (t.state == args.Nstate) {
		*reply = -1
	} else {
		t.state = args.Nstate
		*reply = 0
	}
	return nil
	} else {
		fmt.Println("Queried an incorrect device type")
		*reply = -1
		return nil
	}
}

/* Possible values of reply and its indication are as below:
  Value  -> Meaning
  =====     =======
   -1   -> The DeviceID in the args send by the gateway was incorrect so no state change has been done
	0    -> Device ID is correct but the device is already in the state requested by gateway eg: Changestate to Motionstart
		    for a motion device in start state
	1    -> Device ID is correct and state toggle by new state change
*/

func (t *SmartOutlet) Changestate(args *Newstate, reply *int) error {
	if(args.Deviceid == t.Deviceid){
		if (t.state == args.Nstate) {
		*reply = -1
	} else {
		t.state = args.Nstate
		*reply = 0
	}
	return nil
	} else {
		fmt.Println("Queried an incorrect device type")
		*reply = -1
		return nil
	}
}
 
func main(){
soutlet := new(SmartOutlet)

//TODO: add the code for registration
soutlet.state = On
soutlet.Deviceid = 3
rpc.Register(soutlet)

// Listening string hardcode or input from user
listener, e := net.Listen("tcp", ":1234")
 if e != nil {
log.Fatal("listen error:", e)
 }
 for {
 if conn, err := listener.Accept(); err != nil {
log.Fatal("accept error: " + err.Error())
 } else {
log.Printf("new connection established\n")
go rpc.ServeConn(conn)
 }
 }
}