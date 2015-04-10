package main

import (
	"encoding/json"
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"io/ioutil"
	"log"
)

type Config struct {
	DatabaseDirectory       string
	GatewayTempPollInterval int
	IPAddress               string
	Ordering                string
	Port                    string
	ProcessDescriptions     map[string]ProcessDescription
	StartLocalProcesses     []string
}

type ProcessDescription struct {
	IPAddress string
	Port      string
}

type Test struct {
	Instructions []Instruction
}

type Instruction struct {
	Command string
	State   api.State
	Target  string
	Time    int
}

func parse(fileName *string, structPtr interface{}) {
	var file []byte
	var err error
	file, err = ioutil.ReadFile(*fileName)
	if err != nil {
		log.Printf("File error: %+v\n", err)
		return
	}
	err = json.Unmarshal(file, &structPtr)
	if err != nil {
		log.Printf("JSON error: %+v\n", err)
		return
	}
	log.Printf("Struct: %+v\n", structPtr)
}
