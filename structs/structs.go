// This file declares all the structs and interfaces needed by the gateway

package structs

import (
	"bufio"
	"encoding/json"
	"fmt"
	//"github.com/nu7hatch/gouuid"
	//"github.com/oleiade/lane"
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"log"
	"math"
	"os"
	//"sort"
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
		i: 0, //Returning back to zero. Gateways and databases no longer have IDs.
	}
}

//Add a new device/sensor and create an new ID
func (s *SyncMapIntRegParam) AddNewRegParam(regParam *api.RegisterParams) int {
	var i int
	s.Lock()
	s.m[s.i] = regParam
	i = s.i
	s.i++
	s.Unlock()
	return i
}

//Add a new device/sensor that already has an ID
func (s *SyncMapIntRegParam) AddExistingRegParam(regParam *api.RegisterParams, id int) {
	s.Lock()
	s.m[id] = regParam
	// increment the internal ID counter if a higher one is given.
	if s.i <= id {
		s.i++
	}
	s.Unlock()
}

// Get device/sensor info
func (s *SyncMapIntRegParam) GetRegParam(id int) (*api.RegisterParams, bool) {
	s.RLock()
	defer s.RUnlock()
	var r *api.RegisterParams
	var ok bool
	r, ok = s.m[id]
	if !ok {
		return nil, false
	}
	var ret api.RegisterParams = *r
	return &ret, true
}

