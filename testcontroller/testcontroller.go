package main

import (
	"bufio"
	"fmt"
	"github.com/oleiade/lane"
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"github.com/ppegusii/cs677-smart-homes-IoT/util"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"
)

const (
	ipSwitch          string = "-i"
	portSwitch        string = "-p"
	otherIpSwitch     string = "-I"
	otherPortSwitch   string = "-P"
	anotherIpSwitch   string = "-I2"
	anotherPortSwitch string = "-P2"
	orderingSwitch    string = "-o"
	pollingSwitch     string = "-s"
	repIpSwitch       string = "-ri"
	repPortSwitch     string = "-rp"
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
	var database1Ip string = this.config.ProcessDescriptions["database1"].IPAddress
	var database1Port string = this.config.ProcessDescriptions["database1"].Port
	var gateway1Ip string = this.config.ProcessDescriptions["gateway1"].IPAddress
	var gateway1Port string = this.config.ProcessDescriptions["gateway1"].Port
	var database2Ip string = this.config.ProcessDescriptions["database2"].IPAddress
	var database2Port string = this.config.ProcessDescriptions["database2"].Port
	var gateway2Ip string = this.config.ProcessDescriptions["gateway2"].IPAddress
	var gateway2Port string = this.config.ProcessDescriptions["gateway2"].Port
	var command string
	var args []string
	//if databases are local processes start them first
	if stringInSlice(this.config.StartLocalProcesses, "database1") {
		command = "database"
		args = []string{ipSwitch, database1Ip, portSwitch,
			database1Port, orderingSwitch, this.config.Ordering}
		this.startProcess("database1", command, args)
		//allow database to come up, gateway depends on it
		waitFor(time.Second)
	}
	if stringInSlice(this.config.StartLocalProcesses, "database2") {
		command = "database"
		args = []string{ipSwitch, database2Ip, portSwitch,
			database2Port, orderingSwitch, this.config.Ordering}
		this.startProcess("database2", command, args)
		//allow database to come up, gateway depends on it
		waitFor(time.Second)
	}
	//next start gateways if they're local processes
	if stringInSlice(this.config.StartLocalProcesses, "gateway1") {
		command = "gateway"
		args = []string{ipSwitch, gateway1Ip, portSwitch, gateway1Port,
			pollingSwitch, strconv.Itoa(this.config.GatewayTempPollInterval),
			orderingSwitch, this.config.Ordering, repIpSwitch,
			gateway2Ip, repPortSwitch, gateway2Port}
		this.startProcess("gateway1", command, args)
		//allow gateway to come up, all other processes depend on it
		waitFor(time.Second)
	}
	if stringInSlice(this.config.StartLocalProcesses, "gateway2") {
		command = "gateway"
		args = []string{ipSwitch, gateway2Ip, portSwitch, gateway2Port,
			pollingSwitch, strconv.Itoa(this.config.GatewayTempPollInterval),
			orderingSwitch, this.config.Ordering, repIpSwitch,
			gateway1Ip, repPortSwitch, gateway1Port}
		this.startProcess("gateway2", command, args)
		//allow gateway to come up, all other processes depend on it
		waitFor(time.Second)
	}
	//start all other processes
	for _, process := range this.config.StartLocalProcesses {
		if process == "gateway1" || process == "database1" ||
			process == "gateway2" || process == "database2" {
			continue
		}
		procDesc = this.config.ProcessDescriptions[process]
		args = []string{ipSwitch, procDesc.IPAddress, portSwitch,
			procDesc.Port, otherIpSwitch, gateway1Ip, otherPortSwitch,
			gateway1Port, anotherIpSwitch, gateway2Ip, anotherPortSwitch,
			gateway2Port, orderingSwitch, this.config.Ordering}
		this.startProcess(process, process, args)
	}
	//make sure all processes are up and registered
	//extra time for replica election
	waitFor(6 * time.Second)
}

