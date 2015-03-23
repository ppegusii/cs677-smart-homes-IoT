package main

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"os"
)

type Gateway struct {
	bulbDev   syncMapIntBool
	mode      syncMode
	motionSen syncMapIntBool
	outletDev syncMapIntBool
	port      string
	senAndDev syncMapIntRegParam
	tempSen   syncMapIntBool
}

func newGateway(mode Mode, port string) *Gateway {
	return &Gateway{
		bulbDev: syncMapIntBool{
			m: make(map[int]bool),
		},
		mode: syncMode{
			m: mode,
		},
		motionSen: syncMapIntBool{
			m: make(map[int]bool),
		},
		outletDev: syncMapIntBool{
			m: make(map[int]bool),
		},
		port: port,
		senAndDev: syncMapIntRegParam{
			m: make(map[int]*RegisterParams),
		},
		tempSen: syncMapIntBool{
			m: make(map[int]bool),
		},
	}
}

func (g *Gateway) start() {
	var err error = rpc.Register(Interface(g))
	if err != nil {
		fmt.Printf("rpc.Register error: %s\n", err)
		os.Exit(1)
	}
	var listener net.Listener
	listener, err = net.Listen("tcp", ":"+g.port)
	if err != nil {
		fmt.Printf("net.Listen error: %s\n", err)
		os.Exit(1)
	}
	rpc.Accept(listener)
	go 	g.pollTempSensors()
}

func (g *Gateway) pollTempSensors() {
	args := &RegisterParams{0}
	fmt.Println("Connecting to Sensor")
	client, err := rpc.Dial("tcp", "127.0.0.1:1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	var reply *StateResponse

/* This is the call for registration populate the deviceID field accordingly */
	err = client.Call("temperatureSensor.QueryState", args, &reply)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Connection Established with Temperature Sensor...")
		fmt.Println("Temperarture returned from sensor is:", reply.state, &reply)
	}
}

func (g *Gateway) Register(params *RegisterParams, reply *int) error {
	var err error = nil
	var id int
	switch params.Type {
	case Sensor:
		switch params.Name {
		case Motion:
			id = g.senAndDev.addRegParam(params)
			g.motionSen.addInt(id)
			break
		case Temperature:
			id = g.senAndDev.addRegParam(params)
			g.tempSen.addInt(id)
			break
		default:
			err = errors.New(fmt.Sprintf("Invalid Sensor Name: %v", params.Name))
			break
		}
		break
	case Device:
		switch params.Name {
		case Bulb:
			id = g.senAndDev.addRegParam(params)
			g.bulbDev.addInt(id)
			break
		case Outlet:
			id = g.senAndDev.addRegParam(params)
			g.outletDev.addInt(id)
			break
		default:
			err = errors.New(fmt.Sprintf("Invalid Device Name: %v", params.Name))
		}
		break
	default:
		err = errors.New(fmt.Sprintf("Invalid Type: %v", params.Type))
	}
	return err
}

func (g *Gateway) ReportState(params *ReportStateParams, _ *struct{}) error {
	return nil
}

func (g *Gateway) ChangeMode(params *ChangeModeParams, _ *struct{}) error {
	var err error = nil
	switch params.Mode {
	case Home:
	case Away:
		g.mode.setMode(params.Mode)
		break
	default:
		err = errors.New(fmt.Sprintf("Invalid Mode: %v", params.Mode))
	}
	return err
}
