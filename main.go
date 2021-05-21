package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/CodersSquad/dc-final/api"
	"github.com/CodersSquad/dc-final/controller"
	"github.com/CodersSquad/dc-final/scheduler"
)

func main() {
	log.Println("Welcome to the Distributed and Parallel Image Processing System")

	// Start Controller
	go controller.Start()

	// Start Scheduler
	jobs := make(chan scheduler.Job)
	go scheduler.Start(jobs)
	// Send sample jobs
	sampleJob := scheduler.Job{Address: "localhost:50051", RPCName: "hello"}

	// API
	go api.Start()

	for {
		sampleJob.RPCName = fmt.Sprintf("hello-%v", rand.Intn(10000))
		jobs <- sampleJob
		time.Sleep(time.Second * 5)
	}
}
