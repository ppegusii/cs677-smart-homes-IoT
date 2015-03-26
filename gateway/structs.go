package main

import (
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"sync"
	"time"
)

type syncMapIntBool struct {
	sync.RWMutex
	m map[int]bool
}

func (s *syncMapIntBool) addInt(i int) {
	s.Lock()
	s.m[i] = false
	s.Unlock()
}

func (s *syncMapIntBool) getInts() *map[int]bool {
	var newM map[int]bool = make(map[int]bool)
	s.RLock()
	for i, _ := range s.m {
		newM[i] = false
	}
	s.RUnlock()
	return &newM
}

func (s *syncMapIntBool) exists(i int) bool {
	s.RLock()
	_, ok := s.m[i]
	s.RUnlock()
	return ok
}

type syncMapIntRegParam struct {
	sync.RWMutex
	m map[int]*api.RegisterParams
	i int
}

func (s *syncMapIntRegParam) addRegParam(regParam *api.RegisterParams) int {
	var i int
	s.Lock()
	s.m[s.i] = regParam
	i = s.i
	s.i++
	s.Unlock()
	return i
}

func (s *syncMapIntRegParam) getRegParams(is *map[int]bool) *map[int]*api.RegisterParams {
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

type syncMode struct {
	sync.RWMutex
	m api.Mode
}

func (s *syncMode) getMode() api.Mode {
	s.RLock()
	var mode api.Mode = s.m
	s.RUnlock()
	return mode
}

func (s *syncMode) setMode(mode api.Mode) {
	s.Lock()
	s.m = mode
	s.Unlock()
}

type syncTimer struct {
	d time.Duration
	f func()
	sync.Mutex
	t *time.Timer
}

func (s *syncTimer) reset() bool {
	s.Lock()
	var active bool = s.t.Stop()
	s.t = time.AfterFunc(s.d, s.f)
	s.Unlock()
	return active
}

func newSyncTimer(d time.Duration, f func()) *syncTimer {
	var s *syncTimer = &syncTimer{
		d: d,
		f: f,
		t: time.NewTimer(d),
	}
	s.t.Stop()
	return s
}
