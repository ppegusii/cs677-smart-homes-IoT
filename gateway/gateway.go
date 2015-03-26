package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"time"
)

type Gateway struct {
	bulbDev         syncMapIntBool
	bulbTimer       syncTimer
	mode            syncMode
	motionSen       syncMapIntBool
	outletDev       syncMapIntBool
	outletMode      syncMode
	pollingInterval int
	port            string
	senAndDev       syncMapIntRegParam
	tempSen         syncMapIntBool
}

func newGateway(mode Mode, pollingInterval int, port string) *Gateway {
	var g *Gateway = &Gateway{
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
		outletMode: syncMode{
			m: OutletsOff,
		},
		pollingInterval: pollingInterval,
		port:            port,
		senAndDev: syncMapIntRegParam{
			m: make(map[int]*RegisterParams),
		},
		tempSen: syncMapIntBool{
			m: make(map[int]bool),
		},
	}
	g.bulbTimer = *newSyncTimer(5*time.Minute, g.turnBulbsOff)
	return g
}

func (g *Gateway) start() {
	var err error = rpc.Register(Interface(g))
	if err != nil {
		log.Printf("rpc.Register error: %s\n", err)
		os.Exit(1)
	}
	var listener net.Listener
	listener, err = net.Listen("tcp", ":"+g.port)
	if err != nil {
		log.Printf("net.Listen error: %s\n", err)
		os.Exit(1)
	}
	rpc.Accept(listener)
	g.pollTempSensors()
}

func (g *Gateway) pollTempSensors() {
	//this function would need changes if there were
	//many temperature sensors
	var ticker *time.Ticker = time.NewTicker(time.Duration(g.pollingInterval) * time.Second)
	for range ticker.C {
		var tempIdRegParams map[int]*RegisterParams = *g.senAndDev.getRegParams(g.tempSen.getInts())
		if len(tempIdRegParams) != 0 {
			var tempVal float64 = 0
			for tempId, regParams := range tempIdRegParams {
				var client *rpc.Client
				var err error
				client, err = rpc.Dial("tcp", regParams.Address+":"+regParams.Port)
				if err != nil {
					log.Printf("dialing error: %v", err)
				}
				err = client.Call("TemperatureSensor.QueryState", &tempId, &tempVal)
				if err != nil {
					log.Printf("calling error: %v", err)
				}
			}
			//just using the last tempVal
			var s State
			var outletState Mode = g.outletMode.getMode()
			if tempVal < 1 && outletState == OutletsOff {
				s = On
				g.outletMode.setMode(OutletsOn)
			} else if tempVal > 2 && outletState == OutletsOn {
				s = Off
				g.outletMode.setMode(OutletsOff)
			} else {
				switch outletState {
				case OutletsOff:
					s = Off
					break
				case OutletsOn:
					s = On
					break
				}
			}
			var outletIdRegParams map[int]*RegisterParams = *g.senAndDev.getRegParams(g.outletDev.getInts())
			if len(outletIdRegParams) != 0 {
				var empty struct{}
				for outletId, regParams := range outletIdRegParams {
					var client *rpc.Client
					var err error
					client, err = rpc.Dial("tcp", regParams.Address+":"+regParams.Port)
					if err != nil {
						log.Printf("dialing error: %v", err)
					}
					client.Go("SmartOutlet.ChangeState", ChangeStateParams{outletId, s}, empty, nil)
					log.Printf("calling error: %v", err)
				}
			}
		}
	}
}

/*
//I commented out this function because of compilation issues.
func (g *Gateway) pollTempSensors() {
	args := &RegisterParams{0}
	fmt.Println("Connecting to Sensor")
	client, err := rpc.Dial("tcp", "127.0.0.1:1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	var reply *StateResponse

//This is the call for registration populate the deviceID field accordingly
	err = client.Call("temperatureSensor.QueryState", args, &reply)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Connection Established with Temperature Sensor...")
		fmt.Println("Temperarture returned from sensor is:", reply.state, &reply)
	}
}
*/

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

func (g *Gateway) ReportMotion(params *ReportMotionParams, _ *struct{}) error {
	//only expecting motion sensor
	var exists bool = g.motionSen.exists(params.DeviceId)
	if !exists {
		return errors.New(fmt.Sprintf("Device with following id not motion sensor or not registered: %v", params.DeviceId))
	}
	switch g.mode.getMode() {
	case Home:
		g.turnBulbsOn()
		break
	case Away:
		//TODO g.sendText()
		break
	}
	return nil
}

func (g *Gateway) turnBulbsOn() {
	var timerActive bool = g.bulbTimer.reset()
	if !timerActive {
		g.changeBulbStates(On)
	}
}

func (g *Gateway) turnBulbsOff() {
	g.changeBulbStates(Off)
}

func (g *Gateway) changeBulbStates(s State) {
	var bulbIdRegParams map[int]*RegisterParams = *g.senAndDev.getRegParams(g.bulbDev.getInts())
	var empty struct{}
	for bulbId, regParams := range bulbIdRegParams {
		var client *rpc.Client
		var err error
		client, err = rpc.Dial("tcp", regParams.Address+":"+regParams.Port)
		if err != nil {
			log.Printf("dialing error: %v", err)
		}
		client.Go("SmartBulb.ChangeState", ChangeStateParams{bulbId, s}, empty, nil)
	}
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
