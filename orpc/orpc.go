package rpc

import ()

func Dial(*type string, *ip string) (int, int){
	client, err = rpc.Dial(type,ip)
	return client, err
}

func 