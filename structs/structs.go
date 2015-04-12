// This file declares all the structs and interfaces needed by the gateway

package structs

import (
	"fmt"
	"github.com/nu7hatch/gouuid"
	"github.com/oleiade/lane"
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"log"
	"math"
	"os"
	"sort"
	"sync"
	"time"
)

type SyncMapIntBool struct {
	sync.RWMutex
	m map[int]bool
}

//create a new map
func NewSyncMapIntBool() *SyncMapIntBool {
	return &SyncMapIntBool{
		m: make(map[int]bool),
	}
}

//add to the map
func (s *SyncMapIntBool) AddInt(i int) {
	s.Lock()
	s.m[i] = false
	s.Unlock()
}

//get values
func (s *SyncMapIntBool) GetInts() *map[int]bool {
	var newM map[int]bool = make(map[int]bool)
	s.RLock()
	for i, _ := range s.m {
		newM[i] = false
	}
	s.RUnlock()
	return &newM
}

//check if present
func (s *SyncMapIntBool) Exists(i int) bool {
	s.RLock()
	_, ok := s.m[i]
	s.RUnlock()
	return ok
}

// A map for the Sensors and Device registered in the system.
//The int i keeps a track of last deviceid assigned to the most recent registered device in the system
type SyncMapIntRegParam struct {
	sync.RWMutex
	m map[int]*api.RegisterParams
	i int
}

//create a new SyncMapIntRegParam
func NewSyncMapIntRegParam() *SyncMapIntRegParam {
	return &SyncMapIntRegParam{
		m: make(map[int]*api.RegisterParams),
		i: 1,
	}
}

//Add a new device/sensor
func (s *SyncMapIntRegParam) AddRegParam(regParam *api.RegisterParams) int {
	var i int
	s.Lock()
	s.m[s.i] = regParam
	i = s.i
	s.i++
	s.Unlock()
	return i
}

//Fetch the values from the map
func (s *SyncMapIntRegParam) GetRegParams(is *map[int]bool) *map[int]*api.RegisterParams {
	var newM map[int]*api.RegisterParams = make(map[int]*api.RegisterParams)
	s.RLock()
	for i, _ := range *is {
		r, ok := s.m[i]
		if ok {
			newM[i] = r
		}
	}
	s.RUnlock()
	return &newM
}

// Defines the Mode of the system : Home or Away
type SyncMode struct {
	sync.RWMutex
	m api.Mode
}

//create new SyncMode
func NewSyncMode(mode api.Mode) *SyncMode {
	return &SyncMode{
		m: mode,
	}
}

//fetch the current mode value
func (s *SyncMode) GetMode() api.Mode {
	s.RLock()
	var mode api.Mode = s.m
	s.RUnlock()
	return mode
}

//assign a value to the mode
func (s *SyncMode) SetMode(mode api.Mode) {
	s.Lock()
	s.m = mode
	s.Unlock()
}

type SyncState struct {
	sync.RWMutex
	s api.State
}

func NewSyncState(state api.State) *SyncState {
	return &SyncState{
		s: state,
	}
}

func (s *SyncState) GetState() api.State {
	s.RLock()
	var state api.State = s.s
	s.RUnlock()
	return state
}

func (s *SyncState) SetState(state api.State) {
	s.Lock()
	s.s = state
	s.Unlock()
}

type SyncTimer struct {
	d time.Duration
	f func()
	sync.Mutex
	t *time.Timer
}

func (s *SyncTimer) Reset() bool {
	s.Lock()
	var active bool = s.t.Stop()
	s.t = time.AfterFunc(s.d, s.f)
	s.Unlock()
	return active
}

func (s *SyncTimer) Stop() bool {
	s.Lock()
	var active bool = s.t.Stop()
	s.Unlock()
	return active
}

func NewSyncTimer(d time.Duration, f func()) *SyncTimer {
	var s *SyncTimer = &SyncTimer{
		d: d,
		f: f,
		t: time.NewTimer(d),
	}
	s.t.Stop()
	return s
}

type SyncRegGatewayUserParam struct {
	sync.RWMutex
	u *api.RegisterGatewayUserParams
}

