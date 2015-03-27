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
	ip              string
	mode            structs.SyncMode
	motionSen       structs.SyncMapIntBool
	outletDev       structs.SyncMapIntBool
	outletMode      structs.SyncMode
	pollingInterval int
	port            string
	senAndDev       structs.SyncMapIntRegParam
	tempSen         structs.SyncMapIntBool
	user            structs.SyncRegUserParam
}

func newGateway(ip string, mode api.Mode, pollingInterval int, port string) *Gateway {
	var g *Gateway = &Gateway{
		bulbDev:         *structs.NewSyncMapIntBool(),
		ip:              ip,
		mode:            *structs.NewSyncMode(mode),
		motionSen:       *structs.NewSyncMapIntBool(),
		outletDev:       *structs.NewSyncMapIntBool(),
		outletMode:      *structs.NewSyncMode(api.OutletsOff),
		pollingInterval: pollingInterval,
		port:            port,
		senAndDev:       *structs.NewSyncMapIntRegParam(),
		tempSen:         *structs.NewSyncMapIntBool(),
		user:            *structs.NewSyncRegUserParam(),
	}
	g.bulbTimer = *structs.NewSyncTimer(5*time.Minute, g.turnBulbsOff)
	return g
}

func (g *Gateway) start() {
	var err error = rpc.Register(api.GatewayInterface(g))
	if err != nil {
		log.Fatal("rpc.Register error: %s\n", err)
	}
	var listener net.Listener
	listener, err = net.Listen("tcp", g.ip+":"+g.port)
	if err != nil {
		log.Fatal("net.Listen error: %s\n", err)
	}
	go rpc.Accept(listener)
	g.pollTempSensors()
}

func (g *Gateway) pollTempSensors() {
	//this function would need changes if there were
	//many temperature sensors
	var ticker *time.Ticker = time.NewTicker(time.Duration(g.pollingInterval) * time.Second)
	for range ticker.C {
		var tempIdRegParams map[int]*api.RegisterParams = *g.senAndDev.GetRegParams(g.tempSen.GetInts())
		if len(tempIdRegParams) != 0 {
			var tempReply api.QueryTemperatureParams
			for tempId, regParams := range tempIdRegParams {
				var client *rpc.Client
				var err error
				client, err = rpc.Dial("tcp", regParams.Address+":"+regParams.Port)
				if err != nil {
					log.Printf("dialing error: %+v", err)
				}
				err = client.Call("TemperatureSensor.QueryState", &tempId, &tempReply)
				if err != nil {
					log.Printf("calling error: %+v", err)
				}
				log.Printf("Received temp: %f", tempReply.Temperature)
			}
			//update the outlets
			//just using the last tempVal
			var tempVal float64 = tempReply.Temperature
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
					}
					client.Go("SmartOutlet.ChangeState", api.ChangeStateParams{outletId, s}, empty, nil)
				}
			}
		}
	}
}

func (g *Gateway) RegisterUser(params *api.RegisterUserParams, _ *struct{}) error {
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

func (g *Gateway) ReportMotion(params *api.ReportMotionParams, _ *struct{}) error {
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
		g.sendText()
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
	var regUserParams api.RegisterUserParams = g.user.Get()
	var msg string = "There's something moving in your house!"
	var empty struct{}
	client, err = rpc.Dial("tcp", regUserParams.Address+":"+regUserParams.Port)
	if err != nil {
		log.Printf("dialing error: %+v", err)
	}
	client.Go("User.TextMessage", &msg, empty, nil)
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
		}
		client.Go("SmartBulb.ChangeState", api.ChangeStateParams{bulbId, s}, empty, nil)
	}
}

func (g *Gateway) ChangeMode(params *api.Mode, _ *struct{}) error {
	log.Printf("Received change mode request with this info: %+v", params)
	var err error = nil
	switch *params {
	case api.Home:
		if g.mode.GetMode() == api.Home {
			break
		}
		g.mode.SetMode(*params)
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
		g.bulbTimer.Stop()
		g.turnBulbsOff()
		break
	default:
		err = errors.New(fmt.Sprintf("Invalid Mode: %+v", *params))
	}
	return err
}

func (g *Gateway) checkForMotion() bool {
	var motionIdRegParams map[int]*api.RegisterParams = *g.senAndDev.GetRegParams(g.motionSen.GetInts())
	if len(motionIdRegParams) != 0 {
		var queryStateParams api.QueryStateParams
		for motionId, regParams := range motionIdRegParams {
			var client *rpc.Client
			var err error
			client, err = rpc.Dial("tcp", regParams.Address+":"+regParams.Port)
			if err != nil {
				log.Printf("dialing error: %+v", err)
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
