package main

import (
	"sync"
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

type syncMapIntRegParam struct {
	sync.RWMutex
	m map[int]*RegisterParams
	i int
}

func (s *syncMapIntRegParam) addRegParam(regParam *RegisterParams) int {
	var i int
	s.Lock()
	s.m[s.i] = regParam
	i = s.i
	s.i++
	s.Unlock()
	return i
}

func (s *syncMapIntRegParam) getRegParams(is *map[int]bool) *map[*RegisterParams]bool {
	var newM map[*RegisterParams]bool = make(map[*RegisterParams]bool)
	s.RLock()
	for i, _ := range *is {
		r, ok := s.m[i]
		if ok {
			newM[r] = false
		}
	}
	s.RUnlock()
	return &newM
}

type syncMode struct {
	sync.RWMutex
	m Mode
}

func (s *syncMode) getMode() Mode {
	s.RLock()
	var mode Mode = s.m
	s.RUnlock()
	return mode
}

func (s *syncMode) setMode(mode Mode) {
	s.Lock()
	s.m = mode
	s.Unlock()
}
