package main

import (
	"errors"
	"fmt"
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"github.com/ppegusii/cs677-smart-homes-IoT/structs"
	"log"
	"net"
	"net/rpc"
	"time"
)

type Gateway struct {
	bulbDev         structs.SyncMapIntBool
	bulbTimer       structs.SyncTimer
	database        structs.SyncRegGatewayUserParam
	doorSen         structs.SyncMapIntBool
	ip              string
	mode            structs.SyncMode
	motionSen       structs.SyncMapIntBool
	outletDev       structs.SyncMapIntBool
	outletMode      structs.SyncMode
	pollingInterval int
	port            string
	senAndDev       structs.SyncMapIntRegParam
	tempSen         structs.SyncMapIntBool
	user            structs.SyncRegGatewayUserParam
}

func newGateway(dbIP string, dbPort string, ip string, mode api.Mode, pollingInterval int, port string) *Gateway {
	var g *Gateway = &Gateway{
		bulbDev:         *structs.NewSyncMapIntBool(),
		database:        *structs.NewSyncRegGatewayUserParam(),
		doorSen:         *structs.NewSyncMapIntBool(),
		ip:              ip,
		mode:            *structs.NewSyncMode(mode),
		motionSen:       *structs.NewSyncMapIntBool(),
		outletDev:       *structs.NewSyncMapIntBool(),
		outletMode:      *structs.NewSyncMode(api.OutletsOff),
		pollingInterval: pollingInterval,
		port:            port,
		senAndDev:       *structs.NewSyncMapIntRegParam(),
		tempSen:         *structs.NewSyncMapIntBool(),
		user:            *structs.NewSyncRegGatewayUserParam(),
	}
	g.database.Set(api.RegisterGatewayUserParams{Address: dbIP, Port: dbPort})
	g.bulbTimer = *structs.NewSyncTimer(5*time.Minute, g.turnBulbsOff)
	return g
}

func (g *Gateway) start() {
	//start RPC server
	var err error = rpc.Register(api.GatewayInterface(g))
	if err != nil {
		log.Fatal("rpc.Register error: %s\n", err)
	}
	var listener net.Listener
	listener, err = net.Listen("tcp", g.ip+":"+g.port)
	if err != nil {
		log.Fatal("net.Listen error: %s\n", err)
	}
	logCurrentMode(g.mode.GetMode())
	go rpc.Accept(listener)
	//register with database
	var client *rpc.Client
	var db api.RegisterGatewayUserParams = g.database.Get()
	var empty struct{}
	client, err = rpc.Dial("tcp", db.Address+":"+db.Port)
	if err != nil {
		log.Fatal("error dialing gateway: %+v", err)
	}
	err = client.Call("Database.RegisterGateway", &api.RegisterGatewayUserParams{Address: g.ip, Port: g.port}, &empty)
	if err != nil {
		log.Fatal("error registering with gateway: %+v", err)
	}
	//start polling temperature sensors
	g.pollTempSensors()
}

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
				var client *rpc.Client
				var err error
				client, err = rpc.Dial("tcp", regParams.Address+":"+regParams.Port)
				if err != nil {
					log.Printf("dialing error: %+v", err)
					continue
				}
				err = client.Call("TemperatureSensor.QueryState", &tempId, &tempReply)
				if err != nil {
					log.Printf("calling error: %+v", err)
				}
				log.Printf("Received temp: %d", tempReply.State)
				//Write temperature sensor state to database
				g.writeStateInfo("Database.AddState", &tempReply)
			}
			//update the outlets
			//just using the last tempVal
			var tempVal api.State = tempReply.State
			var s api.State
			var outletState api.Mode = g.outletMode.GetMode()
			if tempVal < 1 && outletState == api.OutletsOff {
				s = api.On
				g.outletMode.SetMode(api.OutletsOn)
			} else if tempVal > 2 && outletState == api.OutletsOn {
				s = api.Off
				g.outletMode.SetMode(api.OutletsOff)
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
			var outletIdRegParams map[int]*api.RegisterParams = *g.senAndDev.GetRegParams(g.outletDev.GetInts())
			if len(outletIdRegParams) != 0 {
				var empty struct{}
				for outletId, regParams := range outletIdRegParams {
					var client *rpc.Client
					var err error
					client, err = rpc.Dial("tcp", regParams.Address+":"+regParams.Port)
					if err != nil {
						log.Printf("dialing error: %+v", err)
						continue
					}
					client.Go("SmartOutlet.ChangeState", api.StateInfo{DeviceId: outletId, State: s}, &empty, nil)
				}
			}
		}
	}
}

func (g *Gateway) RegisterUser(params *api.RegisterGatewayUserParams, _ *struct{}) error {
	log.Printf("Registering user with info: %+v", params)
	g.user.Set(*params)
	return nil
}

