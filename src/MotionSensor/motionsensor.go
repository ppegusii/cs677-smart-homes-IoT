// Push based sensor. Push event when it sees motion. State reported is "yes" or "no"

// query_state(int device id) results in "yes" or "no" : Reply has 2 parts : Deviceid and state
// call report_state(int device id, state) when motion to no motion or reverse is detected ; no motion detected for 5 mins.

package main

import (
 "net"
 "net/rpc"
 "log"
// "fmt"
)

func (m *SmartAppliance) Querystate(args *int, reply *SmartAppliance) error {
	if(*args == m.Deviceid) {
		reply.Deviceid = m.Deviceid
		reply.State = m.State
		log.Printf("Returned the current motion state %d of device %d %d to the gateway",*args, m.Deviceid, m.State)
		return nil
	} else {
		log.Printf("Incorrect device ID",m.State)
		reply.Deviceid = m.Deviceid
		reply.State = -1
		return nil // TODO: Error code in gateway to be returned
	}
}

func (m *SmartAppliance) Negatestate(args *SmartAppliance, reply *int) error {
	if(m.State == On) {
		*reply = 0
	} else {
		*reply = 1
	}
	return nil
}

func (m *SmartAppliance) ManualMotion(args *SmartAppliance, reply *int) error {
	if (m.State == args.State) {
		//Device state is same as the new state requested by gateway, so no change in state
		*reply = 0
	} else {
		m.State = args.State
		*reply = 1
		//TODO: Issue a call to report_state(int device id, state) interface in gateway when change in state is observed
	}
	return nil
}
 
func main(){
msensor := new(SmartAppliance)
// Register to get the Deviceid and decide on initial state
msensor.State = MotionStop
msensor.Deviceid = 2
rpc.Register(msensor)

//TODO: Register with Gateway and get the device ID

// Listening string hardcode or input from user
listener, e := net.Listen("tcp", ":2345")
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