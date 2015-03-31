package main

import (
	"bufio"
	"fmt"
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"log"
	"net"
	"net/rpc"
	"os"
)

type User struct {
	gatewayIp   string
	gatewayPort string
	selfIp      string
	selfPort    string
}

func newUser(gatewayIp string, gatewayPort string, selfIp string, selfPort string) *User {
	return &User{
		gatewayIp:   gatewayIp,
		gatewayPort: gatewayPort,
		selfIp:      selfIp,
		selfPort:    selfPort,
	}
}

func (u *User) start() {
	//RPC server
	var err error = rpc.Register(api.UserInterface(u))
	if err != nil {
		log.Fatal("rpc.Register error: %s\n", err)
	}
	var listener net.Listener
	listener, err = net.Listen("tcp", u.selfIp+":"+u.selfPort)
	if err != nil {
		log.Fatal("net.Listen error: %s\n", err)
	}
	go rpc.Accept(listener)
	//register with gateway
	var client *rpc.Client
	client, err = rpc.Dial("tcp", u.gatewayIp+":"+u.gatewayPort)
	if err != nil {
		log.Fatal("dialing error: %+v", err)
	}
	var empty struct{}
	err = client.Call("Gateway.RegisterUser", &api.RegisterGatewayUserParams{Address: u.selfIp, Port: u.selfPort}, &empty)
	if err != nil {
		log.Fatal("calling error: %+v", err)
	}
	//listen on stdin for user input
	u.getInput()
}

func (u *User) getInput() {
	//http://stackoverflow.com/questions/20895552/how-to-read-input-from-console-line
	reader := bufio.NewReader(os.Stdin)
	var empty struct{}
	var mode api.Mode
	for {
		fmt.Print("Enter (0/1) to change gateway mode (Home/Away): ")
		input, _ := reader.ReadString('\n')
		switch input {
		case "0\n":
			mode = api.Home
			break
		case "1\n":
			mode = api.Away
			break
		default:
			fmt.Println("Invalid input")
			continue
		}
		var client *rpc.Client
		var err error
		client, err = rpc.Dial("tcp", u.gatewayIp+":"+u.gatewayPort)
		if err != nil {
			log.Printf("dialing error: %+v", err)
			continue
		}
		client.Go("Gateway.ChangeMode", &mode, empty, nil)
	}
}

func (u *User) TextMessage(params *string, _ *struct{}) error {
	log.Printf("Received text: %s", *params)
	return nil
}
