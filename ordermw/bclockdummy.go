package ordermw

import (
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"github.com/ppegusii/cs677-smart-homes-IoT/structs"
	"log"
	//"net"
	"fmt"
	"net/rpc"
	"time"
)

type BClockDummy struct {
	id             int
	ip             string
	port           string
	reportStates   *structs.SyncMapNameReportState
	peers          map[int]string // To keep a track of all peers
	pInterval      int            // Define the polling interval to check the Aliveness of current leader
	currentLeader  int            //Who is the current leader
	leaderElection bool           //Did I WIN the election
	offset         int32
}

func BClockNewDummy(id int, ip string, port string) *BClockDummy {
	var d *BClockDummy = &BClockDummy{
		id:             id,
		ip:             ip,
		port:           port,
		reportStates:   structs.NewSyncMapNameReportState(),
		peers:          make(map[int]string),
		pInterval:      2,
		leaderElection: true,
		offset:         0,
	}
	d.peers[id] = ip + ":" + port
	d.pInterval = 5
	if id == 0 {
		d.currentLeader = id
	} else {
		d.currentLeader = -1
	}
	fmt.Println("currentLeader is", d.currentLeader)
	d.start()
	return d
}

func (this *BClockDummy) start() {
	//register RPC server
	var err error = rpc.Register(api.OrderingMiddlewareRPCInterface(this))
	if err != nil {
		log.Fatal("rpc.Register error: %s\n", err)
	}
	/*
		var listener net.Listener
		listener, err = net.Listen("tcp", this.ip+":"+this.port)
		if err != nil {
			log.Fatal("net.Listen error: %s\n", err)
		}
		rpc.Accept(listener)
	*/
}

// ShowPeer() is mainly used for testing if the peertable is updated correctly
func (this *BClockDummy) ShowPeer() {
	for key, value := range this.peers {
		fmt.Println(key, value)
	}
}

//Multicasts new node notification to all other nodes.
//Called only by the gateway front-end application.
//From Application -> to the middleware
func (this *BClockDummy) SendNewNodeNotify(o api.OrderingNode) error {
	//Add the entry in the peer table
	this.peers[o.ID] = o.Address + ":" + o.Port
	// Testing code
	fmt.Println("Peer table in the middleware looks as below")
	this.ShowPeer()
	// Testing code ends
	return nil
}

//Accepts new node notifications
//Called only by other ordering implementations.
func (this *BClockDummy) ReceiveNewNodesNotify(params map[int]api.OrderingNode, _ *struct{}) error {
	//Check is the peer already exists in the peertable
	/*
		for key, value := range this.peers {
			if this.Exists(params.ID) {
				//Do nothing; peer exists
			} else {
				this.peers[params.ID] = params.Address + ":" + params.Port
			}
		}
		// Testing code
		for key, value := range this.peers {
			fmt.Println(this.peers[key], key, value)
		}
		// Testing code ends
	*/
	return nil
}

//**Ordinary unicast for clock sync.
//Logical clocks:
//Multicasts event notification to all other nodes.
//Called by applications instead of reporting state directly to another process.
func (this *BClockDummy) SendState(s api.StateInfo, destAddr string, destPort string) error {
	var event api.Event = api.Event{
		IsAck:      false,
		SrcAddress: this.ip,
		SrcId:      s.DeviceId,
		SrcPort:    this.port,
		StateInfo:  s,
	}
	var client *rpc.Client
	var err error
	client, err = rpc.Dial("tcp", destAddr+":"+destPort)
	if err != nil {
		log.Fatal("dialing error: %+v", err)
	}
	var empty struct{}
	err = client.Call("BClockDummy.ReceiveEvent", event, &empty)
	if err != nil {
		log.Fatal("calling error: %+v", err)
		return err
	}
	return nil
}

//**Simple delivery of state info to registered report state functions for clock sync.
//Logical clocks:
//Multicasts acknowledgement of event to all other nodes.
//Maintains a queue of messages delivering the one with the least clock value once
//all acknowledgments have been received. Therefore, there is a total ordering
//on messages delivered to the application. Those messages are delivered to
//registered report state functions.
//Called only by other ordering implementations.
func (this *BClockDummy) ReceiveEvent(params *api.Event, _ *struct{}) error {
	var rsPtr *api.ReportState
	var ok bool
	rsPtr, ok = this.reportStates.Get(params.StateInfo.DeviceName)
	if !ok {
		log.Printf("No registered func to handle device name: %d", params.StateInfo.DeviceName)
		return nil
	}
	var empty struct{}
	var rs api.ReportState = *rsPtr
	return rs(&(params.StateInfo), &empty)
}

