//This file defines all the structs, data types and interfaces used by the database

package main

import (
	"errors"
	"fmt"
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"github.com/ppegusii/cs677-smart-homes-IoT/ordermw"
	"github.com/ppegusii/cs677-smart-homes-IoT/structs"
	"github.com/ppegusii/cs677-smart-homes-IoT/util"
	"log"
	"net"
	"net/rpc"
	"strconv"
)

// This struct contains all the attributes of the database and information needed for
// ordering for clock synchronization, peer table to keep a track of ip of the peers and reference to its middleware
type Database struct {
	devSen      *structs.SyncFile
	events      *structs.SyncMapIntSyncFile
	gateway     *structs.SyncRegGatewayUserParam
	gwMode      *structs.SyncFile
	ip          string
	orderMW     api.OrderingMiddlewareInterface
	port        string
	stateCaches *structs.SyncMapIntSyncLatestStateInfos
	states      *structs.SyncMapIntSyncFile
}

// initialize a new database
func newDatabase(ip string, port string, ordering api.Ordering) *Database {
	return &Database{
		events:      structs.NewSyncMapIntSyncFile(),
		gateway:     structs.NewSyncRegGatewayUserParam(),
		ip:          ip,
		orderMW:     ordermw.GetOrderingMiddleware(ordering, int(api.DatabaseOID), ip, port),
		port:        port,
		stateCaches: structs.NewSyncMapIntSyncLatestStateInfos(2),
		states:      structs.NewSyncMapIntSyncFile(),
	}
}

func (d *Database) start() {
	var err error
	//create file for device and sensor info
	d.devSen, err = structs.NewSyncFile("dev_sen.tbl")
	if err != nil {
		log.Fatal("Error creating devSen file: %s\n", err)
	}
	//create file for gateway mode
	d.gwMode, err = structs.NewSyncFile("gw_modes.tbl")
	if err != nil {
		log.Fatal("Error creating gwMode file: %s\n", err)
	}
	//start RPC server
	err = rpc.Register(api.DatabaseInterface(d))
	if err != nil {
		log.Fatal("rpc.Register error: %s\n", err)
	}
	var listener net.Listener
	listener, err = net.Listen("tcp", d.ip+":"+d.port)
	if err != nil {
		log.Fatal("net.Listen error: %s\n", err)
	}
	rpc.Accept(listener)
}

//Writes object information to table.
//Creates tables to track object states and events.
func (d *Database) LogMode(params api.ModeAndClock, _ *struct{}) error {
	var err error = nil
	var modeStr string = ""
	if params.Mode == api.Away {
		modeStr = "away"
	}
	if params.Mode == api.Home {
		modeStr = "home"
	}
	//Write object information to table.
	_, err = d.gwMode.WriteString(fmt.Sprintf("%d,%s\n",
		params.Clock,
		modeStr))
	return err
}

//Writes object information to table.
//Creates tables to track object states and events.
func (d *Database) AddDeviceOrSensor(params *api.RegisterParams, _ *struct{}) error {
	var err error
	//Writes object information to table.
	/*
		_, err = d.devSen.WriteString(fmt.Sprintf("%d,%s,%s,%s,%s\n",
			params.DeviceId,
			util.TypeToString(params.Type),
			util.NameToString(params.Name),
			params.Address,
			params.Port))
	*/
	err = d.devSen.WriteJson(*params)
	if err != nil {
		return err
	}
	//Creates tables to track object states and events.
	var f *structs.SyncFile
	f, err = structs.NewSyncFile(fmt.Sprintf("%d_%s_events.tbl",
		params.DeviceId,
		util.NameToString(params.Name)))
	if err != nil {
		return err
	}
	d.events.Set(params.DeviceId, f)
	f, err = structs.NewSyncFile(fmt.Sprintf("%d_%s_states.tbl",
		params.DeviceId,
		util.NameToString(params.Name)))
	if err != nil {
		return err
	}
	d.states.Set(params.DeviceId, f)
	return nil
}

//Write event to table
func (d *Database) AddEvent(params *api.StateInfo, _ *struct{}) error {
	f, ok := d.events.Get(params.DeviceId)
	if !ok {
		return errors.New(fmt.Sprintf("Invalid device ID: %d", params.DeviceId))
	}
	_, err := d.writeStateInfo(params, f)
	return err
}

//Write state to table
func (d *Database) AddState(params *api.StateInfo, _ *struct{}) error {
	f, ok := d.states.Get(params.DeviceId)
	if !ok {
		return errors.New(fmt.Sprintf("Invalid device ID: %d", params.DeviceId))
	}
	_, err := d.writeStateInfo(params, f)
	d.stateCaches.Get(params.DeviceId).AddStateInfo(*params)
	return err
}

//Given a clock value and device id, return the cached state info that happened just before
func (d *Database) GetHappensBefore(params api.StateInfo, reply *api.StateInfo) error {
	var before, _ *api.StateInfo = d.stateCaches.Get(params.DeviceId).GetBeforeAndAfter(params.Clock)
	log.Printf("before = %+v\n", before)
	if before == nil {
		return nil
	}
	*reply = *before
	return nil
}

// Register to the gateway
func (d *Database) RegisterGateway(params *api.RegisterGatewayUserParams, _ *struct{}) error {
	d.gateway.Set(*params)
	return nil
}

func (d *Database) writeStateInfo(stateInfo *api.StateInfo, f *structs.SyncFile) (int, error) {
	var line string
	var i int
	var err error
	var stateStr string
	if stateInfo.DeviceName == api.Temperature {
		stateStr = strconv.Itoa(int(stateInfo.State))
	} else {
		stateStr = util.StateToString(stateInfo.State)
	}
	line = fmt.Sprintf("%d,%d,%s,%s\n",
		stateInfo.Clock,
		stateInfo.DeviceId,
		util.NameToString(stateInfo.DeviceName),
		stateStr)
	i, err = f.WriteString(line)
	return i, err
}
