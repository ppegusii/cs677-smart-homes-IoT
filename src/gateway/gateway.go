/* This is the dummy Gateway code used to test the sensor code */
package main

import (
	"fmt"
	"net"
	"net/rpc"
)

type Gateway struct{}

type RegisterParams struct {
	SDType string
	Name string
}

func (this *Gateway) Register(args *RegisterParams, reply *int64) error { //*int64
//	var response string = "Hey!"
//	reply = &response
	*reply = 1
	fmt.Println("Reply is", *reply, reply)
	return nil
}

func gateway() {
	rpc.Register(new(Gateway))
	ln, err := net.Listen("tcp", ":9999")
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

/* dummy snippet to test Temperature returned from Temperature Sensor : pull based */

func pulltemp() {
	c, err := rpc.Dial("tcp", "127.0.0.1:9000")
	if err != nil {
		fmt.Println(err)
		return
	}
	var result int64

		pulltemp()

	err = c.Call("TemperatureSensor.QueryState", int(1), &result)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Value of temperature returned by Temperature Sensor is: ", result, &result)
	}
}

func main() {
	gateway()
}