//Remove a device/sensor
func (s *SyncMapIntRegParam) RemoveRegParam(id int) {
	s.Lock()
	delete(s.m, id)
	s.Unlock()
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

// Get all values in the map
func (s *SyncMapIntRegParam) GetAllRegParams() *[]api.RegisterParams {
	s.RLock()
	var all []api.RegisterParams = make([]api.RegisterParams, len(s.m))
	var idx int = 0
	for _, r := range s.m {
		all[idx] = *r
		idx++
	}
	s.RUnlock()
	return &all
}

// Get the size.
func (s *SyncMapIntRegParam) Size() int {
	s.RLock()
	var z int = len(s.m)
	s.RUnlock()
	return z
}

// Defines the Mode of the system : Home or Away
type SyncModeClock struct {
	sync.RWMutex
	mc api.ModeAndClock
}

//create new SyncMode
func NewSyncModeClock(modeClock api.ModeAndClock) *SyncModeClock {
	return &SyncModeClock{
		mc: modeClock,
	}
}

//fetch the current mode value
func (s *SyncModeClock) GetModeClock() api.ModeAndClock {
	s.RLock()
	var mc api.ModeAndClock = s.mc
	s.RUnlock()
	return mc
}

//assign a value to the mode
func (s *SyncModeClock) SetModeAndClock(mc api.ModeAndClock) {
	s.Lock()
	s.mc = mc
	s.Unlock()
}

//assign a value to the mode
func (s *SyncModeClock) SetModeClock(m api.Mode, c int64) {
	s.Lock()
	s.mc.Mode = m
	s.mc.Clock = c
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

func (s *SyncMapIntSyncFile) GetAll() map[int]*SyncFile {
	s.RLock()
	r := s.m
	s.RUnlock()
	return r
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

func (s *SyncFile) writeJson(o interface{}) error {
	var err error
	var b []byte
	var str string
	b, err = json.Marshal(o)
	if err != nil {
		return err
	}
	str = string(b)
	str += "\n"
	_, err = s.f.WriteString(str)
	return err
}

// This is bad.
// Assume clock values are unique
// Read all structs into a map of clock to struct
// Write structs back to file
func (s *SyncFile) WriteRegParam(things *[]api.RegisterParams) error {
	s.Lock()
	defer s.Unlock()
	var m map[int64]api.RegisterParams = make(map[int64]api.RegisterParams)
	old, _ := s.getRegParamsSince(-1)
	for _, thing := range append(*old, *things...) {
		m[thing.Clock] = thing
	}
	s.f.Seek(0, 0)
	var err error
	for _, thing := range m {
		err = s.writeJson(thing)
	}
	return err
}

// Locking wrapper to internal function
func (s *SyncFile) GetRegParamsSince(startTime int64) (*[]api.RegisterParams, error) {
	s.Lock()
	defer s.Unlock()
	return s.getRegParamsSince(startTime)
}

// Unmarshal all lines into structs
// Return the lines that have times >= the given clock
func (s *SyncFile) getRegParamsSince(startTime int64) (*[]api.RegisterParams, error) {
	s.f.Seek(0, 0)
	var scanner *bufio.Scanner = bufio.NewScanner(s.f)
	var err error
	var b []byte
	var things []api.RegisterParams = make([]api.RegisterParams, 0)
	// http://stackoverflow.com/questions/8757389/reading-file-line-by-line-in-go
	for scanner.Scan() {
		b = []byte(scanner.Text())
		var thing api.RegisterParams
		err = json.Unmarshal(b, &thing)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		if thing.Clock >= startTime {
			things = append(things, thing)
		}
	}
	return &things, nil
}

// This is bad.
// Assume clock values are unique
// Read all structs into a map of clock to struct
// Write structs back to file
func (s *SyncFile) WriteStateInfo(things *[]api.StateInfo) error {
	s.Lock()
	defer s.Unlock()
	var m map[int64]api.StateInfo = make(map[int64]api.StateInfo)
	old, _ := s.getStateInfoSince(-1)
	for _, thing := range append(*old, *things...) {
		m[thing.Clock] = thing
	}
	s.f.Seek(0, 0)
	var err error
	for _, thing := range m {
		err = s.writeJson(thing)
	}
	return err
}

// Locking wrapper to internal function
func (s *SyncFile) GetStateInfoSince(startTime int64) (*[]api.StateInfo, error) {
	s.Lock()
	defer s.Unlock()
	return s.getStateInfoSince(startTime)
}

// Unmarshal all lines into structs
// Return the lines that have times >= the given clock
func (s *SyncFile) getStateInfoSince(startTime int64) (*[]api.StateInfo, error) {
	s.f.Seek(0, 0)
	var scanner *bufio.Scanner = bufio.NewScanner(s.f)
	var err error
	var b []byte
	var things []api.StateInfo = make([]api.StateInfo, 0)
	// http://stackoverflow.com/questions/8757389/reading-file-line-by-line-in-go
	for scanner.Scan() {
		b = []byte(scanner.Text())
		var thing api.StateInfo
		err = json.Unmarshal(b, &thing)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		if thing.Clock >= startTime {
			things = append(things, thing)
		}
	}
	return &things, nil
}

//Get the last state info recorded just before "when".
func (s *SyncFile) GetStateInfoHappensBefore(when int64) *api.StateInfo {
	s.Lock()
	defer s.Unlock()
	var stateInfos *[]api.StateInfo
	var err error
	stateInfos, err = s.getStateInfoSince(-1)
	if err != nil {
		log.Printf("Error getting state infos: %+v\n", err)
		return nil
	}
	var before int64 = math.MinInt64
	var beforeSI *api.StateInfo = nil
	for _, si := range *stateInfos {
		if si.Clock > before && si.Clock < when {
			beforeSI = &si
			before = si.Clock
		}
	}
	return beforeSI
}

//Get the latest state info recorded.
func (s *SyncFile) GetLatestStateInfo() *api.StateInfo {
	s.Lock()
	defer s.Unlock()
	var stateInfos *[]api.StateInfo
	var err error
	stateInfos, err = s.getStateInfoSince(-1)
	if err != nil {
		log.Printf("Error getting state infos: %+v\n", err)
		return nil
	}
	var stateInfo *api.StateInfo
	var latest int64 = math.MinInt64
	for _, si := range *stateInfos {
		if si.Clock > latest {
			stateInfo = &si
			latest = si.Clock
		}
	}
	return stateInfo
}

/*
// Locking wrapper to internal function
func (s *SyncFile) GetThingSince(startTime int, t api.ClockInterface) (*[]api.ClockInterface, error) {
	s.Lock()
	defer s.Unlock()
	return s.getThingSince(startTime, t)
}

// Unmarshal all lines into structs
// Return the lines that have times greater than the given clock
func (s *SyncFile) getThingSince(startTime int, t api.ClockInterface) (*[]api.ClockInterface, error) {
	s.f.Seek(0, 0)
	var scanner *bufio.Scanner = bufio.NewScanner(s.f)
	var err error
	var b []byte
	var things []api.ClockInterface = make([]api.ClockInterface, 0)
	// http://stackoverflow.com/questions/8757389/reading-file-line-by-line-in-go
	for scanner.Scan() {
		b = []byte(scanner.Text())
		sw
		var thing api.ClockInterface
		err = json.Unmarshal(b, &thing)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		if thing.GetClock() > startTime {
			things = append(things, thing)
		}
	}
	return &things, nil
}
*/

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

//Synchronizes concurrent access to an int64.
type SyncInt64 struct {
	i int64
	sync.RWMutex
}

//Creates an new instance of the struct.
func NewSyncInt64(i int64) *SyncInt64 {
	return &SyncInt64{i: i}
}

//Returns the int.
func (s *SyncInt64) Get() int64 {
	var r int64
	s.RLock()
	r = s.i
	s.RUnlock()
	return r
}

//Set the value of the int.
func (s *SyncInt64) Set(n int64) {
	s.Lock()
	s.i = n
	s.Unlock()
}

/*
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
*/

//Synchronizes concurrent access to an bool.
type SyncBool struct {
	b bool
	sync.RWMutex
}

//Creates an new instance of the struct.
func NewSyncBool(b bool) *SyncBool {
	return &SyncBool{b: b}
}

//Returns the bool.
func (s *SyncBool) Get() bool {
	var r bool
	s.RLock()
	r = s.b
	s.RUnlock()
	return r
}

//Set the value of the bool.
func (s *SyncBool) Set(b bool) {
	s.Lock()
	s.b = b
	s.Unlock()
}

//Main Cache structure
type Cache struct {
	lock      sync.RWMutex
	used      int
	datamap   map[int]api.StateInfo
	size      int           //size of cache
	evictlist map[int]int64 // key is the index of record of datamap and value is the reference count
}

//Create new cache
func NewCache(maxEntries int) *Cache {
	return &Cache{
		used:      api.EMPTY,
		size:      maxEntries,
		datamap:   make(map[int]api.StateInfo),
		evictlist: make(map[int]int64),
	}
}

//Get a specific record from the cache
func (c *Cache) Get(key int) *api.StateInfo {
	c.lock.Lock()
	defer c.lock.Unlock()
	//Check that the cache is non-Empty
	if c.used == api.EMPTY {
		return nil
	} else {
		data, exists := c.datamap[key]
		if exists == false {
			return nil
		} else {
			c.evictlist[key] = int64(time.Now().Unix()) //set the reference time to current timestamp
			return &data
		}
	}
}

//Add a new record in the cache
func (c *Cache) Set(key int, d *api.StateInfo) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.datamap[key] = *d
	c.evictlist[key] = int64(time.Now().Unix())
	c.used++
}

// Delete a cache entry
func (c *Cache) Delete(key int) bool {
	var s bool
	c.lock.Lock()
	defer c.lock.Unlock()
	//Check if cache has non zero length
	if c.used > 0 {
		delete(c.datamap, key)
		c.evictlist[key] = 0
		c.used-- //Decrement the number of cache entries in the cache map
		s = true
	} else {
		s = false
	}
	return s
}

//Find the length of the cache; returns the number of entries in the cache
func (c *Cache) UsedCache() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.used
}

//Find the size of the cache; returns the max number of entries the cache can hold
func (c *Cache) LenCache() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.size
}