func (g *Gateway) Register(params *api.RegisterParams, reply *int) error {
	log.Printf("Attempting to register device with this info: %+v", params)
	var err error = nil
	var id int
	switch params.Type {
	case api.Sensor:
		switch params.Name {
		case api.Door:
			id = g.senAndDev.AddRegParam(params)
			g.doorSen.AddInt(id)
			break
		case api.Motion:
			id = g.senAndDev.AddRegParam(params)
			g.motionSen.AddInt(id)
			break
		case api.Temperature:
			id = g.senAndDev.AddRegParam(params)
			g.tempSen.AddInt(id)
			break
		default:
			err = errors.New(fmt.Sprintf("Invalid Sensor Name: %+v", params.Name))
			break
		}
		break
	case api.Device:
		switch params.Name {
		case api.Bulb:
			id = g.senAndDev.AddRegParam(params)
			g.bulbDev.AddInt(id)
			break
		case api.Outlet:
			id = g.senAndDev.AddRegParam(params)
			g.outletDev.AddInt(id)
			break
		default:
			err = errors.New(fmt.Sprintf("Invalid Device Name: %+v", params.Name))
		}
		break
	default:
		err = errors.New(fmt.Sprintf("Invalid Type: %+v", params.Type))
	}
	*reply = id
	return err
}

func (g *Gateway) ReportMotion(params *api.StateInfo, _ *struct{}) error {
	log.Printf("Received motion report with this info: %+v", params)
	var exists bool = g.motionSen.Exists(params.DeviceId)
	if !exists {
		return errors.New(fmt.Sprintf("Device with following id not motion sensor or not registered: %+v", params.DeviceId))
	}
	switch g.mode.GetMode() {
	case api.Home:
		switch params.State {
		case api.MotionStart:
			var timerActive bool = g.bulbTimer.Stop()
			if !timerActive {
				g.turnBulbsOn()
			}
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

func (g *Gateway) sendText() {
	if !g.user.Exists() {
		return
	}
	var client *rpc.Client
	var err error
	var regUserParams api.RegisterGatewayUserParams = g.user.Get()
	var msg string = "There's something moving in your house!"
	var empty struct{}
	client, err = rpc.Dial("tcp", regUserParams.Address+":"+regUserParams.Port)
	if err != nil {
		log.Printf("dialing error: %+v", err)
		return
	}
	client.Go("User.TextMessage", &msg, &empty, nil)
}

func (g *Gateway) turnBulbsOn() {
	g.changeBulbStates(api.On)
}

func (g *Gateway) turnBulbsOff() {
	g.changeBulbStates(api.Off)
}

func (g *Gateway) changeBulbStates(s api.State) {
	var bulbIdRegParams map[int]*api.RegisterParams = *g.senAndDev.GetRegParams(g.bulbDev.GetInts())
	var empty struct{}
	for bulbId, regParams := range bulbIdRegParams {
		var client *rpc.Client
		var err error
		client, err = rpc.Dial("tcp", regParams.Address+":"+regParams.Port)
		if err != nil {
			log.Printf("dialing error: %+v", err)
			continue
		}
		client.Go("SmartBulb.ChangeState", api.StateInfo{DeviceId: bulbId, State: s}, &empty, nil)
	}
}

func (g *Gateway) ChangeMode(params *api.Mode, _ *struct{}) error {
	log.Printf("Received change mode request with this mode: %+v", *params)
	var err error = nil
	switch *params {
	case api.Home:
		if g.mode.GetMode() == api.Home {
			break
		}
		g.mode.SetMode(*params)
		logCurrentMode(g.mode.GetMode())
		var anyMotion bool = g.checkForMotion()
		if anyMotion {
			g.turnBulbsOn()
		}
		break
	case api.Away:
		if g.mode.GetMode() == api.Away {
			break
		}
		g.mode.SetMode(*params)
		logCurrentMode(g.mode.GetMode())
		g.bulbTimer.Stop()
		g.turnBulbsOff()
		break
	default:
		err = errors.New(fmt.Sprintf("Invalid Mode: %+v", *params))
	}
	return err
}

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

func (g *Gateway) checkForMotion() bool {
	var motionIdRegParams map[int]*api.RegisterParams = *g.senAndDev.GetRegParams(g.motionSen.GetInts())
	if len(motionIdRegParams) != 0 {
		var queryStateParams api.StateInfo
		for motionId, regParams := range motionIdRegParams {
			var client *rpc.Client
			var err error
			client, err = rpc.Dial("tcp", regParams.Address+":"+regParams.Port)
			if err != nil {
				log.Printf("dialing error: %+v", err)
				continue
			}
			err = client.Call("MotionSensor.QueryState", &motionId, &queryStateParams)
			if err != nil {
				log.Printf("calling error: %+v", err)
			}
			log.Printf("Received motion status: %+v", queryStateParams)
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
	//TODO write to database, do interesting happens before analysis to change mode to Home/Away
	return nil
}

//Send state info to the specified rpc on the database
func (g *Gateway) writeStateInfo(rpcName string, stateInfo *api.StateInfo) {
	var client *rpc.Client
	var empty struct{}
	var err error
	var db api.RegisterGatewayUserParams = g.database.Get()
	client, err = rpc.Dial("tcp", db.Address+":"+db.Port)
	if err != nil {
		log.Printf("Error dialing database: %+v", err)
		return
	}
	err = client.Call(rpcName, stateInfo, &empty)
	if err != nil {
		log.Printf("Error calling database: %+v", err)
	}
}
