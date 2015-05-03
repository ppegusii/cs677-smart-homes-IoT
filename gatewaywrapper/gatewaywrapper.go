// This file is wraps the gateway, but provides no new functionality.
// It demonstrates the generic method that can be used to extend
// the gateway.

package gatewaywrapper

import (
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"github.com/ppegusii/cs677-smart-homes-IoT/util"
	"log"
)

type GatewayWrapper struct {
	api.GatewayInterface
}

func NewGatewayWrapper() api.GatewayWrapperInterface {
	return &GatewayWrapper{}
}

func (this *GatewayWrapper) RpcSync(ip, port, rpcName string, args interface{}, reply interface{}, isErrFatal bool) error {
	log.Printf("Before RPC reply: %+v\n", reply)
	var err error = util.RpcSync(ip, port, rpcName, args, reply, isErrFatal)
	log.Printf("After RPC reply: %+v\n", reply)
	return err
}

func (this *GatewayWrapper) SetGateway(g api.GatewayInterface) {
	this.GatewayInterface = g
}

func (g *GatewayWrapper) Register(params *api.RegisterParams, reply *int) error {
	log.Printf("Before Register id: %d\n", *reply)
	var err error = g.GatewayInterface.Register(params, reply)
	log.Printf("After Register id: %d\n", *reply)
	return err
}
