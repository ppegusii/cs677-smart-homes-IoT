package main

import (
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"github.com/ppegusii/cs677-smart-homes-IoT/structs"
	"log"
	"net"
	"net/rpc"
)

type Database struct {
	events     *structs.SyncMapIntSyncFile
	gateway    *structs.SyncRegGatewayUserParam
	ip         string
	port       string
	stateCache *structs.SyncMapIntState
	states     *structs.SyncMapIntSyncFile
}

func newDatabase(ip string, port string) *Database {
	return &Database{
		events:     structs.NewSyncMapIntSyncFile(),
		gateway:    structs.NewSyncRegGatewayUserParam(),
		ip:         ip,
		port:       port,
		stateCache: structs.NewSyncMapIntState(),
		states:     structs.NewSyncMapIntSyncFile(),
	}
}

func (g *Database) start() {
	var err error = rpc.Register(api.DatabaseInterface(g))
	if err != nil {
		log.Fatal("rpc.Register error: %s\n", err)
	}
	var listener net.Listener
	listener, err = net.Listen("tcp", g.ip+":"+g.port)
	if err != nil {
		log.Fatal("net.Listen error: %s\n", err)
	}
	rpc.Accept(listener)
}

func (g *Database) AddDeviceOrSensor(params *int, reply *api.RegisterParams) error {
	return nil
}

func (g *Database) AddEvent(params *api.StateInfo, _ *struct{}) error {
	return nil
}

func (g *Database) AddState(params *api.StateInfo, _ *struct{}) error {
	return nil
}

func (g *Database) GetState(params *int, reply *api.StateInfo) error {
	return nil
}

func (g *Database) RegisterGateway(params *api.RegisterGatewayUserParams, _ *struct{}) error {
	return nil
}
