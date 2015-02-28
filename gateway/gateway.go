package gateway

import (
	"errors"
	"fmt"
	"sync"
)

type gateway struct {
	mode     Mode
	modeLock sync.Mutex
}

func NewGateway(mode Mode) Interface {
	return &gateway{
		mode: mode,
	}
}

func (g *gateway) Register(params *RegisterParams, reply *int) error {
	return nil
}

func (g *gateway) ReportState(params *ReportStateParams, _ *struct{}) error {
	return nil
}

func (g *gateway) ChangeMode(params *ChangeModeParams, _ *struct{}) error {
	var err error = nil
	switch params.Mode {
	case Home:
	case Away:
		g.modeLock.Lock()
		g.mode = params.Mode
		g.modeLock.Unlock()
		break
	default:
		err = errors.New(fmt.Sprintf("Invalid Mode: %v", params.Mode))
	}
	return err
}
