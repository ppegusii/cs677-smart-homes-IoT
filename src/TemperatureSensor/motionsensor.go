// Pull based sensor. Report current temperature of the device

// query_state(int device id) results in "current temperature" : Reply has 2 parts : Deviceid and state
// temperature to be maintained should be between 1C and 2C for the SmartOutlet.

package main

import (
	"fmt"
//	"flag"
	"net"
	"net/rpc"
	"os"
	"log"
)

type struct temperatureSensor

func newtempSensor(temperature float64,address string, port string) *temperatureSensor {
	return &temperatureSensor{
		devType : Sensor,
		name : Temperature,
		currTemp : temperature,
		deviceID : -1, // Device ID -1 implies device is unregistered	
		port : port,
		address : address,
	}
}

func incrTemp(ts *TemperatureSensor) float64 {
    ts.currTemp +=0.5
    return ts.currTemp
}

func decrTemp(ts *TemperatureSensor) float64 {
    ts.currTemp -=0.5
    return ts.currTemp
}

//This function gets own IP address to send it to the gateway

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

// Return the current temperature to Gateway
func (ts *temperatureSensor) QueryState(params *StateRequest, reply *StateResponse) error {
	reply.deviceID = ts.deviceID
	reply.state = ts.currTemp
	log.Printf("Returned the current motion state %d of device %d %d to the gateway",*args, m.Deviceid, m.State)
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