//Register functions that handle the states received inside events.
func (this *BClockDummy) RegisterReportState(name api.Name, reportState api.ReportState) {
	this.reportStates.Set(name, &reportState)
}

//Receive PeerTable from other middleware and update your own peertable
func (this *BClockDummy) ReceivePeertableNotification(params *api.PMAP, _ *struct{}) error {
	//Check is the peer already exists in the peertable
	for key, _ := range *params {
		if this.peers[key] == (*params)[key] {
			//Do nothing; peer exists
		} else {
			this.peers[key] = (*params)[key]
		}
	}
	// Testing code
	fmt.Println("The middleware of gateway has send the following peertable to middleware of application")
	for key, value := range this.peers {
		fmt.Println(this.peers[key], key, value)
	}
	// Testing code ends
	go this.Bully()
	if this.id == this.currentLeader {
		go this.GetTime()
	}
	return nil
}

//Send PeerTable to other middlewares
func (this *BClockDummy) SendPeertableNotification(i int) {
	var params api.PMAP
	params = this.peers //peer table

	var empty struct{}
	var client *rpc.Client
	var err error

	for key, value := range this.peers {
		client, err = rpc.Dial("tcp", value)
		if err != nil {
			log.Println("error dialing from SendPeertableNotification : %+v", err)
			delete(this.peers, key)
		}
		fmt.Println("Sending the peertable to the middleware of device id ", key, this.peers[i])
		err = client.Call("BClockDummy.ReceivePeertableNotification", params, &empty)
		if err != nil {
			log.Fatal("calling error: %+v", err)
		}
	}
}

//Leader Election Algorithm : Algorithm implemented is Bully Algorithm
func (this *BClockDummy) Bully() {
	var i int
	var empty struct{}
	var client *rpc.Client
	var err error = nil
	// put a ticker for every 5 seconds...
	var ticker *time.Ticker = time.NewTicker(time.Duration(this.pInterval) * time.Second)
	for range ticker.C {
		if this.currentLeader > -1 {
			client, err = rpc.Dial("tcp", this.peers[this.currentLeader])
		}
		if (this.currentLeader == -1) || (err != nil) {
			//Send an election message to all higher deviceid's
			for key, value := range this.peers {
				if key > this.id {
					client, err = rpc.Dial("tcp", value)
					if err != nil {
						this.ShowPeer()
						log.Println("error dialing from Bully : %+v", key, err)
						delete(this.peers, key)
						for key, value := range this.peers {
							fmt.Println(key, value)
						}
					} else {
						this.leaderElection = true
						fmt.Println("Sending an Election Message from Device ID to device ID", this.id, key)
						err = client.Call("BClockDummy.Election", this.id, &empty)
						if err != nil {
							log.Println("calling error: %+v", err)
							delete(this.peers, key)
						}
					}
				}
			}
			time.Sleep(time.Second * 3)
			//Check if no OK was send
			if this.leaderElection == true {
				//Send IWIN notifications to everyone
				for key, value := range this.peers {
					client, err = rpc.Dial("tcp", value)
					if err != nil {
						log.Println("error dialing from Bully IWIN part: %+v", err)
						delete(this.peers, i)
					}
					fmt.Println("Sending an IWIN Message from Device ID to deviceID", this.id, key)
					this.currentLeader = this.id
					err = client.Call("BClockDummy.IWIN", this.id, &empty)
					if err != nil {
						log.Println("calling error: %+v", err)
						delete(this.peers, key)
					}
				}
			}
		}
		/*		if(this.id == this.currentLeader){
					go this.GetTime()
				}
		*/
	} //end of ticker code
}

