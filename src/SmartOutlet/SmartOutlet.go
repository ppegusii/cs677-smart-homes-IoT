package main

import (
 "net"
 "net/rpc"
 "log"
 "fmt"
 "os"
 "flag"
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

func NewSmartOutlet(state State, address string, port string) *SmartOutlet {
	return &SmartOutlet {
		Type : Device,
		Name : Outlet,
		State : state,
		Deviceid : -1, // Device ID -1 implies device is unregistered
		Port : port,
		Address : address,
	}
}

func getOwnIP() string{     
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, address := range addrs {
    	// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return (ipnet.IP.String())
			}
		}
	}
	return "" //To exit or continue?
}
 
func main(){
	//parse input args
	if len(os.Args) != 2 {
		fmt.Println("Usage: ", os.Args[0], "server:port") //Server and port address of Gateway
		fmt.Println("NOTE: server:port address of the gateway")
		os.Exit(1)
	}

	var port *string = flag.String("p", "1234", "port") //Listening port of the sensor
	flag.Parse()

// Dial Gateway
	service := os.Args[1]
	client, err := rpc.Dial("tcp", service)
	if err != nil {
		log.Fatal("dialing:", err)
	}

	address := getOwnIP()
	state := On 
	var so *SmartOutlet = NewSmartOutlet(state, address, *port)

// Register Device
	var reply int
	args := &RegisterParams{so.Type, so.Name, so.Address, so.Port}

	err = client.Call("Gateway.Register", args, &reply)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Connection Established with Gateway...")
		so.Deviceid = reply
	}
	client.Close()

	soutlet := new(SmartAppliance)
	soutlet.State = so.State
	soutlet.Deviceid = so.Deviceid
	rpc.Register(soutlet)

// Listening string hardcode or input from user
	listener, e := net.Listen("tcp", ":"+(*port))
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