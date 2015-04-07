package middleware

import (
	//	"errors"
	//	"fmt"
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	//	"github.com/ppegusii/cs677-smart-homes-IoT/structs"
	//	"github.com/ppegusii/cs677-smart-homes-IoT/gateway"
	"log"
	//	"net"
	"net/rpc"
	"time"
)

type Middleware struct {
	id              int
	ip              string
	port            string
	peers           api.PMAP // To keep a track of all peers
	pollingInterval int
	currentLeader   int
	leaderElection  bool
}

func NewMiddleware(id int, ip string, pollingInterval int, port string) *Middleware {
	var m *Middleware = &Middleware{
		id:              id,
		ip:              ip,
		pollingInterval: pollingInterval,
		port:            port,
		peers:           make(map[int]string),
		currentLeader:   -1, //Check this and modify
		leaderElection:  true,
	}

	// Populate the pertable of middleware
	// How to do this thinking
	return m
}

/*
func (m *Middleware) start() {
	var err error = rpc.Register(api.Middleware(m))
	if err != nil {
		log.Fatal("rpc.Register error: %s\n", err)
	}
	var listener net.Listener
	listener, err = net.Listen("tcp", m.ip+":"+m.port)
	if err != nil {
		log.Fatal("net.Listen error: %s\n", err)
	}
	go rpc.Accept(listener)
	g.pollLeader()
}
*/

func (m *Middleware) pollLeader() {
	//this function checks if the current leader is alive
	var ticker *time.Ticker = time.NewTicker(time.Duration(m.pollingInterval) * time.Second)
	for range ticker.C {
		var client *rpc.Client
		var err error
		client, err = rpc.Dial("tcp", m.peers[m.currentLeader])
		if err != nil {
			log.Printf("Leader is unreachable ... start elections")
			m.Bully()
			continue
		} else {
			log.Printf("Leader is alive, My Son! Hold on to fulfill your dream of becoming a leader")
			continue
		}

	}
}

func (m *Middleware) Bully() int {
	var i int
	//send an election message to all higher deviceid's
	for i = (m.id + 1); i < len-1; i++ {
		(*peers)[i] = m.peers.FindPeerAddress(i)
		//Middleware 1-to-1 call
		g.middleware.sendtopeer(i) //MIDDLEWARE FUNCTION CALL : This has to be a blocking time out call
		//now, wait for a reply within the next n secs
		//if OK message => then drop from election set the flag = 0
		if deviceid > m.id { //	if (reply = OK from middleware && deviceid > m.id){ // MIDDLEWARE FUNCTION CALL
			leaderElection = false
		} else {
			//else no OK receied => Broadcast I won message
			m.currentleader = m.id
			m.middleware.sendtoall("IWON") ///MIDDLEWARE FUNCTION CALL
		}
	}
}

//
func (g *Gateway) BerkeleyClock() {}
