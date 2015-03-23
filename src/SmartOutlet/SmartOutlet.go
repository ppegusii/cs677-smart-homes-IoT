package main

import (
 "net"
 "net/rpc"
 "log"
 "fmt"
)

func (t *SmartOutlet) Querystate(args *Newstate, reply *State) error {
	*reply = t.state
	return nil
}

func (t *SmartOutlet) Negatestate(args *Newstate, reply *int) error {
	if(t.state == On) {
		*reply = 0
	} else {
		*reply = 1
	}
	return nil
}

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
cal := new(SmartOutlet)
// Register to get the Deviceid and decide on initial state
cal.state = On
cal.Deviceid = 3
rpc.Register(cal)
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