//Find the oldest entry in the Cache for eviction, returns the index of map to be evicted
func (c *Cache) OldCache() int {
	c.lock.RLock()
	defer c.lock.RUnlock()

	min := c.evictlist[0]
	minindex := 0
	for index, value := range c.evictlist {
		if value < min && value > 0 {
			min = value
			minindex = index
		}
	}
	fmt.Println("Evicting page with index number %d and timestamp is %d", minindex, min)
	return (minindex)
}

//search for index with timestamp as 0
func (c *Cache) Get0timestamp() int {
	var Zindex int = -1
	c.lock.RLock()
	defer c.lock.RUnlock()
	for index, value := range c.evictlist {
		if value == 0 {
			Zindex = index
			break
		}
	}
	if Zindex > -1 {
		fmt.Println("Hole found at index number ", Zindex)
	} else {
		fmt.Println("No hole found... go kill that stale page ... Evict it")
	}
	return Zindex
}

//this function is for testing the reference timestamp
//search for index with timestamp as 0
func (c *Cache) Gettimestamp(key int) int64 {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return (c.evictlist[key])
}

//Check if entry exists in the cachemap
func (c *Cache) Exists(key int) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()

	_, ok := c.datamap[key]
	if ok {
		return true
	} else {
		return false
	}
}

