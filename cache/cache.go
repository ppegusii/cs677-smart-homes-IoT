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
		Clock:      int64(time.Now().Unix()), //current timestamp for event ordering
		DeviceId:   i,
		DeviceName: api.Motion,
		State:      api.MotionStart,
		}
	time.Sleep(500 * time.Millisecond)
	cachemap.AddEntry(stateInfo)
	}


	fmt.Println("Cachemap is as below\n", cachemap)
	fmt.Println("Cachemap is as below\n", cachemap.Gettimestamp(6))
	cachemap.Get(6)
	fmt.Println("Cachemap is as below\n", cachemap.Gettimestamp(6))
	fmt.Println("Cachemap is as below\n", cachemap.Gettimestamp(0))
	cachemap.Get(0)
	fmt.Println("Cachemap is as below\n", cachemap.Gettimestamp(0))
	cachemap.Delete(1)
	fmt.Println("Cachemap is as below\n", cachemap)
//	evict := cachemap.OldCache()
//	fmt.Println("evict %s entry",evict)

	stateInfo = &api.StateInfo{
		Clock:      int64(time.Now().Unix()), //current timestamp for event ordering
		DeviceId:   10,
		DeviceName: api.Motion,
		State:      api.MotionStart,
		}

	cachemap.AddEntry(stateInfo)

	cachemap.Get(3)
	time.Sleep(1000 * time.Millisecond)
	cachemap.Get(4)
	time.Sleep(1000 * time.Millisecond)
	cachemap.Get(5)
	time.Sleep(1000 * time.Millisecond)
	fmt.Println("Cachemap is as below\n", cachemap)

	stateInfo = &api.StateInfo{
		Clock:      int64(time.Now().Unix()), //current timestamp for event ordering
		DeviceId:   4,
		DeviceName: api.Temperature,
		State:      api.On,
		}

	cachemap.AddEntry(stateInfo) //Should at at 2

	stateInfo = &api.StateInfo{
		Clock:      int64(time.Now().Unix()), //current timestamp for event ordering
		DeviceId:   100,
		DeviceName: api.Temperature,
		State:      api.On,
		}

	cachemap.AddEntry(stateInfo) //Should at at 7 or 8 depends if cachemap[0] and [1] have same values or diff

	cachemap.LookupDeviceID(100)

	stateInfo = &api.StateInfo{
		Clock:      int64(time.Now().Unix()), //current timestamp for event ordering
		DeviceId:   100,
		DeviceName: api.Temperature,
		State:      api.On,
		}

	cachemap.AddEntry(stateInfo) 

	d := cachemap.GetEntry(100)
	fmt.Println("StateInfo for device 100 is as below \n",d)

}