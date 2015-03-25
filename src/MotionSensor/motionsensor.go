// Push based sensor. Push event when it sees motion. State reported is "yes" or "no"

// query_state(int device id) results in "yes" or "no" : Reply has 2 parts : Deviceid and state
// call report_state(int device id, state) when motion to no motion or reverse is detected ; no motion detected for 5 mins.

package main

import (
 "net"
 "net/rpc"
 "log"
 "fmt"
)

func (m *Motionsensor) Querystate(args *int, reply *Motionsensor) error {
	if(*args == m.Deviceid) {
		reply.Deviceid = m.Deviceid
		reply.state = m.state
		log.Printf("Returned the current motion state of %d to the gateway",m.state)
		return nil
	} else {
		log.Printf("Returned the current motion state of %d to the gateway",m.state)
		return nil // TODO: Error code in gateway to be returned
	}
}

func (m *Motionsensor) Negatestate(args *Newstate, reply *int) error {
	if(m.state == On) {
		*reply = 0
	} else {
		*reply = 1
	}
	return nil
}

func (m *Motionsensor) ManualMotion(args *Newstate, reply *int) error {
	if (m.state == args.Nstate) {
		//Device state is same as the new state requested by gateway, so no change in state
		*reply = 0
	} else {
		m.state = args.Nstate
		*reply = 1
		//TODO: Issue a call to report_state(int device id, state) interface in gateway when change in state is observed
	}
	return nil
}
 
func main(){
msensor := new(Motionsensor)
// Register to get the Deviceid and decide on initial state
msensor.state = MotionStart
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