//Starts a local process.
func (this *TestController) startProcess(name, command string, args []string) {
	log.Printf("Starting process: %s %+v\n", command, args)
	var cmd *exec.Cmd = exec.Command(command, args...)
	err := cmd.Start()
	if err != nil {
		log.Printf("Error starting process \"%s %+v\": %+v\n", command, args, err)
		this.killAllLocalProcesses()
		os.Exit(1)
	}
	this.processes[name] = cmd
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
	var empty struct{}
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
		go util.RpcSync(process.IPAddress, process.Port,
			"Gateway.Query", name, &empty, false)
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
		go util.RpcSync(process.IPAddress, process.Port, rpcName,
			api.StateInfo{State: inst.State}, &empty, false)
		return
	case "KillProcess":
		if !stringInSlice(this.config.StartLocalProcesses, inst.Target) {
			log.Printf("Will not kill process not started by this testcontroller: %s\n", inst.Target)
			return
		}
		if !(inst.Target == "gateway1" || inst.Target == "database1" ||
			inst.Target == "gateway2" || inst.Target == "database2") {
			log.Printf("Will not kill non-replica: %s\n", inst.Target)
			return
		}
		this.killLocalProcess(inst.Target)
	case "StartProcess":
		var procDesc ProcessDescription
		var database1Ip string = this.config.ProcessDescriptions["database1"].IPAddress
		var database1Port string = this.config.ProcessDescriptions["database1"].Port
		var gateway1Ip string = this.config.ProcessDescriptions["gateway1"].IPAddress
		var gateway1Port string = this.config.ProcessDescriptions["gateway1"].Port
		var database2Ip string = this.config.ProcessDescriptions["database2"].IPAddress
		var database2Port string = this.config.ProcessDescriptions["database2"].Port
		var gateway2Ip string = this.config.ProcessDescriptions["gateway2"].IPAddress
		var gateway2Port string = this.config.ProcessDescriptions["gateway2"].Port
		var command string
		var args []string
		switch inst.Target {
		case "database1":
			command = "database"
			args = []string{ipSwitch, database1Ip, portSwitch,
				database1Port, orderingSwitch, this.config.Ordering}
			this.startProcess("database1", command, args)
		case "gateway1":
			command = "gateway"
			args = []string{ipSwitch, gateway1Ip, portSwitch, gateway1Port,
				pollingSwitch, strconv.Itoa(this.config.GatewayTempPollInterval),
				orderingSwitch, this.config.Ordering, repIpSwitch,
				gateway2Ip, repPortSwitch, gateway2Port}
			this.startProcess("gateway1", command, args)
		case "database2":
			command = "database"
			args = []string{ipSwitch, database2Ip, portSwitch,
				database2Port, orderingSwitch, this.config.Ordering}
			this.startProcess("database2", command, args)
		case "gateway2":
			command = "gateway"
			args = []string{ipSwitch, gateway2Ip, portSwitch, gateway2Port,
				pollingSwitch, strconv.Itoa(this.config.GatewayTempPollInterval),
				orderingSwitch, this.config.Ordering, repIpSwitch,
				gateway1Ip, repPortSwitch, gateway1Port}
			this.startProcess("gateway2", command, args)
		default:
			procDesc, ok = this.config.ProcessDescriptions[inst.Target]
			if !ok {
				log.Printf("No ProcessDescription cannot start: %s\n", inst.Target)
				return
			}
			args = []string{ipSwitch, procDesc.IPAddress, portSwitch,
				procDesc.Port, otherIpSwitch, gateway1Ip, otherPortSwitch,
				gateway1Port, anotherIpSwitch, gateway2Ip, anotherPortSwitch,
				gateway2Port, orderingSwitch, this.config.Ordering}
			this.startProcess(inst.Target, inst.Target, args)
		}
	default:
		log.Printf("Invalid instruction command: %s\n", inst.Command)
		return
	}
}

func waitFor(duration time.Duration) {
	timer := time.NewTimer(duration)
	<-timer.C
}

func (this *TestController) requestEnd() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Press enter to exit")
	reader.ReadString('\n')
}

func (this *TestController) killAllLocalProcesses() {
	var processNames []string = make([]string, len(this.processes))
	var idx int = 0
	for processName, _ := range this.processes {
		processNames[idx] = processName
		idx++
	}
	for _, processName := range processNames {
		this.killLocalProcess(processName)
	}
}

//http://stackoverflow.com/questions/11886531/terminating-a-process-started-with-os-exec-in-golang
func (this *TestController) killLocalProcess(process string) {
	cmd, ok := this.processes[process]
	if !ok {
		log.Printf("killLocalProcess given invalid name: %s\n", process)
		return
	}
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	select {
	case <-time.After(time.Millisecond):
		if err := cmd.Process.Kill(); err != nil {
			log.Printf("failed to kill %s: ", process, err)
			return
		}
		<-done // allow goroutine to exit
		log.Printf("process %s killed", process)
	case err := <-done:
		if err != nil {
			log.Printf("process %s done with error = %v", process, err)
		}
	}
	delete(this.processes, process)
}

func stringInSlice(slice []string, s string) bool {
	for _, val := range slice {
		if s == val {
			return true
		}
	}
	return false
}
