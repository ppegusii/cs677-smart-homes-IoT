package main

import (
	"github.com/oleiade/lane"
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"log"
	"net/rpc"
	"os"
	"os/exec"
	"strconv"
	"time"
)

type TestController struct {
	config    *Config
	processes map[string]*exec.Cmd
	test      *Test
}

func NewTestController(configFileName *string, testFileName *string) *TestController {
	var config Config
	var test Test
	parse(configFileName, &config)
	parse(testFileName, &test)
	return &TestController{
		config:    &config,
		test:      &test,
		processes: make(map[string]*exec.Cmd),
	}
}

//Starts all the local processes from the config file.
func (this *TestController) startProcesses() {
	var procDesc ProcessDescription
	var ipSwitch string = "-i"
	var portSwitch string = "-p"
	var otherIpSwitch string = "-I"
	var otherPortSwitch string = "-P"
	var orderingSwitch string = "-o"
	var pollingSwitch string = "-s"
	var databaseIp string = this.config.ProcessDescriptions["database"].IPAddress
	var databasePort string = this.config.ProcessDescriptions["database"].Port
	var gatewayIp string = this.config.ProcessDescriptions["gateway"].IPAddress
	var gatewayPort string = this.config.ProcessDescriptions["gateway"].Port
	var command string
	var args []string
	//if database is a local process start it first
	if stringInSlice(this.config.StartLocalProcesses, "database") {
		command = "database"
		args = []string{ipSwitch, databaseIp, portSwitch, databasePort, orderingSwitch, this.config.Ordering}
		this.startProcess(command, args)
		//allow database to come up, gateway depends on it
		waitFor(time.Second)
	}
	//next start gateway if it's a local process
	if stringInSlice(this.config.StartLocalProcesses, "gateway") {
		command = "gateway"
		args = []string{ipSwitch, gatewayIp, portSwitch, gatewayPort, pollingSwitch, strconv.Itoa(this.config.GatewayTempPollInterval), orderingSwitch, this.config.Ordering}
		this.startProcess(command, args)
		//allow gateway to come up, all other processes depend on it
		waitFor(time.Second)
	}
	//start all other processes
	for _, process := range this.config.StartLocalProcesses {
		if process == "gateway" || process == "database" {
			continue
		}
		procDesc = this.config.ProcessDescriptions[process]
		args = []string{ipSwitch, procDesc.IPAddress, portSwitch, procDesc.Port, otherIpSwitch, gatewayIp, otherPortSwitch, gatewayPort, orderingSwitch, this.config.Ordering}
		this.startProcess(process, args)
	}
	//make sure all processes are up and registered
	waitFor(time.Second)
}

//Starts a local process.
func (this *TestController) startProcess(command string, args []string) {
	var cmd *exec.Cmd = exec.Command(command, args...)
	err := cmd.Start()
	if err != nil {
		log.Printf("Error starting process \"%s %+v\": %+v\n", command, args, err)
		this.killLocalProcesses()
		os.Exit(1)
	}
	this.processes[command] = cmd
}

//Runs all the inststructions in the test file.
func (this *TestController) runTestCase() {
	//waitFor(time.Second * 10)
	//schedule the instructions
	var q *lane.PQueue = lane.NewPQueue(lane.MINPQ)
	for _, instruction := range this.test.Instructions {
		q.Push(instruction, instruction.Time)
	}
	//run the instructions
	var startTime time.Time = time.Now()
	for q.Size() > 0 {
		next, ms := q.Pop()
		//calculate time to run instruction
		var curTime time.Time = time.Now()
		var nextTime time.Time = startTime.Add(time.Millisecond * time.Duration(ms))
		waitFor(nextTime.Sub(curTime))
		//run instruction
		this.runInstruction(next.(Instruction))
	}
}

//Run an inststruction.
func (this *TestController) runInstruction(inst Instruction) {
	log.Printf("Will run instruction: %+v\n", inst)
	var client *rpc.Client
	var empty struct{}
	var err error
	var ok bool
	var process ProcessDescription
	switch inst.Command {
	//all QueryState commands will be made to RPCs to gateway
	case "QueryState":
		var name api.Name
		switch inst.Target {
		case "motionsensor":
			name = api.Motion
			break
		case "temperaturesensor":
			name = api.Temperature
			break
		default:
			log.Printf("Query state unimplemented for: %s\n", inst.Target)
			return
		}
		process, ok = this.config.ProcessDescriptions["gateway"]
		if !ok {
			log.Printf("Gateway process does not exist: %s\n", inst.Target)
		}
		client, err = rpc.Dial("tcp", process.IPAddress+":"+process.Port)
		if err != nil {
			log.Printf("Error dialing gateway: %+v\n", err)
		}
		client.Go("Gateway.Query", name, &empty, nil)
		return
	case "ChangeState":
		//all ChangeState commands will be made to RPCs to sensors
		var rpcName string
		switch inst.Target {
		case "motionsensor":
			rpcName = "MotionSensor.ChangeState"
			break
		case "temperaturesensor":
			rpcName = "TemperatureSensor.ChangeState"
			break
		case "doorsensor":
			rpcName = "DoorSensor.ChangeState"
			break
		default:
			log.Printf("Invalid sensor for change state: %s\n", inst.Target)
			return
		}
		process, ok = this.config.ProcessDescriptions[inst.Target]
		if !ok {
			log.Printf("Process does not exist: %s\n", inst.Target)
		}
		client, err = rpc.Dial("tcp", process.IPAddress+":"+process.Port)
		if err != nil {
			log.Printf("Error dialing %s: %+v\n", inst.Target, err)
		}
		client.Go(rpcName, api.StateInfo{State: inst.State}, &empty, nil)
		return
	default:
		log.Printf("Invalid instruction command: %s\n", inst.Command)
		return
	}
}

func waitFor(duration time.Duration) {
	timer := time.NewTimer(duration)
	<-timer.C
}

//http://stackoverflow.com/questions/11886531/terminating-a-process-started-with-os-exec-in-golang
func (this *TestController) killLocalProcesses() {
	for process, cmd := range this.processes {
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()
		select {
		case <-time.After(time.Millisecond):
			if err := cmd.Process.Kill(); err != nil {
				log.Fatal("failed to kill %s: ", process, err)
			}
			<-done // allow goroutine to exit
			log.Printf("process %s killed", process)
		case err := <-done:
			if err != nil {
				log.Printf("process %s done with error = %v", process, err)
			}
		}
	}
}

func stringInSlice(slice []string, s string) bool {
	for _, val := range slice {
		if s == val {
			return true
		}
	}
	return false
}
