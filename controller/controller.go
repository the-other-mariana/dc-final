package controller

import (
	"fmt"
	//"log"
	"os"
	"time"
	"strings"
	"strconv"

	"go.nanomsg.org/mangos"
	//"go.nanomsg.org/mangos/protocol/pub"
	//"dc-labs/mangos/protocol/surveyor"
	"go.nanomsg.org/mangos/protocol/surveyor"

	// register transports
	_ "go.nanomsg.org/mangos/transport/all"
)

var controllerAddress = "tcp://localhost:40899"
var sock mangos.Socket
var done = make(chan string)

var actions = make(map[string]Action)
var Workers = make(map[string]Worker)
var Workloads = make(map[string]Workload)
//var filters = make(map[string]ImageService)

type Worker struct {
	Name     string `json:"name"`
	Tags     string `json:"tags"`
	Status   string `json:"status"`
	Usage    int    `json:"usage"`
	URL      string `json:"url"`
	Active   bool   `json:"active"`
	Port     int    `json:"port"`
	JobsDone int    `json:"jobsDone"` 
}

type Action struct {
	id		int
	worker 	string
}
type ImageService struct{
	id int
	image string
	worker string
}
type Workload struct{
	Id string
	Filter string
	Name string
	Status string
	Jobs int
	Imgs []string
}

func die(format string, v ...interface{}) {
	fmt.Fprintln(os.Stderr, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func date() string {
	return time.Now().Format(time.ANSIC)
}

func Start() {
	var sock mangos.Socket
	var err error
	if sock, err = surveyor.NewSocket(); err != nil {
		die("can't get new surveyor socket: %s", err)
	}
	if err = sock.Listen(controllerAddress); err != nil {
		die("can't listen on surveyor socket: %s", err.Error())
	}
	err = sock.SetOption(mangos.OptionSurveyTime, time.Second)
	if err != nil {
		die("At set option: %s", err.Error())
	}
	
	var resp []byte
	for {
		// Could also use sock.RecvMsg to get header
		err = sock.Send([]byte("Welcome workers"))
		if err != nil {
			die("No workers %+v", err.Error())
		}
		for {
			if resp, err = sock.Recv(); err != nil {
				break
			}
			exists := false
			worker := GetWorkerInfo(string(resp))
			for _, w := range Workers {
				if w.Name == worker.Name {
					exists = true
				}
			}
			if !exists {
				Workers[worker.Name] = worker
			}
			PrintWorker(worker)
		}
	}
}

func PrintWorker(worker Worker) {
	fmt.Println(Workers[worker.Name].Name, " serves in localhost:", Workers[worker.Name].Port, "\n")
}

func GetWorkerInfo(resp string) (Worker) {
	worker := Worker{}
	
	msg := strings.Split(resp, " ")

	worker.Name = msg[0]
	worker.Status = "free"
	usage, _ := strconv.Atoi(msg[2])
	worker.Usage = usage
	worker.Tags = msg[3]
	port, _ := strconv.Atoi(msg[4])
	worker.Port = port
	jobsDone, _ := strconv.Atoi(msg[5])
	worker.JobsDone = jobsDone
	worker.Active = true
	worker.URL = "localhost:" + msg[4]

	return worker
}

func Register(name string, num int) {
	actions[strconv.Itoa(num)] = Action{id: num, worker: name}
}

func UpdateStatus(name string) {
	if w, ok := Workers[name]; ok {
		if w.Status == "free" {
			w.Status = "busy"
		} else {
			w.Status = "free"
		}
	}
}

func UpdateUsage(name string) {
	if w, ok := Workers[name]; ok {
		w.Usage += 1
		w.JobsDone += 1
		Workers[name] = w
	}
}

func GetWorker(id int) string {
	name := actions[strconv.Itoa(id)].worker
	return name
}

