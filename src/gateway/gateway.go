package main

import (
	"fmt"
	"net"
	"net/rpc"
)

type Server struct{}

func (this *Server) Register(i int64, reply *int64) error { //*int64
//	var response string = "Hey!"
//	reply = &response
	*reply = i+1
	fmt.Println("Reply is", *reply, reply)
	return nil
}

func hello() {
	fmt.Printf("Hello, world.\n")
}

func server() {
	rpc.Register(new(Server))
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

func main() {
	server()
}