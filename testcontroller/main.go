package main

import (
	"flag"
	"log"
)

func main() {
	var configFileName *string = flag.String("c", "config.json", "config file")
	var testFileName *string = flag.String("t", "test.json", "test file")
	flag.Parse()
	log.Printf("Starting TestController\n")
	tc := NewTestController(configFileName, testFileName)
	tc.startProcesses()
	tc.runTestCase()
	tc.killLocalProcesses()
}