//Call AddEntry to add new stateinfo in cache...
// No need to check if the device record exists in cache... AddEntry will handle it all
func (cachemap *Cache) AddEntry(d *api.StateInfo) {
	var Zindex, evict int = -1, - 1
	//Check if the Cache already contains this device info
	Zindex = cachemap.LookupDeviceID(d.DeviceId)
	if Zindex > -1 {
		//Found an existing entry, so just update its cache and touch the ref timestamp
		cachemap.Set(Zindex, d)
		fmt.Println("Updated the cachemap value to the new stateInfo value and touched the ref timestamp")
	} else {
		//The entry does not exist add it in the cache ... But, where... Follow the code
		//Check if the Cache is full
		fmt.Println(cachemap.UsedCache(), cachemap.LenCache(), cachemap.Exists(cachemap.LenCache()-1))
		if cachemap.UsedCache() < cachemap.LenCache() && !cachemap.Exists(cachemap.LenCache()-1) {
			//Since there are indices not inserted into the cache map use
			// c.used value to find the index to add the value at
			fmt.Println("The cache is not full, so appending entry at the end of cachemap")
			cachemap.Set(cachemap.UsedCache(), d)
		} else {
			// The cache entries seem to be used but wait there might be holes inside,
			//So, let us use the timestamp field to find any 0's inside indicating the corresponding index is free
			// due to a prior delete request of a particular block
			Zindex = cachemap.Get0timestamp()
			if Zindex != -1 {
				//Yea, we found a hole in the cache map. Now, get that damn new entry at this spot.
				fmt.Println("The cache has a hole, so replacing hole with the new entry in cachemap at index ", Zindex)
				cachemap.Set(Zindex, d)
			} else {
				//Ok, so no hole found in the cachemap. I command you to evict an entry based on LRU
				fmt.Println("The cache has no holes, so evicting the LRU entry in cachemap")
				evict = cachemap.OldCache()
				cachemap.Set(evict, d)
			}
		}
	}
}

//Find the oldest entry in the Cache for eviction, returns the index of map to be evicted
func (c *Cache) LookupDeviceID(id int) int {
	var devindex int = -1
	c.lock.RLock()
	defer c.lock.RUnlock()

	for index, _ := range c.datamap {
		if c.datamap[index].DeviceId == id {
			devindex = index
			break
		}
	}
	if devindex == -1 {
		fmt.Println("No such device found")
	} else {
		fmt.Println("Found the device at index ", devindex)
	}
	return (devindex)
}

// To retrive the StateInfo of a device for cache
//If GetEntry returns a nil it means cache does not have that value, fetch it from the database
func (c *Cache) GetEntry(id int) *api.StateInfo {
	var Zindex int = -1
	//Check if the Cache already contains this device info
	Zindex = c.LookupDeviceID(id)
	if (Zindex > -1){
		fmt.Println("Device information exists")
		return c.Get(Zindex)
	} else {
		return nil
	}
}
