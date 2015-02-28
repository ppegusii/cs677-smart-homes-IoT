package main

import (
	"fmt"
	"net/rpc"
)

func client() {
	c, err := rpc.Dial("tcp", "127.0.0.1:9999")
	if err != nil {
		fmt.Println(err)
		return
	}
	var result int64

	err = c.Call("Server.Register", int64(1), &result)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Yea... I established a connection")
		fmt.Println("Value of result returned is: ", result, &result)
	}
}

func main() {
	client()
}