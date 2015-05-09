// This file declares all the structs and interfaces needed by gateway
package main

import (
	"errors"
	"fmt"
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"github.com/ppegusii/cs677-smart-homes-IoT/structs"
	"github.com/ppegusii/cs677-smart-homes-IoT/util"
	"log"
	//"net"
	//"net/rpc"
	"time"
)

// This struct keeps a track of all the attributes of gateway, reference to peertable, middleware,
// ordering events and types of devices registered in the system
type Gateway struct {
	bulbDev         structs.SyncMapIntBool
	bulbTimer       structs.SyncTimer
	cache           *structs.Cache
	database        structs.SyncRegGatewayUserParam
	doorSen         structs.SyncMapIntBool
	ip              string
	mode            structs.SyncModeClock
	motionSen       structs.SyncMapIntBool
	outletDev       structs.SyncMapIntBool
	outletMode      structs.SyncModeClock
	pollingInterval int
	port            string
	rpcSync         api.RpcSyncInterface
	senAndDev       structs.SyncMapIntRegParam
	tempSen         structs.SyncMapIntBool
	user            structs.SyncRegGatewayUserParam
}

// create and initialize the fields of gateway
func newGateway(dbIP string, dbPort string, ip string, mode api.Mode, pollingInterval int, port string, rpcSync api.RpcSyncInterface, cacheSize int) api.GatewayInterface {
	var g *Gateway = &Gateway{
		bulbDev:  *structs.NewSyncMapIntBool(),
		cache:    structs.NewCache(cacheSize),
		database: *structs.NewSyncRegGatewayUserParam(),
		doorSen:  *structs.NewSyncMapIntBool(),
		ip:       ip,
		mode: *structs.NewSyncModeClock(
			api.ModeAndClock{
				Mode:  mode,
				Clock: util.GetTime(),
			}),
		motionSen: *structs.NewSyncMapIntBool(),
		outletDev: *structs.NewSyncMapIntBool(),
		outletMode: *structs.NewSyncModeClock(
			api.ModeAndClock{
				Mode:  api.OutletsOff,
				Clock: util.GetTime(),
			}),
		pollingInterval: pollingInterval,
		port:            port,
		rpcSync:         rpcSync,
		senAndDev:       *structs.NewSyncMapIntRegParam(),
		tempSen:         *structs.NewSyncMapIntBool(),
		user:            *structs.NewSyncRegGatewayUserParam(),
	}
	g.database.Set(api.RegisterGatewayUserParams{Address: dbIP, Port: dbPort})
	g.bulbTimer = *structs.NewSyncTimer(5*time.Minute, g.turnBulbsOff)
	return g
}

func (g *Gateway) Start() {
	logCurrentMode(g.mode.GetModeClock().Mode)
	//register with database
	var db api.RegisterGatewayUserParams = g.database.Get()
	var empty struct{}
	g.rpcSync.RpcSync(db.Address, db.Port, "Database.RegisterGateway",
		&api.RegisterGatewayUserParams{Address: g.ip, Port: g.port},
		&empty, true)
	//start polling temperature sensors
	g.pollTempSensors()
}

// Poll the temperature sensor every n secs, n is determined by the value of pollingInterval
func (g *Gateway) pollTempSensors() {
	//this function would need changes if there were
	//many temperature sensors
	var ticker *time.Ticker = time.NewTicker(time.Duration(g.pollingInterval) * time.Second)
	for range ticker.C {
		var tempIdRegParams map[int]*api.RegisterParams = *g.senAndDev.GetRegParams(g.tempSen.GetInts())
		if len(tempIdRegParams) != 0 {
			var tempReply api.StateInfo
			for tempId, regParams := range tempIdRegParams {
				//Query temperature sensor
				g.rpcSync.RpcSync(regParams.Address, regParams.Port,
					"TemperatureSensor.QueryState",
					&tempId, &tempReply, false)
				log.Printf("Received temp: %d", tempReply.State)
				//Write temperature sensor state to database
				g.writeStateInfo("Database.AddState", &tempReply)
			}
			g.updateOutlets(tempReply.State)
		}
	}
}

