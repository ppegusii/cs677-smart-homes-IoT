package structs

import (
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"log"
	"os"
	"sync"
	"time"
	"fmt"
)

type SyncMapIntBool struct {
	sync.RWMutex
	m map[int]bool
}

func NewSyncMapIntBool() *SyncMapIntBool {
	return &SyncMapIntBool{
		m: make(map[int]bool),
	}
}

func (s *SyncMapIntBool) AddInt(i int) {
	s.Lock()
	s.m[i] = false
	s.Unlock()
}

func (s *SyncMapIntBool) GetInts() *map[int]bool {
	var newM map[int]bool = make(map[int]bool)
	s.RLock()
	for i, _ := range s.m {
		newM[i] = false
	}
	s.RUnlock()
	return &newM
}

func (s *SyncMapIntBool) Exists(i int) bool {
	s.RLock()
	_, ok := s.m[i]
	s.RUnlock()
	return ok
}

type SyncMapIntRegParam struct {
	sync.RWMutex
	m map[int]*api.RegisterParams
	i int
}

func NewSyncMapIntRegParam() *SyncMapIntRegParam {
	return &SyncMapIntRegParam{
		m: make(map[int]*api.RegisterParams),
	}
}

func (s *SyncMapIntRegParam) AddRegParam(regParam *api.RegisterParams) int {
	var i int
	s.Lock()
	s.m[s.i] = regParam
	i = s.i
	s.i++
	s.Unlock()
	return i
}

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

type SyncMode struct {
	sync.RWMutex
	m api.Mode
}

func NewSyncMode(mode api.Mode) *SyncMode {
	return &SyncMode{
		m: mode,
	}
}

func (s *SyncMode) GetMode() api.Mode {
	s.RLock()
	var mode api.Mode = s.m
	s.RUnlock()
	return mode
}

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
	m map[int]*api.OrderingNode
}

func NewSyncMapIntOrderingNode() *SyncMapIntOrderingNode {
	return &SyncMapIntOrderingNode{
		m: make(map[int]*api.OrderingNode),
	}
}

func (s *SyncMapIntOrderingNode) Get(i int) (*api.OrderingNode, bool) {
	s.RLock()
	node, ok := s.m[i]
	s.RUnlock()
	return node, ok
}

func (s *SyncMapIntOrderingNode) Set(i int, node *api.OrderingNode) {
	s.Lock()
	s.m[i] = node
	s.Unlock()
}

func (s *SyncMapIntOrderingNode) GetMap() map[int]*api.OrderingNode {
	s.RLock()
	//shallow copy
	n := s.m
	s.RUnlock()
	return n
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
	p api.PMAP  // peers map[DeviceId] address:port
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

func (s *PeerTable) PeerTableLength() int {
	s.RLock()
	length := len(s.p)
	s.RUnlock()
	return length
}

func (s *PeerTable) DeletePeer(i int) {
	s.Lock()
	delete(s.p,i)
	s.Unlock()
}
