package util

import (
	"errors"
	"fmt"
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"log"
	"net"
	"os"
)

func LogCurrentState(s api.State) {
	var text string
	switch s {
	case api.On:
		text = "On"
		break
	case api.Off:
		text = "Off"
		break
	case api.MotionStart:
		text = "Motion"
		break
	case api.MotionStop:
		text = "No Motion"
		break
	case api.Open:
		text = "Open"
		break
	case api.Closed:
		text = "Closed"
		break
	default:
		text = "Invalid state"
		break
	}
	log.Printf("Current state: %s", text)
}

func StringToOrdering(s string) (api.Ordering, error) {
	switch s {
	case "n":
		return api.NoOrder, nil
	case "l":
		return api.LogicalClock, nil
	case "c":
		return api.ClockSync, nil
	default:
		return api.NoOrder, errors.New(fmt.Sprintf("Invalid ordering switch: %s", s))
	}
}

func GetOwnIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return (ipnet.IP.String())
			}
		}
	}
	log.Fatal("Error")
	return ""
}