// Based on the temperature reported by the temperature sensor send a notification to the smartoutlet
func (g *Gateway) updateOutlets(tempVal api.State) {
	var s api.State
	var outletState api.Mode = g.outletMode.GetModeClock().Mode
	// Ensure that the outlet is On only if the temperarture is between 1 and 2 else the smartoutlet is Off
	if tempVal < 1 && outletState == api.OutletsOff {
		s = api.On
		g.outletMode.SetModeClock(api.OutletsOn, util.GetTime())
	} else if tempVal > 2 && outletState == api.OutletsOn {
		s = api.Off
		g.outletMode.SetModeClock(api.OutletsOff, util.GetTime())
	} else {
		switch outletState {
		case api.OutletsOff:
			s = api.Off
			break
		case api.OutletsOn:
			s = api.On
			break
		}
	}
	// Dial the outlet and send the state
	var outletIdRegParams map[int]*api.RegisterParams = *g.senAndDev.GetRegParams(g.outletDev.GetInts())
	if len(outletIdRegParams) != 0 {
		for outletId, regParams := range outletIdRegParams {
			var stateInfo api.StateInfo = api.StateInfo{
				DeviceId:   outletId,
				DeviceName: api.Outlet,
				State:      s,
			}
			g.writeStateInfo("Database.AddEvent", &stateInfo)
			var reply api.StateInfo
			g.rpcSync.RpcSync(regParams.Address, regParams.Port,
				"SmartOutlet.ChangeState",
				stateInfo, &reply, false)
			g.writeStateInfo("Database.AddState", &stateInfo)
		}
	}
}

// Register user to the Gateway
func (g *Gateway) RegisterUser(params *api.RegisterGatewayUserParams, _ *struct{}) error {
	log.Printf("Registering user with info: %+v", params)
	g.user.Set(*params)
	return nil
}

// Register new devices and sensors to the gateway.
// External wrapper.
func (g *Gateway) Register(params *api.RegisterParams, reply *api.RegisterReturn) error {
	return g.addNode(params, reply, false)
}

// Internal addition of a node. existing means this node already has an ID.
func (g *Gateway) addNode(params *api.RegisterParams, reply *api.RegisterReturn, existing bool) error {
	if !existing {
		params.Clock = util.GetTime()
	}
	log.Printf("Attempting to register device with this info: %+v", params)
	var err error = nil
	var id int
	if existing {
		id = params.DeviceId
	}
	switch params.Type {
	//Register Sensors
	case api.Sensor:
		switch params.Name {
		case api.Door:
			if !existing {
				id = g.senAndDev.AddNewRegParam(params)
			} else {
				g.senAndDev.AddExistingRegParam(params, id)
			}
			g.doorSen.AddInt(id)
			params.DeviceId = id
			g.writeRegInfo(params)
			break
		case api.Motion:
			if !existing {
				id = g.senAndDev.AddNewRegParam(params)
			} else {
				g.senAndDev.AddExistingRegParam(params, id)
			}
			g.motionSen.AddInt(id)
			params.DeviceId = id
			g.writeRegInfo(params)
			break
		case api.Temperature:
			if !existing {
				id = g.senAndDev.AddNewRegParam(params)
			} else {
				g.senAndDev.AddExistingRegParam(params, id)
			}
			g.tempSen.AddInt(id)
			params.DeviceId = id
			g.writeRegInfo(params)
			break
		default:
			err = errors.New(fmt.Sprintf("Invalid Sensor Name: %+v", params.Name))
			break
		}
		break
	case api.Device:
		//Register Device
		switch params.Name {
		case api.Bulb:
			if !existing {
				id = g.senAndDev.AddNewRegParam(params)
			} else {
				g.senAndDev.AddExistingRegParam(params, id)
			}
			g.bulbDev.AddInt(id)
			params.DeviceId = id
			g.writeRegInfo(params)
			break
		case api.Outlet:
			if !existing {
				id = g.senAndDev.AddNewRegParam(params)
			} else {
				g.senAndDev.AddExistingRegParam(params, id)
			}
			g.outletDev.AddInt(id)
			params.DeviceId = id
			g.writeRegInfo(params)
			break
		default:
			err = errors.New(fmt.Sprintf("Invalid Device Name: %+v", params.Name))
		}
		break
	default:
		err = errors.New(fmt.Sprintf("Invalid Type: %+v", params.Type))
	}
	reply.DeviceId = id
	/* Commenting this out to be sure error occurs if IP and port are not properly assigned in gatewayleader
	reply.Address = g.ip //This is just for testing: Update it based on loadbalancing values
	reply.Port = g.port //This is again for testing: Update it based on loadbalancing values
	*/
	params.DeviceId = id

	return err
}

//Motion sensor is a push based sensor it reports motion to the gateway by ReportMotion() interface
func (g *Gateway) ReportMotion(params *api.StateInfo, _ *struct{}) error {
	log.Printf("Received motion report with this info: %+v", params)
	var exists bool = g.motionSen.Exists(params.DeviceId)
	if !exists {
		return errors.New(fmt.Sprintf("Device with following id not motion sensor or not registered: %+v", params.DeviceId))
	}
	g.writeStateInfo("Database.AddState", params)
	switch g.mode.GetModeClock().Mode {
	case api.Home:
		switch params.State {
		case api.MotionStart:
			g.bulbTimer.Stop()
			g.turnBulbsOn()
			break
		case api.MotionStop:
			g.bulbTimer.Reset()
			break
		}
		break
	case api.Away:
		if params.State == api.MotionStart {
			g.sendText()
		}
		break
	}
	return nil
}