func NewSyncRegGatewayUserParam() *SyncRegGatewayUserParam {
	return &SyncRegGatewayUserParam{
		u: nil,
	}
}

func (s *SyncRegGatewayUserParam) Get() api.RegisterGatewayUserParams {
	s.RLock()
	var r api.RegisterGatewayUserParams = *s.u
	s.RUnlock()
	return r
}

func (s *SyncRegGatewayUserParam) Set(r api.RegisterGatewayUserParams) {
	s.Lock()
	s.u = &api.RegisterGatewayUserParams{
		Address: r.Address,
		Port:    r.Port,
	}
	s.Unlock()
}

func (s *SyncRegGatewayUserParam) Exists() bool {
	s.RLock()
	var e bool = s.u != nil
	s.RUnlock()
	return e
}

type SyncMapIntSyncFile struct {
	sync.RWMutex
	m map[int]*SyncFile
}

func NewSyncMapIntSyncFile() *SyncMapIntSyncFile {
	return &SyncMapIntSyncFile{
		m: make(map[int]*SyncFile),
	}
}

func (s *SyncMapIntSyncFile) Get(i int) (*SyncFile, bool) {
	s.RLock()
	f, ok := s.m[i]
	s.RUnlock()
	return f, ok
}

func (s *SyncMapIntSyncFile) Set(i int, f *SyncFile) {
	s.Lock()
	s.m[i] = f
	s.Unlock()
}

type SyncFile struct {
	sync.Mutex
	f *os.File
}

func NewSyncFile(name string) (*SyncFile, error) {
	var f *os.File
	var err error
	f, err = os.Create(name)
	if err != nil {
		log.Printf("Error creating file: %+v", err)
		return nil, err
	}
	return &SyncFile{
		f: f,
	}, nil
}

func (s *SyncFile) WriteString(str string) (int, error) {
	var n int
	var err error
	s.Lock()
	n, err = s.f.WriteString(str)
	s.Unlock()
	return n, err
}

type SyncMapIntStateInfo struct {
	sync.RWMutex
	m map[int]*api.StateInfo
}

func NewSyncMapIntStateInfo() *SyncMapIntStateInfo {
	return &SyncMapIntStateInfo{
		m: make(map[int]*api.StateInfo),
	}
}

func (s *SyncMapIntStateInfo) Get(i int) (*api.StateInfo, bool) {
	s.RLock()
	state, ok := s.m[i]
	s.RUnlock()
	return state, ok
}

func (s *SyncMapIntStateInfo) Set(i int, state *api.StateInfo) {
	s.Lock()
	s.m[i] = state
	s.Unlock()
}

type SyncMapIntOrderingNode struct {
	sync.RWMutex
	m map[int]api.OrderingNode
}

func NewSyncMapIntOrderingNode() *SyncMapIntOrderingNode {
	return &SyncMapIntOrderingNode{
		m: make(map[int]api.OrderingNode),
	}
}

func (s *SyncMapIntOrderingNode) Get(i int) (api.OrderingNode, bool) {
	s.RLock()
	node, ok := s.m[i]
	s.RUnlock()
	return node, ok
}

func (s *SyncMapIntOrderingNode) Set(i int, node api.OrderingNode) {
	s.Lock()
	s.m[i] = node
	s.Unlock()
}

func (s *SyncMapIntOrderingNode) GetMap() map[int]api.OrderingNode {
	s.RLock()
	//shallow copy
	n := s.m
	s.RUnlock()
	return n
}

func (s *SyncMapIntOrderingNode) GetKeys() []int {
	var idx int = 0
	s.RLock()
	var keys []int = make([]int, len(s.m))
	for key := range s.m {
		keys[idx] = key
		idx++
	}
	s.RUnlock()
	return keys
}

type SyncMapNameReportState struct {
	sync.RWMutex
	m map[api.Name]*api.ReportState
}

func NewSyncMapNameReportState() *SyncMapNameReportState {
	return &SyncMapNameReportState{
		m: make(map[api.Name]*api.ReportState),
	}
}

