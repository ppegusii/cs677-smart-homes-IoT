/* This file simulates the user inputs to toggle between HOME/AWAY.
The listening port for user is fixed to 6775 */

package main

import (
	"flag"
)

func main() {
	//parse args
	var gatewayIp *string = flag.String("I", "127.0.0.1", "gateway IP address")
	var gatewayPort *string = flag.String("P", "6770", "gateway TCP port")
	var selfIp *string = flag.String("i", "127.0.0.1", "IP address")
	var selfPort *string = flag.String("p", "6775", "TCP port")
	flag.Parse()

	//start user
	var u *User = newUser(*gatewayIp, *gatewayPort, *selfIp, *selfPort)
	u.start()
}