// Function to send text to the user if Mode is set to AWAY and motion detected in the house
func (g *Gateway) sendText() {
	if !g.user.Exists() {
		return
	}
	var regUserParams api.RegisterGatewayUserParams = g.user.Get()
	var msg string = "There's something moving in your house!"
	var empty struct{}
	g.rpcSync.RpcSync(regUserParams.Address, regUserParams.Port,
		"User.TextMessage",
		&msg, &empty, false)
}

//Change the bulb state to On
func (g *Gateway) turnBulbsOn() {
	g.changeBulbStates(api.On)
}

//Change the bulb state mainted in gateway struct to Off
func (g *Gateway) turnBulbsOff() {
	g.changeBulbStates(api.Off)
}

//Change the state of smartbulb
func (g *Gateway) changeBulbStates(s api.State) {
	var bulbIdRegParams map[int]*api.RegisterParams = *g.senAndDev.GetRegParams(g.bulbDev.GetInts())
	for bulbId, regParams := range bulbIdRegParams {
		var stateInfo api.StateInfo = api.StateInfo{
			DeviceId: bulbId,
			State:    s,
		}
		g.writeStateInfo("Database.AddEvent", &stateInfo)
		var reply api.StateInfo
		g.rpcSync.RpcSync(regParams.Address, regParams.Port,
			"SmartBulb.ChangeState",
			stateInfo, &reply, false)
		g.writeStateInfo("Database.AddState", &stateInfo)
	}
}

// Change the mode of the System to Home or Away
// Was called by the user process in lab 1
func (g *Gateway) ChangeMode(params *api.Mode, _ *struct{}) error {
	log.Printf("Received change mode request with this mode: %+v", *params)
	var err error = nil
	switch *params {
	case api.Home:
		if g.mode.GetModeClock().Mode == api.Home {
			break
		}
		g.mode.SetModeClock(*params, util.GetTime())
		logCurrentMode(g.mode.GetModeClock().Mode)
		var anyMotion bool = g.checkForMotion()
		if anyMotion {
			g.turnBulbsOn()
		}
		break
	case api.Away:
		if g.mode.GetModeClock().Mode == api.Away {
			break
		}
		g.mode.SetModeClock(*params, util.GetTime())
		logCurrentMode(g.mode.GetModeClock().Mode)
		g.bulbTimer.Stop()
		g.turnBulbsOff()
		break
	default:
		err = errors.New(fmt.Sprintf("Invalid Mode: %+v", *params))
	}
	return err
}

//Get the current mode and print it on the console
func logCurrentMode(m api.Mode) {
	var text string
	switch m {
	case api.Home:
		text = "Home"
		break
	case api.Away:
		text = "Away"
		break
	default:
		text = "Invalid mode"
		break
	}
	log.Printf("Current mode: %s", text)
}

//Query motion sensor for current state
func (g *Gateway) checkForMotion() bool {
	var motionIdRegParams map[int]*api.RegisterParams = *g.senAndDev.GetRegParams(g.motionSen.GetInts())
	if len(motionIdRegParams) != 0 {
		var queryStateParams api.StateInfo
		for motionId, regParams := range motionIdRegParams {
			g.rpcSync.RpcSync(regParams.Address, regParams.Port,
				"MotionSensor.QueryState",
				&motionId, &queryStateParams, false)
			log.Printf("Received motion status: %+v", queryStateParams)
			g.writeStateInfo("Database.AddState", &queryStateParams)
			if queryStateParams.State == api.MotionStart {
				return true
			}
		}
	}
	return false
}

