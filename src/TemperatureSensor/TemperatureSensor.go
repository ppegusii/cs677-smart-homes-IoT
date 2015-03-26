package main

import (
	"fmt"
//	"flag"
	"net"
	"net/rpc"
	"os"
	"log"
)

func newtempSensor(temperature float64,address string, port string) *TemperatureSensor {
	return &temperatureSensor{
		Type : Sensor,
		Name : Temperature,
		Temp : temperature,
		Deviceid : -1, // Device ID -1 implies device is unregistered	
		Address : address,
		Port : port,
	}
}

func incrTemp(ts *temperatureSensor) float64 {
    ts.Temp +=0.5
    return ts.Temp
}

func decrTemp(ts *temperatureSensor) float64 {
    ts.Temp -=0.5
    return ts.Temp
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
func (ts *temperatureSensor) QueryState(params *StateRequest, reply *StateTemperature) error {
	reply.Deviceid = ts.Deviceid
	reply.CurrentTemp = ts.Temp
	return nil
}

func (ts *temperatureSensor) listen() {
	var err error = rpc.Register(Interface(ts))
	if err != nil {
		fmt.Printf("rpc.Register error: %s\n", err)
		os.Exit(1)
	}

	var listener net.Listener
	listener, err = net.Listen("tcp", ":"+ts.port)
	if err != nil {
		fmt.Printf("net.Listen error: %s\n", err)
		os.Exit(1)
	}
	rpc.Accept(listener)
}

func (ts *temperatureSensor) start() {

	args := &RegisterParams{ts.devType, ts.name, ts.address, ts.port}

	service := os.Args[1]
	fmt.Printf("Sensor IP and port are %s %s\n", ts.address, ts.port)
	client, err := rpc.Dial("tcp", service)
	if err != nil {
		log.Fatal("dialing:", err)
	}

	var reply int

/* This is the call for registration; populate the deviceID field accordingly */
	err = client.Call("Gateway.Register", args, &reply)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Connection Established with Gateway...")
		fmt.Println("Device id returned from Gateway:", reply, &reply)
		ts.deviceID = reply
	fmt.Printf("Device Type , Device Name, Current Temperature, Device ID, Port number of temperature sensor are as follows : \n")
	fmt.Printf("%d %d %f %d %s\n",ts.devType,ts.name,ts.currTemp,ts.deviceID,ts.port)
	}

	rpc.Register(Interface(ts))
	ln, err := net.Listen("tcp", ":1234")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for {
		c, err := ln.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeConn(c)
	}
}