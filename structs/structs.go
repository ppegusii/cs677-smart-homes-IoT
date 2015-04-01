package structs

import (
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"os"
	"sync"
	"time"
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