//Receives state changes from door sensors.
//Analyzes the happens before relationship between indoor motion sensing and door opening to infer occupancy.
//Based on the occupancy state, it changes the mode of the gateway to Home or Away.
func (g *Gateway) ReportDoorState(params *api.StateInfo, _ *struct{}) error {
	log.Printf("Received door state info: %+v", params)
	g.writeStateInfo("Database.AddState", params)
	//no evaluation if door state is closed
	if params.State == api.Closed {
		return nil
	}
	//get motion sensor ids
	var motionIds *map[int]bool = g.motionSen.GetInts()
	if len(*motionIds) < 1 {
		return nil
	}
	//assume 1 motionsensor get its id
	var motionId int
	for id, _ := range *motionIds {
		motionId = id
		break
	}
	//get motion states that happened before and after this door state
	var stateInfo = api.StateInfo{
		Clock:    params.Clock,
		DeviceId: motionId,
	}
	var before api.StateInfo
	var db api.RegisterGatewayUserParams = g.database.Get()
	g.rpcSync.RpcSync(db.Address, db.Port,
		"Database.GetHappensBefore",
		stateInfo, &before, false)
	//if motion happens before door opening
	//change mode to away
	var empty struct{}
	var newMode api.Mode
	log.Printf("before = %+v\n", before)
	if before.State == api.MotionStart && g.mode.GetModeClock().Mode != api.Away {
		newMode = api.Away
		g.ChangeMode(&newMode, &empty)
		g.writeMode(api.ModeAndClock{
			Clock: params.Clock,
			Mode:  api.Away,
		})
	}
	//if no motion happens before door opening
	//change mode to home
	if before.State == api.MotionStop && g.mode.GetModeClock().Mode != api.Home {
		newMode = api.Home
		g.ChangeMode(&newMode, &empty)
		g.writeMode(api.ModeAndClock{
			Clock: params.Clock,
			Mode:  api.Home,
		})
	}
	return nil
}

func (g *Gateway) writeMode(m api.ModeAndClock) {
	var db api.RegisterGatewayUserParams = g.database.Get()
	var empty struct{}
	g.rpcSync.RpcSync(db.Address, db.Port,
		"Database.LogMode",
		m, &empty, false)
}

//Send state info to the specified rpc on the database
func (g *Gateway) writeStateInfo(rpcName string, stateInfo *api.StateInfo) {
	if rpcName == "Database.AddState" {
		g.cache.AddEntry(stateInfo)
	}
	var db api.RegisterGatewayUserParams = g.database.Get()
	var empty struct{}
	g.rpcSync.RpcSync(db.Address, db.Port,
		rpcName,
		stateInfo, &empty, false)
}

//Send registration info to the database
func (g *Gateway) writeRegInfo(regInfo *api.RegisterParams) {
	var db api.RegisterGatewayUserParams = g.database.Get()
	var empty struct{}
	g.rpcSync.RpcSync(db.Address, db.Port,
		"Database.AddDeviceOrSensor",
		regInfo, &empty, false)
}

//TODO Remove the following function
//RPC function used by other devices to acknowledge the registration.
//The execution of this function results in initiation of Send Peertable call from
//the middleware of gateway to other devices
func (g *Gateway) RegisterAck(id int, _ *struct{}) error {
	log.Printf("Received Registration Acknowledgement from device: %d", id)
	return nil
}

//Query RPC used for testing
func (g *Gateway) Query(params api.Name, _ *struct{}) error {
	var err error = nil
	switch params {
	case api.Bulb:
		err = errors.New("Quering bulb state not implemented")
		break
	case api.Door:
		err = errors.New("Quering door state not implemented")
		break
	case api.Motion:
		go g.checkForMotion()
		break
	case api.Outlet:
		err = errors.New("Quering outlet state not implemented")
		break
	case api.Temperature:
		go g.pollTempSensors()
		break
	default:
		err = errors.New(fmt.Sprintf("Invalid name: %d", params))
		break
	}
	return err
}

// Push date necessary for consistency
func (g *Gateway) PushData(data *api.ConsistencyData, _ *api.Empty) error {
	// do stuff with RegisteredNodes just need to add them as existing
	for _, regParam := range data.RegisteredNodes {
		g.addNode(&regParam, &api.RegisterReturn{}, true)
	}
	// do stuff with User only update if user is blank
	if g.user.Get().Address == "" {
		g.user.Set(data.User)
	}
	// do stuff with Mode only update if time stamp is newer
	// and modes differ
	var mc api.ModeAndClock = g.mode.GetModeClock()
	var empty *struct{}
	if mc.Mode != data.HomeAway.Mode &&
		data.HomeAway.Clock > mc.Clock {
		g.ChangeMode(&(data.HomeAway.Mode), empty)
	}
	// TODO write StateInfos to cache
	return nil
}

// Retrieve all date necessary for consistency since a given time
func (g *Gateway) PullData(clock int64, data *api.ConsistencyData) error {
	var db api.RegisterGatewayUserParams = g.database.Get()
	var err error = g.rpcSync.RpcSync(db.Address, db.Port,
		"Database.GetDataSince", clock, data, true)
	if err != nil {
		return err
	}
	// add User
	data.User = g.user.Get()
	// add Mode
	data.HomeAway = g.mode.GetModeClock()
	return nil
}