func (s *SyncMapNameReportState) Get(n api.Name) (*api.ReportState, bool) {
	s.RLock()
	rs, ok := s.m[n]
	s.RUnlock()
	return rs, ok
}

func (s *SyncMapNameReportState) Set(n api.Name, rs *api.ReportState) {
	s.Lock()
	s.m[n] = rs
	s.Unlock()
}

//PeerTable struct keeps a track of all peers(deviceID and address:port) in the system.
type PeerTable struct {
	p api.PMAP // peers map[DeviceId] address:port
	sync.RWMutex
}

//NewPeerTable() is called whenever a new gateway is created
func NewPeerTable() *PeerTable {
	return &PeerTable{
		p: make(map[int]string),
	}
}

//AddPeer(): When a new Peer registers with the gateway and obtains a new Device ID.
//The DeviceID and Address:Port are added to the PeerTable
func (s *PeerTable) AddPeer(i int, address string) {
	s.Lock()
	s.p[i] = address
	s.Unlock()
}

// ShowPeer() is mainly used for testing if the peertable is updated correctly
func (s *PeerTable) ShowPeer() {
	s.RLock()
	for key, value := range s.p {
		fmt.Println(s.p[key], key, value)
	}
	s.RUnlock()
}

//FindPeer() returns the value of the map
func (s *PeerTable) FindPeerAddress(i int) string {
	s.RLock()
	address := s.p[i]
	s.RUnlock()
	return address
}

//Returns the length of peertable
func (s *PeerTable) PeerTableLength() int {
	s.RLock()
	length := len(s.p)
	s.RUnlock()
	return length
}

//Delete a peer from the peertable
func (s *PeerTable) DeletePeer(i int) {
	s.Lock()
	delete(s.p, i)
	s.Unlock()
}

//Synchronizes concurrent access to an int.
type SyncInt struct {
	i int
	sync.RWMutex
}

//Creates an new instance of the struct.
func NewSyncInt(i int) *SyncInt {
	return &SyncInt{i: i}
}

//Returns the int.
func (s *SyncInt) Get() int {
	var r int
	s.RLock()
	r = s.i
	s.RUnlock()
	return r
}

//Increments the int by 1 then returns it.
func (s *SyncInt) IncThenGet() int {
	var r int
	s.RLock()
	s.i++
	r = s.i
	s.RUnlock()
	return r
}

//Set the value of the int.
func (s *SyncInt) Set(n int) {
	s.Lock()
	s.i = n
	s.Unlock()
}

type SyncLogicalEventContainer struct {
	//maps event ID to a map of device id to booleans
	//true indicates the device has acknowledge the event
	mapEventToAcks map[uuid.UUID]*map[int]bool
	//data structure is already synchornized
	eventQ *lane.PQueue
	sync.Mutex
}

func NewSyncLogicalEventContainer() *SyncLogicalEventContainer {
	return &SyncLogicalEventContainer{
		mapEventToAcks: make(map[uuid.UUID]*map[int]bool),
		eventQ:         lane.NewPQueue(lane.MINPQ),
	}
}

//Add event to container
func (s *SyncLogicalEventContainer) AddEvent(event api.LogicalEvent) {
	//log.Printf("Adding event: %+v", event)
	s.eventQ.Push(event, event.StateInfo.Clock)
}

//Update container with acknowledgement
func (s *SyncLogicalEventContainer) AddAck(event api.LogicalEvent) {
	//log.Printf("Adding ack: %+v", event)
	var acksPtr *map[int]bool
	var ok bool
	s.Lock()
	acksPtr, ok = s.mapEventToAcks[event.EventID]
	if ok {
		//if ack map already exists, add ack
		(*acksPtr)[event.SrcId] = true
	} else {
		//else make ack map, set all acks to false, then add ack
		var acks map[int]bool
		acks = make(map[int]bool)
		s.mapEventToAcks[event.EventID] = &acks
		for idx := range event.DestIDs {
			var id int = event.DestIDs[idx]
			acks[id] = false
		}
		acks[event.SrcId] = true
	}
	s.Unlock()
}

