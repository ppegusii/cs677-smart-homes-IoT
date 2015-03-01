package main

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"sync"
)

type syncInt struct {
	sync.Mutex
	i int
}

type syncMapIntBool struct {
	sync.RWMutex
	m map[int]bool
}

type syncMapIntRegParam struct {
	sync.RWMutex
	m map[int]RegisterParams
}

type syncMode struct {
	sync.RWMutex
	m Mode
}

type Gateway struct {
	mode      syncMode
	nextId    syncInt
	port      string
	pullSen   syncMapIntBool
	senAndDev syncMapIntRegParam
}

func newGateway(mode Mode, port string) *Gateway {
	return &Gateway{
		mode: syncMode{
			m: mode,
		},
		nextId: syncInt{},
		port:   port,
		pullSen: syncMapIntBool{
			m: make(map[int]bool),
		},
		senAndDev: syncMapIntRegParam{
			m: make(map[int]RegisterParams),
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
}

//Unfinished
func (g *Gateway) Register(params *RegisterParams, reply *int) error {
	var err error = nil
	switch params.Type {
	case Sensor:
		break
	case Device:
		break
	default:
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
		g.setMode(params.Mode)
		break
	default:
		err = errors.New(fmt.Sprintf("Invalid Mode: %v", params.Mode))
	}
	return err
}

func (g *Gateway) incAndReturnNextId() int {
	g.nextId.Lock()
	var nextId int = g.nextId.i
	g.nextId.i++
	g.nextId.Unlock()
	return nextId
}

func (g *Gateway) getMode() Mode {
	g.mode.RLock()
	var mode Mode = g.mode.m
	g.mode.RUnlock()
	return mode
}

func (g *Gateway) setMode(mode Mode) {
	g.mode.Lock()
	g.mode.m = mode
	g.mode.Unlock()
	fmt.Printf("Mode changed to: %v", mode)
}
