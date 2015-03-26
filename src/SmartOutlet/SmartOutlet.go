package main

import (
 "net"
 "net/rpc"
 "log"
 "fmt"
)

func (t *SmartAppliance) Querystate(args *SmartAppliance, reply *State) error {
	if(args.Deviceid == t.Deviceid){
		*reply = t.State
		} else {
			log.Println("Incorrect device ID")
		}
				return nil
}

// This would be used to manually change state of the device
func (t *SmartAppliance) Manualswitch(args *SmartAppliance, reply *int) error {
	if(args.Deviceid == t.Deviceid){
		if (t.State == args.State) {
		*reply = -1
	} else {
		t.State = args.State
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

func (t *SmartAppliance) Changestate(args *SmartAppliance, reply *int) error {
	if(args.Deviceid == t.Deviceid){
		if (t.State == args.State) {
		*reply = 0
	} else {
		oldstate := t. State
		t.State = args.State
		fmt.Println("State change from %d %d", oldstate, t.State)
		*reply = 1
	}
	return nil
	} else {
		fmt.Println("Queried an incorrect device type")
		*reply = -1
		return nil
	}
}
 
func main(){
	soutlet := new(SmartAppliance)

//TODO: add the code for registration
	soutlet.State = Off
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