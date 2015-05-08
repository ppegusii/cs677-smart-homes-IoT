//The main interfaces are in structs... this file is just for testing

package main

import (
	"fmt"
		"github.com/ppegusii/cs677-smart-homes-IoT/structs"
		"github.com/ppegusii/cs677-smart-homes-IoT/api"
		"time"
)

func main(){
	var cachemap *structs.Cache = structs.NewCache(10)
	fmt.Println("Cachemap as below\n", cachemap)
	var stateInfo *api.StateInfo

	for i := 0; i <10; i++{
	stateInfo = &api.StateInfo{
		Clock:      int(time.Now().Unix()), //current timestamp for event ordering
		DeviceId:   10,
		DeviceName: api.Motion,
		State:      api.MotionStart,
		}
		time.Sleep(1000 * time.Millisecond)
		cachemap.Set(i, stateInfo)		
	}


	fmt.Println("Cachemap is as below\n", cachemap)
	fmt.Println("Cachemap is as below\n", cachemap.Gettimestamp(6))
	cachemap.Get(6)
	fmt.Println("Cachemap is as below\n", cachemap.Gettimestamp(6))
	fmt.Println("Cachemap is as below\n", cachemap.Gettimestamp(0))
	cachemap.Get(0)
	fmt.Println("Cachemap is as below\n", cachemap.Gettimestamp(0))
	cachemap.Delete(1)
	evict := cachemap.OldCache()
	fmt.Println("evict %s entry",evict)

	stateInfo = &api.StateInfo{
		Clock:      int(time.Now().Unix()), //current timestamp for event ordering
		DeviceId:   10,
		DeviceName: api.Motion,
		State:      api.MotionStart,
		}

	cachemap.AddEntry(stateInfo)
}