//Receive Election message from other middlewares.
func (this *BClockDummy) Election(id int, _ *struct{}) error {
	//If the device id of the sender is lower than the receivers deviceid, Send OK message
	if this.id > id {
		var empty struct{}
		var client *rpc.Client
		var err error
		//Send an OK message back to the device
		client, err = rpc.Dial("tcp", this.peers[id])
		if err != nil {
			log.Println("error dialing from Election : %+v", this.peers[id], err)
			delete(this.peers, id)
			for key, value := range this.peers {
				fmt.Println(key, value)
			}
		} else {
			fmt.Println("Sending an OK Message to Device ID ", id)
			err = client.Call("BClockDummy.OKAY", this.id, &empty)
			if err != nil {
				log.Println("error Calling RPC Okay from Election() : %+v", id, err)
				delete(this.peers, id)
				for key, value := range this.peers {
					fmt.Println(key, value)
				}
			}
		}
	}
	return nil
}

//Receive OKAY message from higher device id middlewares.
func (this *BClockDummy) OKAY(id int, _ *struct{}) error {
	this.leaderElection = false
	return nil
}

//Send IWIN message to peers
func (this *BClockDummy) IWIN(id int, _ *struct{}) error {
	this.currentLeader = id
	fmt.Println("New Elected Leader is", this.currentLeader)
	return nil
}

//Send messages to each peer middleware asking for the timestamp
//Compute average time and then send the offset value back to the peers
func (this *BClockDummy) GetTime() {
	if this.id == this.currentLeader {
		var PeerTimestamps = make(map[int]int32)
		var offsetsum, average, count int32 = 0, 0, 0
		var empty struct{}
		var client *rpc.Client
		var err error
		var timestamp *api.BTimeStamp
		leadertime := int32(time.Now().Unix())
		fmt.Println("LeaderTime", leadertime)
		for key, value := range this.peers {
			fmt.Println(key, value)
		}
		for key, value := range this.peers {
			fmt.Println("Values of key and value are", key, value)
			client, err = rpc.Dial("tcp", value)
			if err != nil {
				log.Println("error dialing from GetTime: %+v", err)
				delete(this.peers, key)
			} else {
				fmt.Println("Sending a GetTimestamp request to Device ID ", key)
				err = client.Call("BClockDummy.SendTime", this.id, &timestamp)
				if err != nil {
					log.Println("calling error: %+v", err)
					delete(this.peers, key)
				} else {
					//Enter the timestamp in the map
					fmt.Println(timestamp)
					PeerTimestamps[timestamp.DeviceId] = timestamp.Timestamp
					fmt.Println("The PeerTimestamps map looks as below after entering the timestamp:")
					for key, value := range PeerTimestamps {
						fmt.Println(key, value)
					}
				}
			}
		}
		fmt.Println("The PeerTimestamps map looks as below:")
		for key, value := range PeerTimestamps {
			fmt.Println(key, value)
		}
		//Now, that we have all timestamps take average of all the timestamps
		for _, value := range PeerTimestamps {
			count++
			offsetsum = offsetsum + value - leadertime
		}
		average = offsetsum / count
		fmt.Println("Average offset and count values are", average, count)
		//Send the offsets back to the devices
		for key, value := range this.peers {
			client, err = rpc.Dial("tcp", value)
			if err != nil {
				log.Println("error dialing from GetTime to return offsets: %+v", err)
				delete(this.peers, key)
			} else {
				fmt.Println("Sending the Timestamp offsets from Device ID ", this.id)
				offset := average - ((offsetsum + PeerTimestamps[key] - leadertime) / count)
				//			offset := average + leadertime - PeerTimestamps[key]
				err = client.Call("BClockDummy.ReceiveOffset", offset, &empty)
				if err != nil {
					log.Println("calling error: %+v", err)
					delete(this.peers, key)
				} else {
					//Enter the timestamp in the map
					PeerTimestamps[timestamp.DeviceId] = timestamp.Timestamp
				}
			}
		}
	}
}

//Send Offsets to peers
func (this *BClockDummy) ReceiveOffset(offset int32, _ *struct{}) error {
	fmt.Println("Offset Send by the Leader is:", offset)
	this.offset = offset
	fmt.Println("Offset Timestamp is", this.offset)
	return nil
}

// This is an RPC call that returns the current Unix timestamp to the leader middleware
func (this *BClockDummy) SendTime(id int, timestamp *api.BTimeStamp) error {
	timestamp.DeviceId = this.id
	timestamp.Timestamp = int32(time.Now().Unix()) + this.offset
	fmt.Println("Unix timestamp is", timestamp.DeviceId, timestamp.Timestamp)
	return nil
}