//Return the event with the lowest clock value if it has been
//acknowledged by all processes.
func (s *SyncLogicalEventContainer) GetHeadIfAcked() (*api.LogicalEvent, bool) {
	s.Lock()
	//log.Printf("eventQ = %+v\n", s.eventQ)
	//log.Printf("mapEventToAcks = %+v\n", s.mapEventToAcks)
	//if queue empty return
	if s.eventQ.Size() == 0 {
		//log.Printf("eventQ.Size == 0\n")
		s.Unlock()
		return nil, false
	}
	//check that the earliest event has been acked by all processes
	head, _ := s.eventQ.Head()
	var event api.LogicalEvent = head.(api.LogicalEvent)
	var acksPtr *map[int]bool
	var ok bool
	acksPtr, ok = s.mapEventToAcks[event.EventID]
	//if no acks exist return
	if !ok {
		//log.Printf("No acks exists\n")
		s.Unlock()
		return nil, false
	}
	//for id, hasAcked := range *acksPtr {
	for _, hasAcked := range *acksPtr {
		if !hasAcked {
			//log.Printf("Node has not acked: %d\n", id)
			s.Unlock()
			return nil, false
		}
	}
	//now we know all processes have acked the event
	//remove the earliest event from the queue
	head, _ = s.eventQ.Pop()
	event = head.(api.LogicalEvent)
	//delete the ack map
	delete(s.mapEventToAcks, event.EventID)
	s.Unlock()
	//return the fully acked event
	//log.Printf("returning event: %+v\n", event)
	return &event, true
}

//Used to cache the latest states in the database.
type SyncLatestStateInfos struct {
	size       int
	stateInfos map[int]api.StateInfo
	sync.RWMutex
}

func NewSyncLatestStateInfos(size int) *SyncLatestStateInfos {
	return &SyncLatestStateInfos{
		size:       size,
		stateInfos: make(map[int]api.StateInfo),
	}
}

func (this *SyncLatestStateInfos) AddStateInfo(s api.StateInfo) {
	this.Lock()
	//if size of map less than size just add new state info
	if len(this.stateInfos) < this.size {
		this.stateInfos[s.Clock] = s
		this.Unlock()
		return
	}
	//find the earliest state info in the map
	var earliest int = math.MaxInt32
	for clock, _ := range this.stateInfos {
		if earliest < clock {
			earliest = clock
		}
	}
	//if it is earlier than the one to be added
	//replace the earliest with the one to be added
	if earliest < s.Clock {
		delete(this.stateInfos, earliest)
		this.stateInfos[s.Clock] = s
	}
	this.Unlock()
}

func (this *SyncLatestStateInfos) GetBeforeAndAfter(clock int) (*api.StateInfo, *api.StateInfo) {
	this.RLock()
	//get the clock values
	var idx int = 0
	var clocks []int = make([]int, len(this.stateInfos))
	for key := range this.stateInfos {
		clocks[idx] = key
		idx++
	}
	//sort the clock values in increasing order
	sort.Ints(clocks)
	//find before and after
	var beforePtr *api.StateInfo = nil
	var afterPtr *api.StateInfo = nil
	for _, c := range clocks {
		if c < clock {
			before := this.stateInfos[c]
			beforePtr = &before
		}
		if clock < c {
			after := this.stateInfos[c]
			afterPtr = &after
			break
		}
	}
	this.RUnlock()
	return beforePtr, afterPtr
}

type SyncMapIntSyncLatestStateInfos struct {
	latestSizes int
	m           map[int]*SyncLatestStateInfos
	sync.Mutex
}

func NewSyncMapIntSyncLatestStateInfos(latestSizes int) *SyncMapIntSyncLatestStateInfos {
	return &SyncMapIntSyncLatestStateInfos{
		latestSizes: latestSizes,
		m:           make(map[int]*SyncLatestStateInfos),
	}
}

func (this *SyncMapIntSyncLatestStateInfos) Get(i int) *SyncLatestStateInfos {
	this.Lock()
	latest, ok := this.m[i]
	if !ok {
		latest = NewSyncLatestStateInfos(this.latestSizes)
		this.m[i] = latest
	}
	this.Unlock()
	return latest
}
