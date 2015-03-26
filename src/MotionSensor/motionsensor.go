// Push based sensor. Push event when it sees motion. State reported is "yes" or "no"

// query_state(int device id) results in "yes" or "no" : Reply has 2 parts : Deviceid and state
// call report_state(int device id, state) when motion to no motion or reverse is detected ; no motion detected for 5 mins.

package main

import (
"net"
"net/rpc"
"log"
"fmt"
"os"
"flag"
)

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

func (m *SmartAppliance) ManualMotion(args *SmartAppliance, reply *int) error {
	if (m.State == args.State) {
		//Device state is same as the new state requested by gateway, so no change in state
		*reply = 0
	} else {
		m.State = args.State
		*reply = 1
		//Issue a call to report_state(int device id, state) interface in gateway when change in state is observed
// Dial Gateway
	client, err := rpc.Dial("tcp", m.GatewayIP)
	if err != nil {
		log.Fatal("dialing:", err)
	}
// Issue RPC call to push a notification to the Gateway about change in state
	var reply int
	args := &ReportMotionParams{m.Deviceid, m.State}

	err = client.Call("Gateway.ReportMotion", args, &reply)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Connection Established with Gateway...")
		m.Deviceid = reply
	}
	client.Close()
	}
	return nil
}

func NewMotionSensor(state State, address string, port string,service string) *MotionSensor {
	return &MotionSensor {
		Type : Sensor,
		Name : Motion,
		State : state,
		Deviceid : -1, // Device ID -1 implies device is unregistered
		Port : port,
		Address : address,
		GatewayIP : service,
	}
} 

func main(){

	//parse input args
	if len(os.Args) != 2 {
		fmt.Println("Usage: ", os.Args[0], "server:port") //Server and port address of Gateway
		fmt.Println("NOTE: server:port address of the gateway")
		os.Exit(1)
	}

	var port *string = flag.String("p", "2345", "port") //Listening port of the sensor
	flag.Parse()

// Dial Gateway
	service := os.Args[1]
	client, err := rpc.Dial("tcp", service)
	if err != nil {
		log.Fatal("dialing:", err)
	}

	address := getOwnIP()
	state := MotionStart 
	var ms *MotionSensor = NewMotionSensor(state, address, *port,service)

// Register Device
	var reply int
	args := &RegisterParams{ms.Type, ms.Name, ms.Address, ms.Port}

	err = client.Call("Gateway.Register", args, &reply)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Connection Established with Gateway...")
		ms.Deviceid = reply
	}
	client.Close()

	msensor := new(SmartAppliance)
	msensor.State = ms.State
	msensor.Deviceid = ms.Deviceid
	msensor.GatewayIP = ms.GatewayIP

	rpc.Register(msensor)

// Listen for rpc calls from gateway
	listener, e := net.Listen("tcp",":"+(*port))
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