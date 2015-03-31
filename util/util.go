package util

import (
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"log"
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
	default:
		text = "Invalid state"
		break
	}
	log.Printf("Current state: %s", text)
}
