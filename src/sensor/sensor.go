package main

import (
	"fmt"
	"net"
	"net/rpc"
)

type TemperatureSensor struct {
	sdtype string
	name string
	temp float64
	id int
}

type RegisterParams struct {
	SDType string
	Name string
}

func sensor() {
	/* We are establishing a connection with the gateway */
	c, err := rpc.Dial("tcp", "127.0.0.1:9999")
	if err != nil {
		fmt.Println(err)
		return
	}
	var result int64

	/* Creating an instance of the temperature sensor */
	tSensor := TemperatureSensor{"sensor","temperature",5.90,-1}

	/* The sensor has just joined the system, so it needs to register.
	For registering , we call the rpc Server.Register with the 2 
	parameters type and name and get a unique device id in return. */

	args := &RegisterParams{tSensor.sdtype, tSensor.name}

	err = c.Call("Gateway.Register", args, &result)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Yea... I established a connection")
		fmt.Println("Device id returned from Gateway: ", result, &result)
	}
	
	pullrequest()
}

func (this *TemperatureSensor) QueryState(device_id int, reply *float64) error {
	*reply = 2
	fmt.Println("Reply is", *reply, reply)
	return nil
}

func pullrequest(){
	rpc.Register(new(TemperatureSensor))
	ln, err := net.Listen("tcp", ":9000")
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		c, err := ln.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeConn(c)
	}
}

func main() {
	sensor()
}